package author

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

func TestAuthorHandler_CreateAuthor(t *testing.T) {
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

	body := map[string]any{
		"name": "John Doe",
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/authors/",
		bytes.NewReader(bodyJSON),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"got status %d, want %d; body: %s",
			recorder.Code,
			http.StatusCreated,
			recorder.Body.String(),
		)
	}

	var response struct {
		Author AuthorResponse `json:"author"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Author.ID == 0 {
		t.Error("expected author ID to be generated")
	}

	if response.Author.Name != "John Doe" {
		t.Errorf("got name %s, want John Doe", response.Author.Name)
	}

	var count int64
	err = db.WithContext(context.Background()).
		Model(&Author{}).
		Where("id = ?", response.Author.ID).
		Count(&count).
		Error

	if err != nil {
		t.Fatalf("failed to verify author in database: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d author in database, want 1", count)
	}
}

func TestAuthorHandler_CreateAuthor_Invalid(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/authors/",
		bytes.NewReader([]byte(`{"invalid": true}`)),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestAuthorHandler_GetAuthor(t *testing.T) {
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

	author := Author{
		Name: "John Doe",
	}

	created, err := repo.Create(context.Background(), author)
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/authors/"+strconv.FormatInt(created.ID, 10),
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got %d, want %d, body %s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}

	var response struct {
		Author AuthorResponse `json:"author"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Author.ID != created.ID {
		t.Fatalf("got id %d, want id %d", response.Author.ID, created.ID)
	}

	if response.Author.Name != created.Name {
		t.Fatalf("got name %s, want %s", response.Author.Name, created.Name)
	}
}

func TestAuthorHandler_GetAuthor_NotFound(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/authors/999",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d, body %s",
			recorder.Code,
			http.StatusNotFound,
			recorder.Body.String(),
		)
	}

	var response struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "author not found" {
		t.Errorf("got error %q, want %q", response.Error, "author not found")
	}
}

func TestAuthorHandler_GetAuthor_InvalidID(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/authors/abc",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestAuthorHandler_GetAuthors(t *testing.T) {
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

	repo.Create(context.Background(), Author{Name: "Author 1"})
	repo.Create(context.Background(), Author{Name: "Author 2"})

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/authors/",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Authors []Author `json:"authors"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Authors) != 2 {
		t.Errorf("got %d authors, want 2", len(response.Authors))
	}
}

func TestAuthorHandler_UpdateAuthor(t *testing.T) {
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

	created, err := repo.Create(context.Background(), Author{Name: "Old Name"})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	body := map[string]any{
		"name": "New Name",
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/authors/"+strconv.FormatInt(created.ID, 10),
		bytes.NewReader(bodyJSON),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}

	var response struct {
		Author AuthorResponse `json:"author"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Author.Name != "New Name" {
		t.Errorf("got name %s, want New Name", response.Author.Name)
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated author: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestAuthorHandler_UpdateAuthor_NotFound(t *testing.T) {
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

	body := map[string]any{
		"name": "New Name",
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/authors/999",
		bytes.NewReader(bodyJSON),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestAuthorHandler_DeleteAuthor(t *testing.T) {
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

	created, err := repo.Create(context.Background(), Author{Name: "To Delete"})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/authors/"+strconv.FormatInt(created.ID, 10),
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrAuthorNotFound {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorHandler_DeleteAuthor_NotFound(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/authors/999",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
