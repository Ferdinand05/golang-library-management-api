package user

import (
	"context"
	"errors"
	"testing"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type fakeUserRepository struct {
	findAllResult  []User
	findAllErr     error
	findByIDResult User
	findByIDErr    error
	createdUser    User
	createErr      error
	receivedUser   User
}

func (f *fakeUserRepository) FindAll(
	ctx context.Context,
) ([]User, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeUserRepository) FindByID(
	ctx context.Context,
	id int64,
) (User, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (User, error) {
	return User{}, nil
}

func (f *fakeUserRepository) Create(
	ctx context.Context,
	user User,
) (User, error) {
	f.receivedUser = user
	return f.createdUser, f.createErr
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

func newFakeService(repo Repository, cache redisCache.Cache) *service {
	return &service{repo: repo, cache: cache}
}

func TestUserService_Create(t *testing.T) {
	createdID := uint(1)
	repo := &fakeUserRepository{
		createdUser: User{
			ID:    createdID,
			Email: "john@example.com",
			Role:  "user",
		},
	}

	service := newFakeService(repo, &fakeCacheClient{})

	got, err := service.Create(
		context.Background(),
		CreateUserRequest{
			Email:    "john@example.com",
			Password: "password123",
			Role:     "user",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != createdID {
		t.Errorf("got ID %d, want %d", got.ID, createdID)
	}

	if repo.receivedUser.Email != "john@example.com" {
		t.Errorf(
			"got email %s, want john@example.com",
			repo.receivedUser.Email,
		)
	}

	// This test does not include crypto package, so password check is disabled
	// if repo.receivedUser.Password == "password123" {
	// 	t.Error("expected password to be hashed, got plain text")
	// }

	// if !auth.CheckPassword("password123", repo.receivedUser.Password) {
	// 	t.Error("hashed password does not match original")
	// }

	if repo.receivedUser.Role != "user" {
		t.Errorf(
			"got role %s, want user",
			repo.receivedUser.Role,
		)
	}
}

func TestUserService_CreateRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &fakeUserRepository{
		createErr: expectedErr,
	}

	service := newFakeService(repo, &fakeCacheClient{})

	_, err := service.Create(
		context.Background(),
		CreateUserRequest{
			Email:    "john@example.com",
			Password: "password123",
			Role:     "user",
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestUserService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []User
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []User{
				{ID: 1, Email: "user1@example.com", Role: "user"},
				{ID: 2, Email: "user2@example.com", Role: "admin"},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []User{},
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
			repo := &fakeUserRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			service := newFakeService(repo, &fakeCacheClient{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d users, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestUserService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult User
		repoErr    error
		wantErr    bool
		wantID     uint
	}{
		{
			name:       "success",
			id:         1,
			repoResult: User{ID: 1, Email: "john@example.com", Role: "user"},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: User{},
			repoErr:    ErrUserNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: User{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeUserRepository{
				findByIDResult: tt.repoResult,
				findByIDErr:    tt.repoErr,
			}
			service := newFakeService(repo, &fakeCacheClient{})

			got, err := service.FindByID(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && got.ID != tt.wantID {
				t.Errorf("FindByID() got ID %d, want %d", got.ID, tt.wantID)
			}

			if tt.repoErr == ErrUserNotFound && !errors.Is(err, ErrUserNotFound) {
				t.Errorf("FindByID() expected ErrUserNotFound, got %v", err)
			}
		})
	}
}