package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxAvatarSize = 5 << 20

func avatarURL(name string) string {
	if name == "" {
		return ""
	}
	return "/api/avatars/" + name
}

func (s *Service) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	s.uploadFile(w, r, true)
}

func (s *Service) uploadFile(w http.ResponseWriter, r *http.Request, avatar bool) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondError(w, 405, "method not allowed")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize+(1<<16))
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		respondError(w, 400, "invalid or oversized upload")
		return
	}
	defer r.MultipartForm.RemoveAll()
	field := "file"
	if avatar && r.URL.Path == "/auth/avatar" {
		field = "avatar"
	}
	file, _, err := r.FormFile(field)
	if err != nil {
		respondError(w, 400, "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil || len(data) > maxAvatarSize {
		respondError(w, 400, "invalid or oversized upload")
		return
	}
	format := ""
	if avatar {
		config, imageFormat, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil || (imageFormat != "jpeg" && imageFormat != "png" && imageFormat != "gif") {
			respondError(w, 400, "unsupported image")
			return
		}
		if config.Width > 10000 || config.Height > 10000 || int64(config.Width)*int64(config.Height) > 40000000 {
			respondError(w, 400, "image dimensions too large")
			return
		}
		if _, _, err = image.Decode(bytes.NewReader(data)); err != nil {
			respondError(w, 400, "invalid image")
			return
		}
		format = "." + imageFormat
	}
	dir := "uploads/files"
	if avatar {
		dir = "uploads/avatars"
	}
	if err = os.MkdirAll(dir, 0755); err != nil {
		respondError(w, 500, "internal server error")
		return
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		respondError(w, 500, "internal server error")
		return
	}
	name := hex.EncodeToString(random) + format
	path := filepath.Join(dir, name)
	if err = os.WriteFile(path, data, 0644); err != nil {
		respondError(w, 500, "internal server error")
		return
	}
	if !avatar {
		respond(w, 201, map[string]string{"url": "/api/files/" + name})
		return
	}
	id, _ := UserIDFromContext(r.Context())
	result := s.db.WithContext(r.Context()).Model(&User{}).Where("id = ?", id).Update("avatar", name)
	if result.Error != nil || result.RowsAffected == 0 {
		os.Remove(path)
		if result.Error != nil {
			respondError(w, 500, "internal server error")
		} else {
			respondError(w, 401, "unauthorized")
		}
		return
	}
	respond(w, 200, map[string]string{"avatar_url": avatarURL(name)})
}

func (s *Service) serveAvatar(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/avatars/")
	parts := strings.Split(name, ".")
	if len(parts) != 2 || len(parts[0]) != 32 || (parts[1] != "jpeg" && parts[1] != "png" && parts[1] != "gif") {
		http.NotFound(w, r)
		return
	}
	if _, err := hex.DecodeString(parts[0]); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filepath.Join("uploads/avatars", name))
}

func (s *Service) files(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		respondError(w, 405, "method not allowed")
		return
	}
	purpose := r.URL.Query().Get("purpose")
	if purpose != "" && purpose != "avatar" {
		respondError(w, 400, "invalid purpose")
		return
	}
	s.uploadFile(w, r, purpose == "avatar")
}

func (s *Service) serveFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		respondError(w, 405, "method not allowed")
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/files/")
	if len(name) != 32 {
		http.NotFound(w, r)
		return
	}
	if _, err := hex.DecodeString(name); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filepath.Join("uploads/files", name))
}
