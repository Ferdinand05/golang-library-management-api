package borrowing

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/member"

	"github.com/gin-gonic/gin"
)

func TestBorrowingHandler_CreateBorrowing(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 5)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, _ := json.Marshal(map[string]any{
		"member_id": testMember.ID,
		"book_id":   testBook.ID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/borrowings/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}

	var response struct {
		Borrowing BorrowingResponse `json:"borrowing"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	if response.Borrowing.ID == 0 {
		t.Error("expected borrowing ID to be generated")
	}

	var currentStock int
	db.Raw("SELECT stock FROM books WHERE id = ?", testBook.ID).Scan(&currentStock)
	if currentStock != 4 {
		t.Errorf("got stock %d, want 4", currentStock)
	}
}

func TestBorrowingHandler_CreateBorrowing_MemberNotFound(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 5)

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, _ := json.Marshal(map[string]any{
		"member_id": 999,
		"book_id":   testBook.ID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/borrowings/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestBorrowingHandler_CreateBorrowing_OutOfStock(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 0)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, _ := json.Marshal(map[string]any{
		"member_id": testMember.ID,
		"book_id":   testBook.ID,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/borrowings/", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
}

func TestBorrowingHandler_GetBorrowing(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 5)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	borrowingRepo := NewRepository(db)
	now := time.Now()
	created, _ := borrowingRepo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/borrowings/"+strconv.FormatUint(uint64(created.ID), 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestBorrowingHandler_ReturnBorrowing(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 4)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	borrowingRepo := NewRepository(db)
	now := time.Now()
	created, _ := borrowingRepo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/borrowings/"+strconv.FormatUint(uint64(created.ID), 10)+"/return", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var currentStock int
	db.Raw("SELECT stock FROM books WHERE id = ?", testBook.ID).Scan(&currentStock)
	if currentStock != 5 {
		t.Errorf("got stock %d, want 5", currentStock)
	}
}

func TestBorrowingHandler_GetBorrowing_NotFound(t *testing.T) {
	db := setupTestDB(t)

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/borrowings/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestBorrowingHandler_GetBorrowings(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 5)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	borrowingRepo := NewRepository(db)
	now := time.Now()
	borrowingRepo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/borrowings/", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Borrowings []BorrowingResponse `json:"borrowings"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	if len(response.Borrowings) < 1 {
		t.Errorf("got %d borrowings, want at least 1", len(response.Borrowings))
	}
}

func TestBorrowingHandler_GetBorrowingByMemberID(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 5)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	borrowingRepo := NewRepository(db)
	now := time.Now()
	borrowingRepo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/borrowings/member/"+strconv.FormatInt(testMember.ID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Borrowings []BorrowingResponse `json:"borrowings"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	if len(response.Borrowings) != 1 {
		t.Errorf("got %d borrowings, want 1", len(response.Borrowings))
	}
}

func TestBorrowingHandler_ReturnBorrowing_AlreadyReturned(t *testing.T) {
	db := setupTestDB(t)

	testAuthor := createTestAuthor(db, "Test Author")
	testCategory := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "978-1234567890", "Test Book", testAuthor.ID, testCategory.ID, 4)
	testMember := createTestMember(db, "MEM001", "John Doe", "john@example.com")

	borrowingRepo := NewRepository(db)
	now := time.Now()
	created, _ := borrowingRepo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	// First return
	borrowingRepo.Return(context.Background(), int64(created.ID))

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/borrowings/"+strconv.FormatUint(uint64(created.ID), 10)+"/return", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}

	var response struct {
		Error string `json:"error"`
	}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	if response.Error != "borrowing already returned" {
		t.Errorf("got error %q, want %q", response.Error, "borrowing already returned")
	}
}

func TestBorrowingHandler_ReturnBorrowing_NotFound(t *testing.T) {
	db := setupTestDB(t)

	bookRepo := book.NewRepository(db)
	memberRepo := member.NewRepository(db)
	borrowingRepo := NewRepository(db)
	service := NewService(db, borrowingRepo, memberRepo, bookRepo)
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/borrowings/999/return", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
