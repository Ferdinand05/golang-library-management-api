package member

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

func TestMemberHandler_CreateMember(t *testing.T) {

	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")

	RegisterRoutes(api, handler)

	body := map[string]any{
		"name":    "John Doe",
		"email":   "john@example.com",
		"phone":   "08123456789",
		"address": "Jakarta",
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/members",
		bytes.NewReader(bodyJson),
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
		Member MemberResponse `json:"member"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Member.ID == 0 {
		t.Error("expected member ID to be generated")
	}

	if response.Member.Name != "John Doe" {
		t.Errorf(
			"got name %s, want John Doe",
			response.Member.Name,
		)
	}

	if response.Member.Email != "john@example.com" {
		t.Errorf(
			"got email %s, want john@example.com",
			response.Member.Email,
		)
	}

	if !IsValidMemberCode(response.Member.MemberCode) {
		t.Errorf(
			"invalid member code: %s",
			response.Member.MemberCode,
		)
	}

	var count int64
	err = db.WithContext(context.Background()).
		Model(&Member{}).
		Where("id = ?", response.Member.ID).
		Count(&count).
		Error

	if err != nil {
		t.Fatalf("failed to verify member in database: %v", err)
	}

	if count != 1 {
		t.Errorf("got %d member in database, want 1", count)
	}

}

func TestMemberHandler_GetMember_NotFound(t *testing.T) {
	// setup database, repo, service, handler, router
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")

	RegisterRoutes(api, handler)

	// request GET /api/v1/members/999

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/members/999",
		nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// assert 404

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d, body %s",
			recorder.Code,
			http.StatusNotFound,
			recorder.Body.String(),
		)
	}

	// assert response error == "member not found"
	var response struct {
		Error string `json:"error"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)

	if err != nil {
		t.Fatalf("failed to decode response: %v",err)
	}

	if response.Error != "member not found" {
		t.Errorf(
			"got error %q, want %q",
			response.Error,
			"member not found",
		)
	}

}


func TestMemberHandler_GetMember(t *testing.T) {
		db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")

	RegisterRoutes(api, handler)

	member := Member{
		MemberCode: "ABC1234",
		Name: "John Doe",
		Phone: "0812311233",
		Email: "john@gmail.com",
	}

	created,err := repo.Create(context.Background(),member)

	if err != nil {
		t.Fatalf("failed to create member: %v",err)
	}

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/members/"+strconv.FormatInt(created.ID, 10),
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
		Member MemberResponse `json:"member"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)

	if err != nil {
		t.Fatalf("failed to decode response: %v",err)
	}

	if response.Member.ID != created.ID {
		t.Fatalf("got id %d, want id %d",response.Member.ID,created.ID)
	}

	if response.Member.Name != created.Name {
		t.Fatalf("failed member name : got %s, want %s",response.Member.Name,created.Name)
	}

	if response.Member.Email != created.Email {
		t.Fatalf("failed member email : got %s, want %s", response.Member.Email, created.Email)
	}
}

func TestMemberHandler_GetMembers(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Member 1",
		Email:      "member1@example.com",
		Phone:      "08111111111",
	})

	repo.Create(context.Background(), Member{
		MemberCode: "DEF5678",
		Name:       "Member 2",
		Email:      "member2@example.com",
		Phone:      "08222222222",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/members", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Members []MemberResponse `json:"members"`
	}

	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Members) != 2 {
		t.Errorf("got %d members, want 2", len(response.Members))
	}
}

func TestMemberHandler_UpdateMember(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	created, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Old Name",
		Email:      "old@example.com",
		Phone:      "08111111111",
	})
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	newAddr := "New Address"
	bodyJSON, err := json.Marshal(map[string]any{
		"name":    "New Name",
		"email":   "new@example.com",
		"phone":   "08222222222",
		"address": newAddr,
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/members/"+strconv.FormatInt(created.ID, 10),
		bytes.NewReader(bodyJSON),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d; body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var response struct {
		Member MemberResponse `json:"member"`
	}

	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Member.Name != "New Name" {
		t.Errorf("got name %s, want New Name", response.Member.Name)
	}

	if response.Member.Email != "new@example.com" {
		t.Errorf("got email %s, want new@example.com", response.Member.Email)
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated member: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestMemberHandler_UpdateMember_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	bodyJSON, err := json.Marshal(map[string]any{
		"name":    "New Name",
		"email":   "new@example.com",
		"phone":   "08222222222",
		"address": "New Address",
	})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/members/999", bytes.NewReader(bodyJSON))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestMemberHandler_DeleteMember(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	created, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "To Delete",
		Email:      "delete@example.com",
		Phone:      "08123456789",
	})
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/members/"+strconv.FormatInt(created.ID, 10), nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusOK)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestMemberHandler_DeleteMember_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/members/999", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestMemberHandler_GetMember_InvalidID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/members/abc", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestMemberHandler_CreateMember_Invalid(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)
	service := NewService(repo, &fakeRedisClient{})
	handler := NewHandler(service)

	gin.SetMode(gin.TestMode)

	router := gin.New()
	api := router.Group("/api/v1")
	RegisterRoutes(api, handler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/members", bytes.NewReader([]byte(`{"invalid": true}`)))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

