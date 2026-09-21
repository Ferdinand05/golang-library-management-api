package author

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

func TestAuthorRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	author := Author{
		Name: "John Doe",
	}

	got, err := repo.Create(context.Background(), author)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.Name != "John Doe" {
		t.Errorf("got name %s, want John Doe", got.Name)
	}

	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}

	if got.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be populated")
	}

	saved, err := repo.FindByID(context.Background(), got.ID)
	if err != nil {
		t.Fatalf("failed to find created author: %v", err)
	}

	if saved.Name != "John Doe" {
		t.Errorf("got name %s, want John Doe", saved.Name)
	}
}

func TestAuthorRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	_, err := repo.Create(context.Background(), Author{Name: "Author 1"})
	if err != nil {
		t.Fatalf("failed to create author 1: %v", err)
	}

	_, err = repo.Create(context.Background(), Author{Name: "Author 2"})
	if err != nil {
		t.Fatalf("failed to create author 2: %v", err)
	}

	_, err = repo.Create(context.Background(), Author{Name: "Author 3"})
	if err != nil {
		t.Fatalf("failed to create author 3: %v", err)
	}

	got, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Errorf("got %d authors, want 3", len(got))
	}
}

func TestAuthorRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d authors, want 0", len(got))
	}
}

func TestAuthorRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Author{Name: "Jane Doe"})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Name != "Jane Doe" {
		t.Errorf("got name %s, want Jane Doe", got.Name)
	}
}

func TestAuthorRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrAuthorNotFound {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorRepository_Update(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Author{Name: "Old Name"})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	updated, err := repo.Update(context.Background(), created.ID, Author{Name: "New Name"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("got name %s, want New Name", updated.Name)
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated author: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestAuthorRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	_, err := repo.Update(context.Background(), 999, Author{Name: "New Name"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrAuthorNotFound {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorRepository_Delete(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Author{Name: "To Delete"})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	err = repo.Delete(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrAuthorNotFound {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)

	t.Cleanup(func() {
		db.Exec("TRUNCATE TABLE books, authors, categories RESTART IDENTITY CASCADE")
	})

	repo := NewRepository(db)

	err := repo.Delete(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrAuthorNotFound {
		t.Errorf("expected ErrAuthorNotFound, got %v", err)
	}
}

func TestAuthorRepository_Delete_WithBooks_ForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	author, err := repo.Create(context.Background(), Author{
		Name: "Author With Books",
	})
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	var categoryID int64

	err = db.Raw(`
		INSERT INTO categories (name)
		VALUES (?)
		RETURNING id
	`, "Fiction").Scan(&categoryID).Error
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	err = db.Exec(`
		INSERT INTO books (
			isbn,
			title,
			author_id,
			category_id,
			stock
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		"123-456",
		"Test Book",
		author.ID,
		categoryID,
		5,
	).Error
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	err = repo.Delete(context.Background(), author.ID)
	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}
