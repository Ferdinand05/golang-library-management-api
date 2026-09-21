package user

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	err := godotenv.Load("../../.env.test")
	if err != nil {
		t.Fatalf("failed to load .env.test: %v", err)
	}

	dbUser := os.Getenv("TEST_DB_USER")
	dbPassword := os.Getenv("TEST_DB_PASSWORD")
	dbHost := os.Getenv("TEST_DB_HOST")
	dbPort := os.Getenv("TEST_DB_PORT")
	dbName := os.Getenv("TEST_DB_NAME")
	dbSSLMode := os.Getenv("TEST_DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
		dbPort,
		dbSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("failed to begin test transaction: %v", tx.Error)
	}

	t.Cleanup(func() {
		tx.Rollback()

		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	})

	return tx
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	user := User{
		Email:    "john@example.com",
		Password: "password123",
		Role:     "user",
	}

	got, err := repo.Create(context.Background(), user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", got.Email)
	}

	if got.Role != "user" {
		t.Errorf("got role %s, want user", got.Role)
	}

	saved, err := repo.FindByID(context.Background(), int64(got.ID))
	if err != nil {
		t.Fatalf("failed to find created user: %v", err)
	}

	if saved.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", saved.Email)
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.Create(context.Background(), User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	_, err = repo.Create(context.Background(), User{
		Email:    "test@example.com",
		Password: "password456",
		Role:     "admin",
	})

	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), User{
		Email:    "john@example.com",
		Password: "password123",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	got, err := repo.FindByID(context.Background(), int64(created.ID))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", got.Email)
	}
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), User{
		Email:    "john@example.com",
		Password: "password123",
		Role:     "user",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	got, err := repo.FindByEmail(context.Background(), "john@example.com")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Email != "john@example.com" {
		t.Errorf("got email %s, want john@example.com", got.Email)
	}
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByEmail(context.Background(), "nonexistent@example.com")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	repo.Create(context.Background(), User{
		Email:    "user1@example.com",
		Password: "password123",
		Role:     "user",
	})

	repo.Create(context.Background(), User{
		Email:    "user2@example.com",
		Password: "password456",
		Role:     "admin",
	})

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d users, want 2", len(got))
	}
}

func TestUserRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d users, want 0", len(got))
	}
}

