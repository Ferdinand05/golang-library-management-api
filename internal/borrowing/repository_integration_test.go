package borrowing

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/category"
	"ferdinand/library-management-system-api/internal/member"

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

func createTestAuthor(db *gorm.DB, name string) author.Author {
	var a author.Author

	db.Raw(`
		INSERT INTO authors (name)
		VALUES (?)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&a)

	return a
}

func createTestCategory(db *gorm.DB, name string) category.Category {
	var c category.Category

	db.Raw(`
		INSERT INTO categories (name)
		VALUES (?)
		RETURNING id, name, created_at, updated_at
	`, name).Scan(&c)

	return c
}

func createTestBook(
	db *gorm.DB,
	isbn string,
	title string,
	authorID int64,
	categoryID int64,
	stock int,
) book.Book {
	var b book.Book

	db.Raw(`
		INSERT INTO books (
			isbn,
			title,
			author_id,
			category_id,
			stock
		)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, isbn, title, author_id, category_id, stock, created_at, updated_at
	`,
		isbn,
		title,
		authorID,
		categoryID,
		stock,
	).Scan(&b)

	return b
}

func createTestMember(
	db *gorm.DB,
	code string,
	name string,
	email string,
) member.Member {
	var m member.Member

	db.Raw(`
		INSERT INTO members (
			member_code,
			name,
			email,
			phone
		)
		VALUES (?, ?, ?, ?)
		RETURNING id, member_code, name, email, phone, created_at, updated_at
	`,
		code,
		name,
		email,
		"08123456789",
	).Scan(&m)

	return m
}

func TestBorrowingRepository_Create(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	borrowing := Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	}

	got, err := repo.Create(context.Background(), borrowing)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.MemberID != testMember.ID {
		t.Errorf("got MemberID %d, want %d", got.MemberID, testMember.ID)
	}

	if got.BookID != testBook.ID {
		t.Errorf("got BookID %d, want %d", got.BookID, testBook.ID)
	}

	if got.ReturnedAt != nil {
		t.Error("expected ReturnedAt to be nil")
	}

	if got.Member.ID == 0 {
		t.Error("expected Member to be preloaded")
	}

	if got.Book.ID == 0 {
		t.Error("expected Book to be preloaded")
	}
}

func TestBorrowingRepository_Create_InvalidMemberID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)

	now := time.Now()
	borrowing := Borrowing{
		MemberID:   999,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	}

	_, err := repo.Create(context.Background(), borrowing)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}

func TestBorrowingRepository_Create_InvalidBookID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	borrowing := Borrowing{
		MemberID:   testMember.ID,
		BookID:     999,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	}

	_, err := repo.Create(context.Background(), borrowing)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}

func TestBorrowingRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	repo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	repo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 14),
	})

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d borrowings, want 2", len(got))
	}

	if got[0].Member.ID == 0 {
		t.Error("expected Member to be preloaded")
	}

	if got[0].Book.ID == 0 {
		t.Error("expected Book to be preloaded")
	}
}

func TestBorrowingRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d borrowings, want 0", len(got))
	}
}

func TestBorrowingRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	created, err := repo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})
	if err != nil {
		t.Fatalf("failed to create borrowing: %v", err)
	}

	got, err := repo.FindByID(context.Background(), int64(created.ID))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Member.ID == 0 {
		t.Error("expected Member to be preloaded")
	}

	if got.Book.ID == 0 {
		t.Error("expected Book to be preloaded")
	}
}

func TestBorrowingRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBorrowingNotFound {
		t.Errorf("expected ErrBorrowingNotFound, got %v", err)
	}
}

func TestBorrowingRepository_FindByMemberID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	member1 := createTestMember(db, "ABC1234", "Member 1", "member1@example.com")
	member2 := createTestMember(db, "DEF5678", "Member 2", "member2@example.com")

	now := time.Now()
	repo.Create(context.Background(), Borrowing{
		MemberID:   member1.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	repo.Create(context.Background(), Borrowing{
		MemberID:   member1.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 14),
	})

	repo.Create(context.Background(), Borrowing{
		MemberID:   member2.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})

	got, err := repo.FindByMemberID(context.Background(), member1.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d borrowings, want 2", len(got))
	}
}

func TestBorrowingRepository_FindByMemberID_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	got, err := repo.FindByMemberID(context.Background(), testMember.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d borrowings, want 0", len(got))
	}
}

func TestBorrowingRepository_Return(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	created, err := repo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})
	if err != nil {
		t.Fatalf("failed to create borrowing: %v", err)
	}

	err = repo.Return(context.Background(), int64(created.ID))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved, err := repo.FindByID(context.Background(), int64(created.ID))
	if err != nil {
		t.Fatalf("failed to find borrowing: %v", err)
	}

	if saved.ReturnedAt == nil {
		t.Error("expected ReturnedAt to be populated")
	}
}

func TestBorrowingRepository_Return_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.Return(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBorrowingNotFound {
		t.Errorf("expected ErrBorrowingNotFound, got %v", err)
	}
}

func TestBorrowingRepository_Return_AlreadyReturned(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	author := createTestAuthor(db, "Test Author")
	category := createTestCategory(db, "Test Category")
	testBook := createTestBook(db, "123-456", "Test Book", author.ID, category.ID, 5)
	testMember := createTestMember(db, "ABC1234", "Test Member", "member@example.com")

	now := time.Now()
	created, err := repo.Create(context.Background(), Borrowing{
		MemberID:   testMember.ID,
		BookID:     testBook.ID,
		BorrowedAt: now,
		DueAt:      now.AddDate(0, 0, 7),
	})
	if err != nil {
		t.Fatalf("failed to create borrowing: %v", err)
	}

	err = repo.Return(context.Background(), int64(created.ID))
	if err != nil {
		t.Fatalf("failed to return first time: %v", err)
	}

	err = repo.Return(context.Background(), int64(created.ID))

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrBorrowingAlreadyReturned {
		t.Errorf("expected ErrBorrowingAlreadyReturned, got %v", err)
	}
}
