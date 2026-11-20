package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"uptime-up/backend/internal/database"
)

func TestAccessTokens(t *testing.T) {
	s := New(nil, "12345678901234567890123456789012", time.Minute, time.Hour)
	sign := func(claims jwt.RegisteredClaims, method jwt.SigningMethod, secret []byte) string {
		token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	secret := s.secret
	valid := jwt.RegisteredClaims{Subject: "123", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}
	tests := []struct {
		name, token string
		ok          bool
	}{
		{"valid", sign(valid, jwt.SigningMethodHS256, secret), true},
		{"expired", sign(jwt.RegisteredClaims{Subject: "123", ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}, jwt.SigningMethodHS256, secret), false},
		{"missing expiry", sign(jwt.RegisteredClaims{Subject: "123"}, jwt.SigningMethodHS256, secret), false},
		{"wrong key", sign(valid, jwt.SigningMethodHS256, []byte("wrong")), false},
		{"wrong algorithm", sign(valid, jwt.SigningMethodHS384, secret), false},
		{"invalid subject", sign(jwt.RegisteredClaims{Subject: "0", ExpiresAt: valid.ExpiresAt}, jwt.SigningMethodHS256, secret), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := s.UserID(tt.token)
			if (err == nil) != tt.ok || (tt.ok && id != 123) {
				t.Fatalf("id=%d err=%v", id, err)
			}
		})
	}
}
func TestValidationAndMiddleware(t *testing.T) {
	s := New(nil, "12345678901234567890123456789012", time.Minute, time.Hour)
	for _, body := range []string{`{}`, `{"login":"a","password":"password"}`, `{"login":"user","password":"short"}`, `{"login":"user","password":"password","extra":1}`, `{} {}`} {
		r := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != 400 {
			t.Fatalf("body %s: %d", body, w.Code)
		}
	}
	profileToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "42", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}).SignedString(s.secret)
	for _, tc := range []struct {
		token, body string
		status      int
	}{
		{"", `{"display_name":"Alice"}`, 401},
		{profileToken, `{"display_name":"  "}`, 400},
		{profileToken, `{"display_name":"Alice","extra":1}`, 400},
	} {
		r := httptest.NewRequest(http.MethodPatch, "/auth/me", bytes.NewBufferString(tc.body))
		if tc.token != "" {
			r.Header.Set("Authorization", "Bearer "+tc.token)
		}
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Errorf("profile update: got %d, want %d", w.Code, tc.status)
		}
	}
	h := s.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := UserIDFromContext(r.Context())
		if !ok || id != 42 {
			t.Error("missing user ID")
		}
		w.WriteHeader(204)
	}))
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "42", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}).SignedString(s.secret)
	for _, tc := range []struct {
		header string
		status int
	}{{"", 401}, {"Bearer invalid", 401}, {"Bearer " + token, 204}} {
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("Authorization", tc.header)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Errorf("header %q: %d", tc.header, w.Code)
		}
	}
}
func TestPostgresAuthFlow(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated test database")
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	s := New(db, "12345678901234567890123456789012", time.Minute, time.Hour)
	h := s.Handler()
	login := fmt.Sprintf("test_%d", time.Now().UnixNano())
	t.Cleanup(func() { db.Exec("DELETE FROM users WHERE login = ?", login) })
	call := func(path string, body interface{}) *httptest.ResponseRecorder {
		data, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("POST", path, bytes.NewReader(data)))
		return w
	}
	credentials := map[string]string{"login": login, "password": "password123"}
	w := call("/auth/register", credentials)
	if w.Code != 201 {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var tokens Tokens
	if err := json.Unmarshal(w.Body.Bytes(), &tokens); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UserID(tokens.AccessToken); err != nil {
		t.Fatal(err)
	}
	profileRequest := func(method, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/auth/me", bytes.NewBufferString(body))
		r.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w = profileRequest("PATCH", `{"display_name":"Новое имя"}`); w.Code != 200 {
		t.Fatalf("update profile: %d %s", w.Code, w.Body.String())
	}
	if w = profileRequest("GET", ""); w.Code != 200 || !bytes.Contains(w.Body.Bytes(), []byte("Новое имя")) {
		t.Fatalf("read profile: %d %s", w.Code, w.Body.String())
	}
	if w = profileRequest("PATCH", `{"display_name":"  "}`); w.Code != 400 {
		t.Fatalf("invalid profile name: %d", w.Code)
	}
	if w = call("/auth/register", credentials); w.Code != 409 {
		t.Fatalf("duplicate: %d", w.Code)
	}
	if w = call("/auth/login", map[string]string{"login": login, "password": "incorrect"}); w.Code != 401 {
		t.Fatalf("invalid login: %d", w.Code)
	}
	if w = call("/auth/login", credentials); w.Code != 200 {
		t.Fatalf("login: %d", w.Code)
	}
	var wg sync.WaitGroup
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- call("/auth/refresh", map[string]string{"refresh_token": tokens.RefreshToken}).Code
		}()
	}
	wg.Wait()
	close(results)
	counts := map[int]int{}
	for code := range results {
		counts[code]++
	}
	if counts[200] != 1 || counts[401] != 1 {
		t.Fatalf("concurrent refresh: %v", counts)
	}
	if w = call("/auth/refresh", map[string]string{"refresh_token": "invalid"}); w.Code != 401 {
		t.Fatalf("invalid refresh: %d", w.Code)
	}
	expired, err := s.Login(context.Background(), login, "password123")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(expired.RefreshToken))
	hash := sum[:]
	if err := db.Model(&RefreshSession{}).Where("token_hash = ?", hash).Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if w = call("/auth/refresh", map[string]string{"refresh_token": expired.RefreshToken}); w.Code != 401 {
		t.Fatalf("expired refresh: %d", w.Code)
	}
}
