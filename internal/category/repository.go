package category

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id int64) (Category, error)
	Create(ctx context.Context, category Category) (Category, error)
	Update(ctx context.Context, id int64, category Category) (Category, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Category, error) {
	var categories []Category

	err := r.db.WithContext(ctx).Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("finding categories:%w", err)
	}

	return categories, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (Category, error) {
	var category Category

	err := r.db.WithContext(ctx).First(&category, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Category{}, ErrCategoryNotFound
		}

		return Category{}, fmt.Errorf("finding category:%w", err)
	}

	return category, nil
}

func (r *repository) Create(ctx context.Context, category Category) (Category, error) {
	err := r.db.WithContext(ctx).Create(&category).Error
	if err != nil {
		return Category{}, fmt.Errorf("inserting category:%w", err)
	}

	return category, nil
}

func (r *repository) Update(ctx context.Context, id int64, category Category) (Category, error) {
	result := r.db.WithContext(ctx).Model(&Category{}).
		Where("id = ?", id).
		Updates(category)

	if result.Error != nil {
		return Category{}, fmt.Errorf("updating category:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Category{}, ErrCategoryNotFound
	}

	category.ID = id
	return category, nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&Category{}, id)

	if result.Error != nil {
		return fmt.Errorf("deleting category:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrCategoryNotFound
	}

	return nil
}