package member

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

func createTestAuthor(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO authors (name)
		VALUES (?)
		RETURNING id
	`, "Test Author").Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test author: %v", err)
	}

	return id
}

func createTestCategory(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	var id int64

	err := db.Raw(`
		INSERT INTO categories (name)
		VALUES (?)
		RETURNING id
	`, "Test Category").Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test category: %v", err)
	}

	return id
}

func createTestBook(
	t *testing.T,
	db *gorm.DB,
	authorID int64,
	categoryID int64,
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
		"123-456",
		"Test Book",
		authorID,
		categoryID,
		5,
	).Scan(&id).Error

	if err != nil {
		t.Fatalf("failed to create test book: %v", err)
	}

	return id
}

func TestMemberRepositoryCreate(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	member := Member{
		MemberCode: "ABC1234",
		Name:       "John",
		Email:      "john@example.com",
		Phone:      "08123456789",
	}

	got, err := repo.Create(context.Background(), member)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID == 0 {
		t.Error("expected ID to be generated, got 0")
	}

	if got.MemberCode != "ABC1234" {
		t.Errorf("got member code %s, want ABC1234", got.MemberCode)
	}

	if got.Name != "John" {
		t.Errorf("got name %s, want John", got.Name)
	}

	saved, err := repo.FindByID(
		context.Background(),
		got.ID,
	)

	if err != nil {
		t.Fatalf("failed to find created member: %v", err)
	}

	if saved.Name != "John" {
		t.Errorf("got name %s, want John", saved.Name)
	}

}

func TestMemberRepository_FindAll(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Member 1",
		Email:      "member1@example.com",
		Phone:      "08111111111",
	})

	repo.Create(context.Background(), Member{
		MemberCode: "DEF5678",
		Name:       "Member 2",
		Email:      "member2@example.com",
		Phone:      "08222222222",
	})

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Errorf("got %d members, want 2", len(got))
	}
}

func TestMemberRepository_FindAll_Empty(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	got, err := repo.FindAll(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("got %d members, want 0", len(got))
	}
}

func TestMemberRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "John Doe",
		Email:      "john@example.com",
		Phone:      "08123456789",
	})
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	got, err := repo.FindByID(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.ID != created.ID {
		t.Errorf("got ID %d, want %d", got.ID, created.ID)
	}

	if got.Name != "John Doe" {
		t.Errorf("got name %s, want John Doe", got.Name)
	}
}

func TestMemberRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.FindByID(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestMemberRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Member 1",
		Email:      "test@example.com",
		Phone:      "08123456789",
	})
	if err != nil {
		t.Fatalf("failed to create first member: %v", err)
	}

	_, err = repo.Create(context.Background(), Member{
		MemberCode: "DEF5678",
		Name:       "Member 2",
		Email:      "test@example.com",
		Phone:      "08987654321",
	})

	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestMemberRepository_Create_DuplicateMemberCode(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Member 1",
		Email:      "member1@example.com",
		Phone:      "08123456789",
	})
	if err != nil {
		t.Fatalf("failed to create first member: %v", err)
	}

	_, err = repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Member 2",
		Email:      "member2@example.com",
		Phone:      "08987654321",
	})

	if err == nil {
		t.Fatal("expected unique constraint error, got nil")
	}
}

func TestMemberRepository_Update(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "Old Name",
		Email:      "old@example.com",
		Phone:      "08111111111",
	})
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	newAddress := "New Address"
	updated, err := repo.Update(context.Background(), created.ID, Member{
		Name:    "New Name",
		Email:   "new@example.com",
		Phone:   "08222222222",
		Address: &newAddress,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("got name %s, want New Name", updated.Name)
	}

	if updated.Email != "new@example.com" {
		t.Errorf("got email %s, want new@example.com", updated.Email)
	}

	saved, err := repo.FindByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("failed to find updated member: %v", err)
	}

	if saved.Name != "New Name" {
		t.Errorf("got name %s, want New Name", saved.Name)
	}
}

func TestMemberRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	_, err := repo.Update(context.Background(), 999, Member{
		Name:  "New Name",
		Email: "new@example.com",
		Phone: "08123456789",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestMemberRepository_Delete(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	created, err := repo.Create(context.Background(), Member{
		MemberCode: "ABC1234",
		Name:       "To Delete",
		Email:      "delete@example.com",
		Phone:      "08123456789",
	})
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	err = repo.Delete(context.Background(), created.ID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.FindByID(context.Background(), created.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}

	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestMemberRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	err := repo.Delete(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != ErrMemberNotFound {
		t.Errorf("expected ErrMemberNotFound, got %v", err)
	}
}

func TestMemberRepository_Delete_WithBorrowings_ForeignKeyConstraint(t *testing.T) {
	db := setupTestDB(t)

	repo := NewRepository(db)

	member, err := repo.Create(
		context.Background(),
		Member{
			MemberCode: "ABC1234",
			Name:       "Member With Borrowings",
			Email:      "member@example.com",
			Phone:      "08123456789",
		},
	)
	if err != nil {
		t.Fatalf("failed to create member: %v", err)
	}

	authorID := createTestAuthor(t, db)
	categoryID := createTestCategory(t, db)

	bookID := createTestBook(
		t,
		db,
		authorID,
		categoryID,
	)

	err = db.Exec(`
		INSERT INTO borrowings (
			member_id,
			book_id,
			borrowed_at,
			due_at
		)
		VALUES (?, ?, NOW(), NOW() + INTERVAL '7 days')
	`,
		member.ID,
		bookID,
	).Error

	if err != nil {
		t.Fatalf("failed to create borrowing: %v", err)
	}

	err = repo.Delete(
		context.Background(),
		member.ID,
	)

	if err == nil {
		t.Fatal("expected foreign key constraint error, got nil")
	}
}
