package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func setupAuthRouter(jwtService *JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/protected", AuthMiddleware(jwtService), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString("userID"),
			"email":  c.GetString("email"),
			"role":   c.GetString("role"),
		})
	})

	return router
}

func TestAuthMiddleware_Success(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

	token, err := jwtService.GenerateToken(1, "john@example.com", "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		UserID string `json:"userID"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.UserID != "1" {
		t.Errorf("got userID %s, want 1", response.UserID)
	}

	if response.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", response.Email)
	}

	if response.Role != "admin" {
		t.Errorf("got role %s, want admin", response.Role)
	}
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	var response struct {
		Error string `json:"error"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "authorization header not found" {
		t.Errorf("got error %q, want %q", response.Error, "authorization header not found")
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	var response struct {
		Error string `json:"error"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "invalid token format" {
		t.Errorf("got error %q, want %q", response.Error, "invalid token format")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	var response struct {
		Error string `json:"error"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "invalid token" {
		t.Errorf("got error %q, want %q", response.Error, "invalid token")
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

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

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_EmptyBearerToken(t *testing.T) {
	jwtService := NewJWTService("secret")
	router := setupAuthRouter(jwtService)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
    jwtService := NewJWTService("secret")

    tokenService := NewJWTService("different-secret")

    token, err := tokenService.GenerateToken(
        1,
        "john@example.com",
        "admin",
    )
    if err != nil {
        t.Fatalf("failed to generate token: %v", err)
    }

    router := setupAuthRouter(jwtService)

    req := httptest.NewRequest(
        http.MethodGet,
        "/protected",
        nil,
    )
    req.Header.Set("Authorization", "Bearer "+token)

    recorder := httptest.NewRecorder()

    router.ServeHTTP(recorder, req)

    if recorder.Code != http.StatusUnauthorized {
        t.Fatalf(
            "got status %d, want %d",
            recorder.Code,
            http.StatusUnauthorized,
        )
    }
}