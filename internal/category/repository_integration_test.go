package category

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

func createTestCategory(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO categories (name)
		VALUES (?)
		RETURNING id
	`, name).Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	return id
}

func createTestAuthor(t *testing.T, db *gorm.DB, name string) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO authors (name)
		VALUES (?)
		RETURNING id
	`, name).Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test author: %v", err)
	}

	return id
}

func TestCategoryRepository_Create(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	category := Category{
		Name: "Fiction",
	}

	got, err := repo.Create(context.Background(), category)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.Name != "Fiction" {
		t.Errorf("got name %s, want Fiction", got.Name)
	}

	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be populated")
	}

	if got.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be populated")
	}

	saved, err := repo.FindByID(context.Background(), got.ID)
	if err != nil {
		t.Fatalf("failed to find created category: %v", err)
	}

	if saved.Name != "Fiction" {
		t.Errorf("got name %s, want Fiction", saved.Name)
	}
}

func TestCategoryRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	_, err := repo.Create(context.Background(), Category{Name: "Fiction"})
	if err != nil {
		t.Fatalf("failed to create Fiction category: %v", err)
	}

	_, err = repo.Create(context.Background(), Category{Name: "Non-Fiction"})
	if err != nil {
		t.Fatalf("failed to create Non-Fiction category: %v", err)
	}

	_, err = repo.Create(context.Background(), Category{Name: "Science"})
	if err != nil {
		t.Fatalf("failed to create Science category: %v", err)
	}

	got, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 3 {
		t.Errorf("got %d categories, want 3", len(got))
	}
}

func TestCategoryRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d categories, want 0", len(got))
	}
}

func TestCategoryRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Category{Name: "History"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Name != "History" {
		t.Errorf("got name %s, want History", got.Name)
	}
}

func TestCategoryRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryRepository_Update(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Category{Name: "Old Name"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	updated, err := repo.Update(context.Background(), created.ID, Category{Name: "New Name"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("got name %s, want New Name", updated.Name)
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated category: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestCategoryRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.Update(context.Background(), 999, Category{Name: "New Name"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryRepository_Delete(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Category{Name: "To Delete"})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	err = repo.Delete(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.Delete(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrCategoryNotFound {
		t.Errorf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestCategoryRepository_Delete_WithBooks_ForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	categoryID := createTestCategory(t, db, "Category With Books")
	authorID := createTestAuthor(t, db, "Test Author")

	err := db.Exec(`
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
		authorID,
		categoryID,
		5,
	).Error

	if err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}

	err = repo.Delete(context.Background(), categoryID)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}
