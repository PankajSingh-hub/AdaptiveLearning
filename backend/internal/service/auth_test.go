package service

import (
	"testing"

	"adaptive-assessment/internal/models"
)

func TestTokenRoundTrip(t *testing.T) {
	s := NewAuthService(nil, "test-secret", 1)
	token, err := s.generateToken(models.User{ID: 7, Role: models.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := s.ParseToken(token)
	if err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	if claims.UserID != 7 || claims.Role != "ADMIN" {
		t.Fatalf("claims wrong: %+v", claims)
	}
}

func TestTokenSignedWithOtherSecretIsRejected(t *testing.T) {
	attacker := NewAuthService(nil, "attacker-secret", 1)
	token, _ := attacker.generateToken(models.User{ID: 1, Role: models.RoleAdmin})
	server := NewAuthService(nil, "real-secret", 1)
	if _, err := server.ParseToken(token); err == nil {
		t.Fatal("token signed with a different secret must be rejected")
	}
}

func TestExpiredTokenIsRejected(t *testing.T) {
	s := NewAuthService(nil, "test-secret", -1) // already expired when issued
	token, _ := s.generateToken(models.User{ID: 1, Role: models.RoleCandidate})
	if _, err := s.ParseToken(token); err == nil {
		t.Fatal("expired token must be rejected")
	}
}

func TestTamperedTokenIsRejected(t *testing.T) {
	s := NewAuthService(nil, "test-secret", 1)
	token, _ := s.generateToken(models.User{ID: 1, Role: models.RoleCandidate})
	if _, err := s.ParseToken(token + "x"); err == nil {
		t.Fatal("tampered token must be rejected")
	}
}