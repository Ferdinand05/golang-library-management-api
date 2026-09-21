package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT_GenerateToken(t *testing.T) {
	service := NewJWTService("secret")

	token, err := service.GenerateToken(1, "john@example.com", "admin")

	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("expected token, got empty string")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("got token parts %d, want 3", len(parts))
	}
}

func TestJWT_ValidateToken(t *testing.T) {
	service := NewJWTService("secret")

	token, err := service.GenerateToken(1, "john@example.com", "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.Subject != "1" {
		t.Errorf("got subject %s, want 1", claims.Subject)
	}

	if claims.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", claims.Email)
	}

	if claims.Role != "admin" {
		t.Errorf("got role %s, want admin", claims.Role)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected expires at, got nil")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Fatal("expected expires at in the future")
	}
}

func TestJWT_ValidateToken_Expired(t *testing.T) {
	service := NewJWTService("secret")

	claims := Claims{
		Email: "john@example.com",
		Role:  "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = service.ValidateToken(signedToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "token expired") {
		t.Errorf("got error %v, want token expired", err)
	}
}

func TestJWT_ValidateToken_Invalid(t *testing.T) {
	service := NewJWTService("secret")

	_, err := service.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "token invalid") {
		t.Errorf("got error %v, want token invalid", err)
	}
}

func TestJWT_ValidateToken_WrongSecret(t *testing.T) {
	service := NewJWTService("secret")
	otherService := NewJWTService("other-secret")

	token, err := service.GenerateToken(1, "john@example.com", "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = otherService.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
