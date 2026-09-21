package borrowing

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/database"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Borrowing, error)
	FindByID(ctx context.Context, id int64) (Borrowing, error)
	FindByMemberID(ctx context.Context, member_id int64) ([]Borrowing, error)
	Create(ctx context.Context, borrowing Borrowing) (Borrowing, error)
	Return(ctx context.Context, borrowingID int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Borrowing, error) {
	var borrowings []Borrowing

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Preload("Member").
		Preload("Book").
		Find(&borrowings).Error

	if err != nil {
		return nil, fmt.Errorf("finding borrowings:%w", err)
	}

	return borrowings, nil

}

func (r *repository) FindByID(ctx context.Context, id int64) (Borrowing, error) {
	var borrowing Borrowing

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Preload("Member").
		Preload("Book").
		First(&borrowing, id).Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Borrowing{}, ErrBorrowingNotFound
		}

		return Borrowing{}, fmt.Errorf("finding borrowing:%w", err)
	}

	return borrowing, nil

}

func (r *repository) FindByMemberID(ctx context.Context, member_id int64) ([]Borrowing, error) {
	var borrowings []Borrowing

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Preload("Member").
		Preload("Book").
		Where("member_id = ?", member_id).
		Find(&borrowings).Error

	if err != nil {
		return nil, fmt.Errorf("finding borrowings by member id:%w", err)
	}

	return borrowings, nil

}

func (r *repository) Create(ctx context.Context, borrowing Borrowing) (Borrowing, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.
		Create(&borrowing).
		Error

	if err != nil {
		return Borrowing{}, fmt.Errorf("inserting borrowing:%w", err)
	}

	var createdBorrowing Borrowing
	err = db.
		Preload("Book").
		Preload("Member").
		First(&createdBorrowing, borrowing.ID).
		Error

	if err != nil {
		return Borrowing{}, fmt.Errorf("finding created borrowing:%w", err)
	}

	return createdBorrowing, nil
}

func (r *repository) Return(ctx context.Context, borrowingID int64) error {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.Model(&Borrowing{}).
		Where("id = ? AND returned_at IS NULL", borrowingID).
		UpdateColumn("returned_at", time.Now())

	if result.Error != nil {
		return fmt.Errorf("returning book: %w", result.Error)
	}

	var borrowing Borrowing
	if result.RowsAffected == 0 {
		err := db.First(&borrowing, borrowingID).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrBorrowingNotFound
		}

		if err != nil {
			return fmt.Errorf("finding borrowing after failed return: %w", err)
		}

		if borrowing.ReturnedAt != nil {
			return ErrBorrowingAlreadyReturned
		}

		return fmt.Errorf("returning borrowing: no rows affected")
	}

	return nil
}
