package book

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

func createTestBook(
	t *testing.T,
	db *gorm.DB,
	isbn string,
	title string,
	authorID int64,
	categoryID int64,
	stock int,
) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO books (
			isbn,
			title,
			author_id,
			category_id,
			stock
		)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`,
		isbn,
		title,
		authorID,
		categoryID,
		stock,
	).Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}

	return id
}

func TestBookRepository_Create(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	description := "Test Description"
	book := Book{
		ISBN:        "123-456-789",
		Title:       "Test Book",
		Description: &description,
		AuthorID:    authorID,
		CategoryID:  categoryID,
		Stock:       10,
	}

	got, err := repo.Create(context.Background(), book)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.Title != "Test Book" {
		t.Errorf("got title %s, want Test Book", got.Title)
	}

	if got.ISBN != "123-456-789" {
		t.Errorf("got ISBN %s, want 123-456-789", got.ISBN)
	}

	if got.Stock != 10 {
		t.Errorf("got stock %d, want 10", got.Stock)
	}

	if got.Author.ID == 0 {
		t.Error("expected Author to be preloaded")
	}

	if got.Category.ID == 0 {
		t.Error("expected Category to be preloaded")
	}
}

func TestBookRepository_Create_DuplicateISBN(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	book1 := Book{
		ISBN:       "123-456",
		Title:      "Book 1",
		AuthorID:   authorID,
		CategoryID: categoryID,
		Stock:      5,
	}

	_, err := repo.Create(context.Background(), book1)
	if err != nil {
		t.Fatalf("failed to create first book: %v", err)
	}

	book2 := Book{
		ISBN:       "123-456",
		Title:      "Book 2",
		AuthorID:   authorID,
		CategoryID: categoryID,
		Stock:      5,
	}

	_, err = repo.Create(context.Background(), book2)

	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestBookRepository_Create_InvalidAuthorID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	categoryID := createTestCategory(t, db, "Test Category")

	book := Book{
		ISBN:       "123-456",
		Title:      "Test Book",
		AuthorID:   999,
		CategoryID: categoryID,
		Stock:      5,
	}

	_, err := repo.Create(context.Background(), book)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}

func TestBookRepository_Create_InvalidCategoryID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")

	book := Book{
		ISBN:       "123-456",
		Title:      "Test Book",
		AuthorID:   authorID,
		CategoryID: 999,
		Stock:      5,
	}

	_, err := repo.Create(context.Background(), book)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}

func TestBookRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	createTestBook(t, db, "111-111", "Book 1", authorID, categoryID, 5)
	createTestBook(t, db, "222-222", "Book 2", authorID, categoryID, 3)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d books, want 2", len(got))
	}

	if got[0].Author.ID == 0 {
		t.Error("expected Author to be preloaded")
	}

	if got[0].Category.ID == 0 {
		t.Error("expected Category to be preloaded")
	}
}

func TestBookRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d books, want 0", len(got))
	}
}

func TestBookRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(
		t,
		db,
		"123-456",
		"Test Book",
		authorID,
		categoryID,
		5,
	)
	got, err := repo.FindByID(context.Background(), bookID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != bookID {
		t.Errorf("got ID %d, want %d", got.ID, bookID)
	}

	if got.Title != "Test Book" {
		t.Errorf("got title %s, want Test Book", got.Title)
	}

	if got.Author.ID == 0 {
		t.Error("expected Author to be preloaded")
	}

	if got.Category.ID == 0 {
		t.Error("expected Category to be preloaded")
	}
}

func TestBookRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_FindByAuthorID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author1ID := createTestAuthor(t, db, "Author 1")
	author2ID := createTestAuthor(t, db, "Author 2")
	categoryID := createTestCategory(t, db, "Test Category")

	createTestBook(t, db, "111-111", "Book 1", author1ID, categoryID, 5)
	createTestBook(t, db, "222-222", "Book 2", author1ID, categoryID, 3)
	createTestBook(t, db, "333-333", "Book 3", author2ID, categoryID, 2)

	got, err := repo.FindByAuthorID(context.Background(), author1ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d books, want 2", len(got))
	}
}

func TestBookRepository_FindByAuthorID_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Author With No Books")

	got, err := repo.FindByAuthorID(context.Background(), authorID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d books, want 0", len(got))
	}
}

func TestBookRepository_FindByCategoryID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	category1ID := createTestCategory(t, db, "Category 1")
	category2ID := createTestCategory(t, db, "Category 2")

	createTestBook(t, db, "111-111", "Book 1", authorID, category1ID, 5)
	createTestBook(t, db, "222-222", "Book 2", authorID, category1ID, 3)
	createTestBook(t, db, "333-333", "Book 3", authorID, category2ID, 2)

	got, err := repo.FindByCategoryID(context.Background(), category1ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d books, want 2", len(got))
	}
}

func TestBookRepository_FindByCategoryID_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	categoryID := createTestCategory(t, db, "Category With No Books")

	got, err := repo.FindByCategoryID(context.Background(), categoryID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d books, want 0", len(got))
	}
}

func TestBookRepository_Update(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(
		t,
		db,
		"123-456",
		"Old Title",
		authorID,
		categoryID,
		5,
	)

	updated, err := repo.Update(context.Background(), bookID, Book{
		Title:      "New Title",
		Stock:      10,
		AuthorID:   authorID,
		CategoryID: categoryID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Title != "New Title" {
		t.Errorf("got title %s, want New Title", updated.Title)
	}

	if updated.Stock != 10 {
		t.Errorf("got stock %d, want 10", updated.Stock)
	}
}

func TestBookRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	_, err := repo.Update(context.Background(), 999, Book{
		Title:      "New Title",
		AuthorID:   authorID,
		CategoryID: categoryID,
		Stock:      10,
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_Delete(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(t, db, "123-456", "To Delete", authorID, categoryID, 5)

	err := repo.Delete(context.Background(), bookID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.FindByID(context.Background(), bookID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.Delete(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func TestBookRepository_DecreaseStock(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(t, db, "123-456", "Test Book", authorID, categoryID, 5)

	err := repo.DecreaseStock(context.Background(), bookID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	book, err := repo.FindByID(context.Background(), bookID)
	if err != nil {
		t.Fatalf("failed to find book: %v", err)
	}

	if book.Stock != 4 {
		t.Errorf("got stock %d, want 4", book.Stock)
	}
}

func TestBookRepository_DecreaseStock_OutOfStock(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(t, db, "123-456", "Test Book", authorID, categoryID, 0)

	err := repo.DecreaseStock(context.Background(), bookID)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookOutOfStock {
		t.Errorf("expected ErrBookOutOfStock, got %v", err)
	}
}

func TestBookRepository_DecreaseStock_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.DecreaseStock(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookOutOfStock {
		t.Errorf("expected ErrBookOutOfStock, got %v", err)
	}
}

func TestBookRepository_IncreaseStock(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(t, db, "123-456", "Test Book", authorID, categoryID, 5)

	err := repo.IncreaseStock(context.Background(), bookID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	book, err := repo.FindByID(context.Background(), bookID)
	if err != nil {
		t.Fatalf("failed to find book: %v", err)
	}

	if book.Stock != 6 {
		t.Errorf("got stock %d, want 6", book.Stock)
	}
}

func TestBookRepository_IncreaseStock_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.IncreaseStock(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBookNotFound {
		t.Errorf("expected ErrBookNotFound, got %v", err)
	}
}

func createTestMember(
	t *testing.T,
	db *gorm.DB,
	memberCode string,
	name string,
	email string,
) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO members (
			member_code,
			name,
			email
		)
		VALUES (?, ?, ?)
		RETURNING id
	`,
		memberCode,
		name,
		email,
	).Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test member: %v", err)
	}

	return id
}

func TestBookRepository_Delete_WithBorrowings_ForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	authorID := createTestAuthor(t, db, "Test Author")
	categoryID := createTestCategory(t, db, "Test Category")

	bookID := createTestBook(
		t,
		db,
		"123-456",
		"Book With Borrowings",
		authorID,
		categoryID,
		5,
	)

	memberID := createTestMember(
		t,
		db,
		"ABC1234",
		"Test Member",
		"test@example.com",
	)

	err := db.Exec(`
	INSERT INTO borrowings (
		member_id,
		book_id,
		borrowed_at,
		due_at
	)
	VALUES (?, ?, NOW(), NOW() + INTERVAL '7 days')
`,
		memberID, bookID,
	).Error

	if err != nil {
		t.Fatalf("failed to create borrowing: %v", err)
	}

	err = repo.Delete(context.Background(), bookID)
	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}
