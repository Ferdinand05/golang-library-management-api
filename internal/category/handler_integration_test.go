package category

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
	"github.com/gin-gonic/gin"
)

type fakeRedisClient struct{}

func (f *fakeRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, redisCache.ErrCacheMiss
}

func (f *fakeRedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (f *fakeRedisClient) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func TestCategoryHandler_CreateCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{"name": "Fiction"})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Category CategoryResponse `json:"category"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Category.ID == 0 {
		t.Error("expected category ID to be generated")
	}

	if response.Category.Name != "Fiction" {
		t.Errorf("got name %s, want Fiction", response.Category.Name)
	}

	var count int64
	err = db.WithContext(context.Background()).Model(&Category{}).Where("id = ?", response.Category.ID).Count(&count).Error
	if err != nil {
		t.Fatalf("failed to verify category in database: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d category in database, want 1", count)
	}
}

func TestCategoryHandler_CreateCategory_Invalid(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/", bytes.NewReader([]byte(`{"invalid": true}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCategoryHandler_GetCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	created, err := repo.Create(context.Background(), Category{Name: "Fiction"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+strconv.FormatInt(created.ID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got %d, want %d, body %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Category CategoryResponse `json:"category"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Category.ID != created.ID {
		t.Fatalf("got id %d, want id %d", response.Category.ID, created.ID)
	}

	if response.Category.Name != created.Name {
		t.Fatalf("got name %s, want %s", response.Category.Name, created.Name)
	}
}

func TestCategoryHandler_GetCategory_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d, body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestCategoryHandler_GetCategory_InvalidID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/abc", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestCategoryHandler_GetCategories(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	repo.Create(context.Background(), Category{Name: "Category 1"})
	repo.Create(context.Background(), Category{Name: "Category 2"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Categories []Category `json:"categories"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Categories) != 2 {
		t.Errorf("got %d categories, want 2", len(response.Categories))
	}
}

func TestCategoryHandler_UpdateCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	created, err := repo.Create(context.Background(), Category{Name: "Old Name"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	bodyJSON, err := json.Marshal(map[string]any{"name": "New Name"})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/categories/"+strconv.FormatInt(created.ID, 10), bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated category: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestCategoryHandler_UpdateCategory_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{"name": "New Name"})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/categories/999", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestCategoryHandler_DeleteCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	created, err := repo.Create(context.Background(), Category{Name: "To Delete"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+strconv.FormatInt(created.ID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryHandler_DeleteCategory_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
