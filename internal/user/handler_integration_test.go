package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupUserRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")

	api.GET("/users", handler.GetUsers)
	api.GET("/users/:id", handler.GetUser)
	api.POST("/users", handler.CreateUser)

	return router
}

func TestUserHandler_CreateUser(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"email":    "john@example.com",
		"password": "password123",
		"role":     "user",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		User UserResponse `json:"user"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.User.ID == 0 {
		t.Error("expected user ID to be generated")
	}

	if response.User.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", response.User.Email)
	}

	if response.User.Role != "user" {
		t.Errorf("got role %s, want user", response.User.Role)
	}

	var count int64
	err = db.WithContext(context.Background()).
		Model(&User{}).
		Where("id = ?", response.User.ID).
		Count(&count).
		Error
	if err != nil {
		t.Fatalf("failed to verify user in database: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d user in database, want 1", count)
	}
}

func TestUserHandler_CreateUser_Invalid(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader([]byte(`{"invalid": true}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_CreateUser_InvalidBinding(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"email":    "not-an-email",
		"password": "123",
		"role":     "superadmin",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	user := User{
		Email:    "john@example.com",
		Password: "password123",
		Role:     "user",
	}

	created, err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/users/"+strconv.FormatUint(uint64(created.ID), 10),
		nil,
	)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		User UserResponse `json:"user"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.User.ID != int64(created.ID) {
		t.Fatalf("got id %d, want %d", response.User.ID, created.ID)
	}

	if response.User.Email != created.Email {
		t.Fatalf("got email %s, want %s", response.User.Email, created.Email)
	}

	if response.User.Role != created.Role {
		t.Fatalf("got role %s, want %s", response.User.Role, created.Role)
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "user not found" {
		t.Errorf("got error %q, want %q", response.Error, "user not found")
	}
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/abc", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestUserHandler_GetUsers(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeCacheClient{})
	handler := NewHandler(service)
	router := setupUserRouter(handler)

	repo.Create(context.Background(), User{
		Email:    "user1@example.com",
		Password: "password123",
		Role:     "user",
	})

	repo.Create(context.Background(), User{
		Email:    "user2@example.com",
		Password: "password456",
		Role:     "admin",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Users []UserResponse `json:"users"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Users) != 2 {
		t.Errorf("got %d users, want 2", len(response.Users))
	}
}