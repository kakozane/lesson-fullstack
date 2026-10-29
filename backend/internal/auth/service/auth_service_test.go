package service

import (
	"sync"
	"testing"
	"time"

	"lesson06/backend/internal/auth/dto"
)

func TestSessions(t *testing.T) {
	s := NewAuthService("test-password")
	if _, err := s.Login(dto.LoginRequest{Username: "admin", Password: "wrong"}); err == nil {
		t.Fatal("wrong password accepted")
	}
	s.sessions["expired"] = time.Now().Add(-time.Second)
	if s.Valid("expired") {
		t.Fatal("expired session accepted")
	}
	result, err := s.Login(dto.LoginRequest{Username: "admin", Password: "test-password"})
	if err != nil {
		t.Fatal(err)
	}
	if !s.Valid(result.AccessToken) || len(s.sessions) != 1 {
		t.Fatal("login or cleanup failed")
	}
	s.Logout(result.AccessToken)
	if s.Valid(result.AccessToken) {
		t.Fatal("logout did not revoke token")
	}
}
func TestConcurrentSessions(t *testing.T) {
	s := NewAuthService("test-password")
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			result, err := s.Login(dto.LoginRequest{Username: "admin", Password: "test-password"})
			if err != nil {
				t.Error(err)
				return
			}
			if !s.Valid(result.AccessToken) {
				t.Error("session missing")
			}
			s.Logout(result.AccessToken)
		})
	}
	wg.Wait()
}
