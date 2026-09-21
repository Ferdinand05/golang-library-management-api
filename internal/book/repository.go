package book

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/database"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Book, error)
	FindByID(ctx context.Context, id int64) (Book, error)
	FindByAuthorID(ctx context.Context, authorID int64) ([]Book, error)
	FindByCategoryID(ctx context.Context, categoryID int64) ([]Book, error)

	Create(ctx context.Context, book Book) (Book, error)
	Update(ctx context.Context, id int64, book Book) (Book, error)
	Delete(ctx context.Context, id int64) error

	DecreaseStock(ctx context.Context, id int64) error
	IncreaseStock(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Book, error) {
	var books []Book

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Preload("Category").Preload("Author").Find(&books).Error
	if err != nil {
		return nil, fmt.Errorf("finding books :%w", err)
	}

	return books, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (Book, error) {
	var book Book

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.Preload("Category").
		Preload("Author").
		First(&book, id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Book{}, ErrBookNotFound
		}

		return Book{}, fmt.Errorf("finding book:%w", err)

	}

	return book, nil
}

func (r *repository) FindByAuthorID(ctx context.Context, authorID int64) ([]Book, error) {
	var books []Book

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Where("author_id = ?", authorID).
		Preload("Category").
		Preload("Author").
		Find(&books).
		Error

	if err != nil {
		return nil, fmt.Errorf("finding books by author:%w", err)
	}

	return books, nil
}

func (r *repository) FindByCategoryID(ctx context.Context, categoryID int64) ([]Book, error) {
	var books []Book

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Where("category_id = ?", categoryID).
		Preload("Category").
		Preload("Author").
		Find(&books).
		Error

	if err != nil {

		return nil, fmt.Errorf("finding books by category:%w", err)
	}

	return books, nil
}

func (r *repository) Create(ctx context.Context, book Book) (Book, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Create(&book).Error
	if err != nil {
		return Book{}, fmt.Errorf("inserting book:%w", err)
	}

	var createdBook Book
	err = db.Preload("Author").
		Preload("Category").
		First(&createdBook, book.ID).
		Error

	if err != nil {
		return Book{}, fmt.Errorf("finding created book: %w", err)
	}

	return createdBook, nil
}

func (r *repository) Update(ctx context.Context, id int64, book Book) (Book, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.
		Model(&Book{}).
		Where("id = ?", id).
		Updates(book)

	if result.Error != nil {
		return Book{}, fmt.Errorf("updating book:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Book{}, ErrBookNotFound
	}

	var updatedBook Book
	err := db.Preload("Author").
		Preload("Category").
		First(&updatedBook, id).
		Error

	if err != nil {
		return Book{}, fmt.Errorf("finding updated book: %w", err)
	}

	return updatedBook, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.
		Delete(&Book{}, id)

	if result.Error != nil {
		return fmt.Errorf("deleting book:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrBookNotFound
	}

	return nil
}

func (r *repository) DecreaseStock(ctx context.Context, id int64) error {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.Model(&Book{}).
		Where("id = ? AND stock > 0", id).
		UpdateColumn("stock", gorm.Expr("stock - ?", 1))

	if result.Error != nil {
		return fmt.Errorf("decreasing book stock:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrBookOutOfStock
	}

	return nil

}

func (r *repository) IncreaseStock(ctx context.Context, id int64) error {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.Model(&Book{}).
		Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("increasing book stock:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrBookNotFound
	}

	return nil

}
