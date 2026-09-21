package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeAuthService struct {
	loginToken string
	loginErr   error
}

func (f *fakeAuthService) Login(ctx context.Context, request LoginRequest) (string, error) {
	return f.loginToken, f.loginErr
}

func setupAuthHandlerRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/login", handler.Login)

	return router
}

func TestAuthHandler_LoginSuccess(t *testing.T) {
	service := &fakeAuthService{
		loginToken: "valid.jwt.token",
	}
	handler := NewHandler(service)
	router := setupAuthHandlerRouter(handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"email":    "john@example.com",
		"password": "password123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Token != "valid.jwt.token" {
		t.Errorf("got token %q, want valid.jwt.token", response.Token)
	}
}

func TestAuthHandler_LoginInvalidRequestBody(t *testing.T) {
	service := &fakeAuthService{}
	handler := NewHandler(service)
	router := setupAuthHandlerRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`{"email": 1}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "invalid request body" {
		t.Errorf("got error %q, want invalid request body", response.Error)
	}
}

func TestAuthHandler_LoginInvalidCredentials(t *testing.T) {
	service := &fakeAuthService{
		loginErr: ErrInvalidCredentials,
	}
	handler := NewHandler(service)
	router := setupAuthHandlerRouter(handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"email":    "john@example.com",
		"password": "wrongpassword",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusUnauthorized, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "invalid credentials" {
		t.Errorf("got error %q, want invalid credentials", response.Error)
	}
}

func TestAuthHandler_LoginServiceError(t *testing.T) {
	service := &fakeAuthService{
		loginErr: errors.New("database error"),
	}
	handler := NewHandler(service)
	router := setupAuthHandlerRouter(handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"email":    "john@example.com",
		"password": "password123",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusInternalServerError, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "internal server error" {
		t.Errorf("got error %q, want internal server error", response.Error)
	}
}
