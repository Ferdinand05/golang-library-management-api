package member

import (
	"context"
	"errors"
	"testing"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

func TestIsValidMemberCode(t *testing.T) {

	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "Valid member code",
			code: "ABC1234",
			want: true,
		},
		{
			name: "Invalid member code",
			code: "ABC",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidMemberCode(tt.code)

			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}

		})
	}

}

type fakeMemberRepository struct {
	findAllResult  []Member
	findAllErr     error
	findByIDResult Member
	findByIDErr    error
	createdMember  Member
	createErr      error
	receivedMember Member
	updateResult   Member
	updateErr      error
	deleteErr      error
}

func (f *fakeMemberRepository) FindAll(
	ctx context.Context,
) ([]Member, error) {
	return f.findAllResult, f.findAllErr
}

func (f *fakeMemberRepository) FindByID(
	ctx context.Context,
	id int64,
) (Member, error) {
	return f.findByIDResult, f.findByIDErr
}

func (f *fakeMemberRepository) Create(
	ctx context.Context,
	member Member,
) (Member, error) {
	f.receivedMember = member
	return f.createdMember, f.createErr
}

func (f *fakeMemberRepository) Update(
	ctx context.Context,
	id int64,
	member Member,
) (Member, error) {
	return f.updateResult, f.updateErr
}

func (f *fakeMemberRepository) Delete(
	ctx context.Context,
	id int64,
) error {
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

func TestMemberServiceCreate(t *testing.T) {

	repo := &fakeMemberRepository{
		createdMember: Member{
			ID:   1,
			Name: "John",
		},
	}

	service := NewService(repo, &fakeCacheClient{})

	address := "Jakarta"
	got, err := service.Create(
		context.Background(),
		CreateMemberRequest{
			Name:    "John",
			Email:   "john@gmail.com",
			Phone:   "0891238123",
			Address: &address,
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != 1 {
		t.Errorf("got ID %d, want 1", got.ID)
	}

	if repo.receivedMember.Name != "John" {
		t.Errorf(
			"got name %s, want John",
			repo.receivedMember.Name,
		)
	}

	if !IsValidMemberCode(repo.receivedMember.MemberCode) {
		t.Errorf(
			"invalid member code: %s",
			repo.receivedMember.MemberCode,
		)
	}

}

func TestMemberServiceCreateRepositoryError(t *testing.T) {
	expectedErr := errors.New("database error")

	repo := &fakeMemberRepository{
		createErr: expectedErr,
	}

	service := NewService(repo, &fakeCacheClient{})

	address := "Jakarta"

	_, err := service.Create(
		context.Background(),
		CreateMemberRequest{
			Name:    "John",
			Email:   "john@example.com",
			Phone:   "08123456789",
			Address: &address,
		},
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}
}

func TestMemberService_FindAll(t *testing.T) {
	tests := []struct {
		name       string
		repoResult []Member
		repoErr    error
		wantErr    bool
		wantLen    int
	}{
		{
			name: "success",
			repoResult: []Member{
				{ID: 1, Name: "Member 1", MemberCode: "ABC1234"},
				{ID: 2, Name: "Member 2", MemberCode: "XYZ5678"},
			},
			repoErr: nil,
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "success empty",
			repoResult: []Member{},
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
			repo := &fakeMemberRepository{
				findAllResult: tt.repoResult,
				findAllErr:    tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			got, err := service.FindAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Fatalf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("FindAll() got %d members, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func TestMemberService_FindByID(t *testing.T) {
	tests := []struct {
		name       string
		id         int64
		repoResult Member
		repoErr    error
		wantErr    bool
		wantID     int64
	}{
		{
			name:       "success",
			id:         1,
			repoResult: Member{ID: 1, Name: "John Doe", MemberCode: "ABC1234"},
			repoErr:    nil,
			wantErr:    false,
			wantID:     1,
		},
		{
			name:       "not found",
			id:         999,
			repoResult: Member{},
			repoErr:    ErrMemberNotFound,
			wantErr:    true,
		},
		{
			name:       "repository error",
			id:         1,
			repoResult: Member{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeMemberRepository{
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

			if tt.repoErr == ErrMemberNotFound && !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("FindByID() expected ErrMemberNotFound, got %v", err)
			}
		})
	}
}

func TestMemberService_Update(t *testing.T) {
	address := "Updated Address"
	tests := []struct {
		name       string
		id         int64
		request    UpdateMemberRequest
		repoResult Member
		repoErr    error
		wantErr    bool
		wantName   string
	}{
		{
			name: "success",
			id:   1,
			request: UpdateMemberRequest{
				Name:    "Updated Name",
				Email:   "updated@example.com",
				Phone:   "08111111111",
				Address: &address,
			},
			repoResult: Member{ID: 1, Name: "Updated Name", Email: "updated@example.com"},
			repoErr:    nil,
			wantErr:    false,
			wantName:   "Updated Name",
		},
		{
			name: "not found",
			id:   999,
			request: UpdateMemberRequest{
				Name:    "Updated Name",
				Email:   "updated@example.com",
				Phone:   "08111111111",
				Address: &address,
			},
			repoResult: Member{},
			repoErr:    ErrMemberNotFound,
			wantErr:    true,
			wantName:   "",
		},
		{
			name: "repository error",
			id:   1,
			request: UpdateMemberRequest{
				Name:    "Updated Name",
				Email:   "updated@example.com",
				Phone:   "08111111111",
				Address: &address,
			},
			repoResult: Member{},
			repoErr:    errors.New("database error"),
			wantErr:    true,
			wantName:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeMemberRepository{
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

			if tt.repoErr == ErrMemberNotFound && !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("Update() expected ErrMemberNotFound, got %v", err)
			}
		})
	}
}

func TestMemberService_Delete(t *testing.T) {
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
			repoErr: ErrMemberNotFound,
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
			repo := &fakeMemberRepository{
				deleteErr: tt.repoErr,
			}
			service := NewService(repo, &fakeCacheClient{})

			err := service.Delete(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.repoErr == ErrMemberNotFound && !errors.Is(err, ErrMemberNotFound) {
				t.Errorf("Delete() expected ErrMemberNotFound, got %v", err)
			}
		})
	}
}
