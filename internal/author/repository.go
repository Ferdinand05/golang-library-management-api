package author

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Author, error)
	FindByID(ctx context.Context, id int64) (Author, error)
	Create(ctx context.Context, author Author) (Author, error)
	Update(ctx context.Context, id int64, author Author) (Author, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Author, error) {
	var authors []Author

	err := r.db.WithContext(ctx).Find(&authors).Error
	if err != nil {
		return nil, fmt.Errorf("finding authors:%w", err)
	}

	return authors, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (Author, error) {
	var author Author

	err := r.db.WithContext(ctx).First(&author, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Author{}, ErrAuthorNotFound
		}

		return Author{}, fmt.Errorf("finding author:%w", err)
	}

	return author, nil
}

func (r *repository) Create(ctx context.Context, author Author) (Author, error) {
	err := r.db.WithContext(ctx).Create(&author).Error
	if err != nil {
		return Author{}, fmt.Errorf("inserting author:%w", err)
	}

	return author, nil
}

func (r *repository) Update(ctx context.Context, id int64, author Author) (Author, error) {
	result := r.db.WithContext(ctx).Model(&Author{}).
		Where("id = ?", id).
		Updates(author)

	if result.Error != nil {
		return Author{}, fmt.Errorf("updating author:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Author{}, ErrAuthorNotFound
	}

	author.ID = id
	return author, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&Author{}, id)

	if result.Error != nil {
		return fmt.Errorf("deleting author:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrAuthorNotFound
	}

	return nil
}
