package category

import (
	"context"
	"errors"
	"testing"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type fakeCategoryRepository struct {
	findAllResult  []Category
	findAllErr     error
	findByIDResult Category
	findByIDErr    error
	createResult   Category
	createErr      error
	updateResult   Category
	updateErr      error
	deleteErr      error
}

func (f *fakeCategoryRepository) FindAll(ctx context.Context) ([]Category, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeCategoryRepository) FindByID(ctx context.Context, id int64) (Category, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeCategoryRepository) Create(ctx context.Context, category Category) (Category, error) {
	return f.createResult, f.createErr
}

func (f *fakeCategoryRepository) Update(ctx context.Context, id int64, category Category) (Category, error) {
	return f.updateResult, f.updateErr
}

func (f *fakeCategoryRepository) Delete(ctx context.Context, id int64) error {
	return f.deleteErr
}

type fakeCacheClient struct{}

func (f *fakeCacheClient) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, redisCache.ErrCacheMiss
}

func (f *fakeCacheClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

func (f *fakeCacheClient) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func TestCategoryService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []Category
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []Category{
				{ID: 1, Name: "Fiction"},
				{ID: 2, Name: "Non-Fiction"},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []Category{},
			repoErr:    nil,
			wantErr:    false,
			wantLen:    0,
		},
		{
			name:       "repository error",
			repoResult: nil,
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCategoryRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d categories, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestCategoryService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult Category
		repoErr    error
		wantErr    bool
		wantID     int64
	}{
		{
			name:       "success",
			id:         1,
			repoResult: Category{ID: 1, Name: "Fiction"},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: Category{},
			repoErr:    ErrCategoryNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: Category{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCategoryRepository{
				findByIDResult: tt.repoResult,
				findByIDErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeRedisClient{})

			got, err := service.FindByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("FindByID() got ID %d, want %d", got.ID, tt.wantID)
			}

			if tt.repoErr == ErrCategoryNotFound && !errors.Is(err, ErrCategoryNotFound) {
				t.Errorf("FindByID() expected ErrCategoryNotFound, got %v", err)
			}
		})
	}
}

func TestCategoryService_Create(t *testing.T) {
	tests := []struct {
		name       string
		request    CreateCategoryRequest
		repoResult Category
		repoErr    error
		wantErr    bool
		wantName   string
	}{
		{
			name:       "success",
			request:    CreateCategoryRequest{Name: "Science"},
			repoResult: Category{ID: 1, Name: "Science"},
			repoErr:    nil,
			wantErr:    false,
			wantName:   "Science",
		},
		{
			name:       "repository error",
			request:    CreateCategoryRequest{Name: "History"},
			repoResult: Category{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCategoryRepository{
				createResult: tt.repoResult,
				createErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeRedisClient{})

			got, err := service.Create(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.Name != tt.wantName {
				t.Errorf("Create() got name %s, want %s", got.Name, tt.wantName)
			}
		})
	}
}

func TestCategoryService_Update(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		request    UpdateCategoryRequest
		repoResult Category
		repoErr    error
		wantErr    bool
		wantName   string
	}{
		{
			name:       "success",
			id:         1,
			request:    UpdateCategoryRequest{Name: "Updated Name"},
			repoResult: Category{ID: 1, Name: "Updated Name"},
			repoErr:    nil,
			wantErr:    false,
			wantName:   "Updated Name",
		},
		{
			name:       "not found",
			id:         999,
			request:    UpdateCategoryRequest{Name: "Updated Name"},
			repoResult: Category{},
			repoErr:    ErrCategoryNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			request:    UpdateCategoryRequest{Name: "Updated Name"},
			repoResult: Category{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCategoryRepository{
				updateResult: tt.repoResult,
				updateErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeRedisClient{})

			got, err := service.Update(context.Background(), tt.id, tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.Name != tt.wantName {
				t.Errorf("Update() got name %s, want %s", got.Name, tt.wantName)
			}

			if tt.repoErr == ErrCategoryNotFound && !errors.Is(err, ErrCategoryNotFound) {
				t.Errorf("Update() expected ErrCategoryNotFound, got %v", err)
			}
		})
	}
}

func TestCategoryService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		repoErr error
		wantErr bool
	}{
		{
			name:    "success",
			id:      1,
			repoErr: nil,
			wantErr: false,
		},
		{
			name:    "not found",
			id:      999,
			repoErr: ErrCategoryNotFound,
			wantErr: true,
		},
		{
			name:    "repository error",
			id:      1,
			repoErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeCategoryRepository{
				deleteErr: tt.repoErr,
			}
			service := NewService(repo, &fakeRedisClient{})

			err := service.Delete(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.repoErr == ErrCategoryNotFound && !errors.Is(err, ErrCategoryNotFound) {
				t.Errorf("Delete() expected ErrCategoryNotFound, got %v", err)
			}
		})
	}
}
