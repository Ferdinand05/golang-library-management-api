package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ferdinand/library-management-system-api/internal/crypto"
	"ferdinand/library-management-system-api/internal/user"
)

type fakeUserRepository struct {
	findByEmailUser user.User
	findByEmailErr  error
}

func (f *fakeUserRepository) FindAll(ctx context.Context) ([]user.User, error) {
	return nil, nil
}

func (f *fakeUserRepository) FindByID(ctx context.Context, id int64) (user.User, error) {
	return user.User{}, nil
}

func (f *fakeUserRepository) FindByEmail(ctx context.Context, email string) (user.User, error) {
	if f.findByEmailErr != nil {
		return user.User{}, f.findByEmailErr
	}
	return f.findByEmailUser, nil
}

func (f *fakeUserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	return user.User{}, nil
}

type fakeJWTGenerator struct {
	generateTokenErr error
}

func (f *fakeJWTGenerator) GenerateToken(userID int64, email string, role string) (string, error) {
	if f.generateTokenErr != nil {
		return "", f.generateTokenErr
	}
	return "fake.token.value", nil
}

func hashPassword(password string) string {
	hashed, _ := crypto.HashPassword(password)
	return hashed
}

func TestAuthService_Login_Success(t *testing.T) {
	hashed := hashPassword("password123")

	userRepo := &fakeUserRepository{
		findByEmailUser: user.User{
			ID:       1,
			Email:    "john@example.com",
			Password: hashed,
			Role:     "admin",
		},
	}

	jwtService := NewJWTService("test-secret")
	service := NewService(userRepo, jwtService)

	token, err := service.Login(context.Background(), LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected token, got empty string")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("got token parts %d, want 3", len(parts))
	}
}

func TestAuthService_Login_EmailNotFound(t *testing.T) {
	userRepo := &fakeUserRepository{
		findByEmailErr: user.ErrUserNotFound,
	}

	jwtService := NewJWTService("test-secret")
	service := NewService(userRepo, jwtService)

	token, err := service.Login(context.Background(), LoginRequest{
		Email:    "notfound@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got error %v, want ErrInvalidCredentials", err)
	}

	if token != "" {
		t.Errorf("got token %q, want empty", token)
	}
}

func TestAuthService_Login_PasswordWrong(t *testing.T) {
	hashed := hashPassword("password123")

	userRepo := &fakeUserRepository{
		findByEmailUser: user.User{
			ID:       1,
			Email:    "john@example.com",
			Password: hashed,
			Role:     "user",
		},
	}

	jwtService := NewJWTService("test-secret")
	service := NewService(userRepo, jwtService)

	token, err := service.Login(context.Background(), LoginRequest{
		Email:    "john@example.com",
		Password: "wrongpassword",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("got error %v, want ErrInvalidCredentials", err)
	}

	if token != "" {
		t.Errorf("got token %q, want empty", token)
	}
}

func TestAuthService_Login_JWTGenerationError(t *testing.T) {
	hashed := hashPassword("password123")

	userRepo := &fakeUserRepository{
		findByEmailUser: user.User{
			ID:       1,
			Email:    "john@example.com",
			Password: hashed,
			Role:     "user",
		},
	}

	expectedErr := errors.New("signing JWT failed")
	jwtGen := &fakeJWTGenerator{
		generateTokenErr: expectedErr,
	}

	service := NewService(userRepo, jwtGen)

	token, err := service.Login(context.Background(), LoginRequest{
		Email:    "john@example.com",
		Password: "password123",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("got error %v, want %v", err, expectedErr)
	}

	if token != "" {
		t.Errorf("got token %q, want empty", token)
	}
}
