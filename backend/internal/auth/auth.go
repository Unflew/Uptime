package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           int64 `gorm:"primaryKey"`
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
type RefreshSession struct {
	ID        int64 `gorm:"primaryKey"`
	UserID    int64
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
type Service struct {
	db                    *gorm.DB
	secret                []byte
	accessTTL, refreshTTL time.Duration
}

func New(db *gorm.DB, secret string, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{db: db, secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

var errCredentials = errors.New("invalid credentials")
var errRefresh = errors.New("invalid refresh token")
var errLoginTaken = errors.New("login already exists")

func (s *Service) issue(tx *gorm.DB, userID int64) (Tokens, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{Subject: strconv.FormatInt(userID, 10), IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL))}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return Tokens{}, err
	}
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return Tokens{}, err
	}
	refresh := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(refresh))
	if err = tx.Create(&RefreshSession{UserID: userID, TokenHash: hash[:], ExpiresAt: now.Add(s.refreshTTL)}).Error; err != nil {
		return Tokens{}, err
	}
	return Tokens{AccessToken: access, RefreshToken: refresh, TokenType: "Bearer", ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}
func validCredentials(login, password string) bool {
	return utf8.ValidString(login) && len(login) >= 3 && len(login) <= 64 && strings.TrimSpace(login) == login && !strings.ContainsAny(login, " \t\r\n") && len(password) >= 8 && len(password) <= 72
}
func (s *Service) Register(ctx context.Context, login, password string) (Tokens, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Tokens{}, err
	}
	var tokens Tokens
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		user := User{Login: login, PasswordHash: string(hash)}
		result := tx.Raw("INSERT INTO users (login, password_hash) VALUES (?, ?) ON CONFLICT (login) DO NOTHING RETURNING id", login, string(hash)).Scan(&user.ID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 || user.ID == 0 {
			return errLoginTaken
		}
		var err error
		tokens, err = s.issue(tx, user.ID)
		return err
	})
	return tokens, err
}
func (s *Service) Login(ctx context.Context, login, password string) (Tokens, error) {
	var user User
	err := s.db.WithContext(ctx).Where("login = ?", login).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Tokens{}, errCredentials
	}
	if err != nil {
		return Tokens{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return Tokens{}, errCredentials
	}
	var tokens Tokens
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { var e error; tokens, e = s.issue(tx, user.ID); return e })
	return tokens, err
}
func (s *Service) Refresh(ctx context.Context, token string) (Tokens, error) {
	if len(token) != 43 {
		return Tokens{}, errRefresh
	}
	if _, err := base64.RawURLEncoding.DecodeString(token); err != nil {
		return Tokens{}, errRefresh
	}
	hash := sha256.Sum256([]byte(token))
	var tokens Tokens
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var session RefreshSession
		result := tx.Raw("DELETE FROM refresh_sessions WHERE token_hash = ? AND expires_at > NOW() RETURNING user_id", hash[:]).Scan(&session)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 || session.UserID == 0 {
			return errRefresh
		}
		var err error
		tokens, err = s.issue(tx, session.UserID)
		return err
	})
	return tokens, err
}
func (s *Service) UserID(token string) (int64, error) {
	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) { return s.secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil || !parsed.Valid {
		return 0, errors.New("invalid access token")
	}
	id, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid access token")
	}
	return id, nil
}

type contextKey struct{}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(contextKey{}).(int64)
	return id, ok
}
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		id, err := s.UserID(parts[1])
		if err != nil {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, id)))
	})
}
func respond(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request, dest interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dest); err != nil {
		return false
	}
	return dec.Decode(&struct{}{}) == io.EOF
}
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	credentials := func(register bool) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.Header().Set("Allow", http.MethodPost)
				respondError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			var req struct {
				Login    string `json:"login"`
				Password string `json:"password"`
			}
			if !decode(w, r, &req) || !validCredentials(req.Login, req.Password) {
				respondError(w, http.StatusBadRequest, "invalid login or password format")
				return
			}
			var tokens Tokens
			var err error
			status := http.StatusOK
			if register {
				tokens, err = s.Register(r.Context(), req.Login, req.Password)
				status = http.StatusCreated
			} else {
				tokens, err = s.Login(r.Context(), req.Login, req.Password)
			}
			switch {
			case errors.Is(err, errLoginTaken):
				respondError(w, http.StatusConflict, "login already exists")
			case errors.Is(err, errCredentials):
				respondError(w, http.StatusUnauthorized, "invalid credentials")
			case err != nil:
				respondError(w, http.StatusInternalServerError, "internal server error")
			default:
				respond(w, status, tokens)
			}
		}
	}
	mux.HandleFunc("/auth/register", credentials(true))
	mux.HandleFunc("/auth/login", credentials(false))
	mux.HandleFunc("/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if !decode(w, r, &req) || req.RefreshToken == "" {
			respondError(w, http.StatusBadRequest, "refresh_token is required")
			return
		}
		tokens, err := s.Refresh(r.Context(), req.RefreshToken)
		if errors.Is(err, errRefresh) {
			respondError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond(w, http.StatusOK, tokens)
	})
	mux.Handle("/auth/me", s.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			respondError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id, _ := UserIDFromContext(r.Context())
		var user User
		err := s.db.WithContext(r.Context()).Select("id", "login", "created_at").First(&user, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		respond(w, http.StatusOK, map[string]interface{}{"login": user.Login, "created_at": user.CreatedAt})
	})))
	return mux
}
