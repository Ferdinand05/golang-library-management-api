package author

import (
	"context"
	"errors"
	"testing"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type fakeAuthorRepository struct {
	findAllResult  []Author
	findAllErr     error
	findByIDResult Author
	findByIDErr    error
	createResult   Author
	createErr      error
	updateResult   Author
	updateErr      error
	deleteErr      error
}

func (f *fakeAuthorRepository) FindAll(ctx context.Context) ([]Author, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeAuthorRepository) FindByID(ctx context.Context, id int64) (Author, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeAuthorRepository) Create(ctx context.Context, author Author) (Author, error) {
	return f.createResult, f.createErr
}

func (f *fakeAuthorRepository) Update(ctx context.Context, id int64, author Author) (Author, error) {
	return f.updateResult, f.updateErr
}

func (f *fakeAuthorRepository) Delete(ctx context.Context, id int64) error {
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

func TestAuthorService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []Author
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []Author{
				{ID: 1, Name: "Author 1"},
				{ID: 2, Name: "Author 2"},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []Author{},
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
			repo := &fakeAuthorRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d authors, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestAuthorService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult Author
		repoErr    error
		wantErr    bool
		wantID     int64
	}{
		{
			name:       "success",
			id:         1,
			repoResult: Author{ID: 1, Name: "John Doe"},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: Author{},
			repoErr:    ErrAuthorNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: Author{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAuthorRepository{
				findByIDResult: tt.repoResult,
				findByIDErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			got, err := service.FindByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("FindByID() got ID %d, want %d", got.ID, tt.wantID)
			}

			if tt.repoErr == ErrAuthorNotFound && !errors.Is(err, ErrAuthorNotFound) {
				t.Errorf("FindByID() expected ErrAuthorNotFound, got %v", err)
			}
		})
	}
}

func TestAuthorService_Create(t *testing.T) {
	tests := []struct {
		name       string
		request    CreateAuthorRequest
		repoResult Author
		repoErr    error
		wantErr    bool
		wantName   string
	}{
		{
			name:       "success",
			request:    CreateAuthorRequest{Name: "Jane Doe"},
			repoResult: Author{ID: 1, Name: "Jane Doe"},
			repoErr:    nil,
			wantErr:    false,
			wantName:   "Jane Doe",
		},
		{
			name:       "repository error",
			request:    CreateAuthorRequest{Name: "John Doe"},
			repoResult: Author{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAuthorRepository{
				createResult: tt.repoResult,
				createErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

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

func TestAuthorService_Update(t *testing.T) {
	tests := []struct {
		name         string
		id           int64
		request      UpdateAuthorRequest
		repoResult   Author
		repoErr      error
		wantErr      bool
		wantName     string
	}{
		{
			name:       "success",
			id:         1,
			request:    UpdateAuthorRequest{Name: "Jane Doe Updated"},
			repoResult: Author{ID: 1, Name: "Jane Doe Updated"},
			repoErr:    nil,
			wantErr:    false,
			wantName:   "Jane Doe Updated",
		},
		{
			name:       "not found",
			id:         999,
			request:    UpdateAuthorRequest{Name: "Non Existent"},
			repoResult: Author{},
			repoErr:    ErrAuthorNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			request:    UpdateAuthorRequest{Name: "John Doe"},
			repoResult: Author{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeAuthorRepository{
				updateResult: tt.repoResult,
				updateErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			got, err := service.Update(context.Background(), tt.id, tt.request)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.Name != tt.wantName {
				t.Errorf("Update() got name %s, want %s", got.Name, tt.wantName)
			}
		})
	}
}

func TestAuthorService_Delete(t *testing.T) {
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
			repoErr: ErrAuthorNotFound,
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
			repo := &fakeAuthorRepository{
				deleteErr: tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			err := service.Delete(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.repoErr == ErrAuthorNotFound && !errors.Is(err, ErrAuthorNotFound) {
				t.Errorf("Delete() expected ErrAuthorNotFound, got %v", err)
			}
		})
	}
}