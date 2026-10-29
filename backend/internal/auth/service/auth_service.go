package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"sync"
	"time"

	"lesson06/backend/internal/auth/dto"
	"lesson06/backend/internal/pkg/apperr"
)

// shortcut: in-memory sessions suit a single local teaching server; use shared persistent sessions before multi-instance deployment.
type AuthService struct {
	password string
	mu       sync.Mutex
	sessions map[string]time.Time
}

func NewAuthService(password string) *AuthService {
	return &AuthService{password: password, sessions: map[string]time.Time{}}
}
func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResult, error) {
	if req.Username != "admin" || subtle.ConstantTimeCompare([]byte(req.Password), []byte(s.password)) != 1 {
		return nil, apperr.New(401, "账号或密码错误")
	}
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(value)
	s.mu.Lock()
	defer s.mu.Unlock()
	for key, expires := range s.sessions {
		if time.Now().After(expires) {
			delete(s.sessions, key)
		}
	}
	s.sessions[token] = time.Now().Add(8 * time.Hour)
	return &dto.LoginResult{AccessToken: token}, nil
}
func (s *AuthService) Valid(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expires, ok := s.sessions[token]
	return ok && time.Now().Before(expires)
}
func (s *AuthService) Logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}
