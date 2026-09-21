package book

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type mockAuthorRepo struct {
	db *gorm.DB
}

func (m *mockAuthorRepo) FindAll(ctx context.Context) ([]author.Author, error) {
	return nil, nil
}

func (m *mockAuthorRepo) FindByID(ctx context.Context, id int64) (author.Author, error) {
	var a author.Author
	err := m.db.WithContext(ctx).First(&a, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return author.Author{}, author.ErrAuthorNotFound
		}
		return author.Author{}, err
	}
	return a, nil
}

func (m *mockAuthorRepo) Create(ctx context.Context, a author.Author) (author.Author, error) {
	return a, nil
}

func (m *mockAuthorRepo) Update(ctx context.Context, id int64, a author.Author) (author.Author, error) {
	return a, nil
}

func (m *mockAuthorRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

type mockCategoryRepo struct {
	db *gorm.DB
}

func (m *mockCategoryRepo) FindAll(ctx context.Context) ([]category.Category, error) {
	return nil, nil
}

func (m *mockCategoryRepo) FindByID(ctx context.Context, id int64) (category.Category, error) {
	var c category.Category
	err := m.db.WithContext(ctx).First(&c, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return category.Category{}, category.ErrCategoryNotFound
		}
		return category.Category{}, err
	}
	return c, nil
}

func (m *mockCategoryRepo) Create(ctx context.Context, c category.Category) (category.Category, error) {
	return c, nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, id int64, c category.Category) (category.Category, error) {
	return c, nil
}

func (m *mockCategoryRepo) Delete(ctx context.Context, id int64) error {
	return nil
}

func TestBookHandler_CreateBook(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"title":       "Test Book",
		"isbn":        "978-1234567890",
		"author_id":   authorID,
		"category_id": categoryID,
		"stock":       10,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Book BookResponse `json:"book"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Book.ID == 0 {
		t.Error("expected book ID to be generated")
	}

	if response.Book.Title != "Test Book" {
		t.Errorf("got title %s, want Test Book", response.Book.Title)
	}

	if response.Book.Author.ID != authorID {
		t.Errorf("got author ID %d, want %d", response.Book.Author.ID, authorID)
	}

	if response.Book.Category.ID != categoryID {
		t.Errorf("got category ID %d, want %d", response.Book.Category.ID, categoryID)
	}

	var count int64
	err = db.WithContext(context.Background()).Model(&Book{}).Where("id = ?", response.Book.ID).Count(&count).Error
	if err != nil {
		t.Fatalf("failed to verify book in database: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d book in database, want 1", count)
	}
}

func TestBookHandler_CreateBook_AuthorNotFound(t *testing.T) {
	db := setupTestDB(t)

	categoryID := createTestCategory(t, db, "Test Category")

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"title":       "Test Book",
		"isbn":        "978-1234567890",
		"author_id":   999,
		"category_id": categoryID,
		"stock":       10,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "author not found" {
		t.Errorf("got error %q, want %q", response.Error, "author not found")
	}
}

func TestBookHandler_CreateBook_CategoryNotFound(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"title":       "Test Book",
		"isbn":        "978-1234567890",
		"author_id":   authorID,
		"category_id": 999,
		"stock":       10,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/books/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Error != "category not found" {
		t.Errorf("got error %q, want %q", response.Error, "category not found")
	}
}

func TestBookHandler_GetBook(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	bookID := createTestBook(t, db, "978-1234567890", "Test Book", authorID, categoryID, 10)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/"+strconv.FormatInt(bookID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got %d, want %d, body %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Book BookResponse `json:"book"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Book.ID != bookID {
		t.Fatalf("got id %d, want id %d", response.Book.ID, bookID)
	}

	if response.Book.Title != "Test Book" {
		t.Fatalf("got title %s, want Test Book", response.Book.Title)
	}

	if response.Book.Author.ID != authorID {
		t.Errorf("got author ID %d, want %d", response.Book.Author.ID, authorID)
	}

	if response.Book.Category.ID != categoryID {
		t.Errorf("got category ID %d, want %d", response.Book.Category.ID, categoryID)
	}
}

func TestBookHandler_GetBook_NotFound(t *testing.T) {
	db := setupTestDB(t)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d, body %s", recorder.Code, http.StatusNotFound, recorder.Body.String())
	}
}

func TestBookHandler_GetBooks(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	createTestBook(t, db, "978-1234567890", "Book 1", authorID, categoryID, 5)
	createTestBook(t, db, "978-0987654321", "Book 2", authorID, categoryID, 3)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/books/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Books []BookResponse `json:"books"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Books) != 2 {
		t.Errorf("got %d books, want 2", len(response.Books))
	}
}

func TestBookHandler_UpdateBook(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	bookID := createTestBook(t, db, "978-1234567890", "Old Title", authorID, categoryID, 10)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	desc := "Updated description"
	bodyJSON, err := json.Marshal(map[string]any{
		"title":       "New Title",
		"description": desc,
		"author_id":   authorID,
		"category_id": categoryID,
		"stock":       15,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/books/"+strconv.FormatInt(bookID, 10), bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	saved, err := bookRepo.FindByID(context.Background(), bookID)
	if err != nil {
		t.Fatalf("failed to find updated book: %v", err)
	}

	if saved.Title != "New Title" {
		t.Errorf("got title %s, want New Title", saved.Title)
	}
}

func TestBookHandler_UpdateBook_AuthorNotFound(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	bookID := createTestBook(t, db, "978-1234567890", "Old Title", authorID, categoryID, 10)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	desc := "desc"
	bodyJSON, err := json.Marshal(map[string]any{
		"title":       "New Title",
		"description": desc,
		"author_id":   999,
		"category_id": categoryID,
		"stock":       15,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/books/"+strconv.FormatInt(bookID, 10), bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBookHandler_DeleteBook(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	bookID := createTestBook(t, db, "978-1234567890", "To Delete", authorID, categoryID, 10)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/books/"+strconv.FormatInt(bookID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	_, err := bookRepo.FindByID(context.Background(), bookID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBookHandler_GetBooksByAuthor(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	createTestBook(t, db, "978-1234567890", "Book 1", authorID, categoryID, 5)
	createTestBook(t, db, "978-0987654321", "Book 2", authorID, categoryID, 3)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/authors/"+strconv.FormatInt(authorID, 10)+"/books", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Books []BookResponse `json:"books"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Books) != 2 {
		t.Errorf("got %d books, want 2", len(response.Books))
	}
}

func TestBookHandler_GetBooksByCategory(t *testing.T) {
	db := setupTestDB(t)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")
	createTestBook(t, db, "978-1234567890", "Book 1", authorID, categoryID, 5)
	createTestBook(t, db, "978-0987654321", "Book 2", authorID, categoryID, 3)

	bookRepo := NewRepository(db)
	authorRepo := &mockAuthorRepo{db: db}
	categoryRepo := &mockCategoryRepo{db: db}
	service := NewService(bookRepo, authorRepo, categoryRepo, &fakeCacheClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+strconv.FormatInt(categoryID, 10)+"/books", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Books []BookResponse `json:"books"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Books) != 2 {
		t.Errorf("got %d books, want 2", len(response.Books))
	}
}
