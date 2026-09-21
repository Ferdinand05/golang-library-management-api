package member

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/database"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]Member, error)
	FindByID(ctx context.Context, id int64) (Member, error)
	Create(ctx context.Context, member Member) (Member, error)
	Update(ctx context.Context, id int64, member Member) (Member, error)
	Delete(ctx context.Context, id int64) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db: db}
}

func (r *repository) FindAll(ctx context.Context) ([]Member, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	var members []Member

	err := db.Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("finding members:%w", err)
	}

	return members, nil
}

func (r *repository) FindByID(ctx context.Context, id int64) (Member, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	var member Member
	err := db.First(&member, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Member{}, ErrMemberNotFound
		}

		return Member{}, fmt.Errorf("finding member :%w", err)
	}

	return member, nil
}

func (r *repository) Create(ctx context.Context, member Member) (Member, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	err := db.Create(&member).Error
	if err != nil {
		return Member{}, fmt.Errorf("creating member:%w", err)
	}

	return member, nil
}

func (r *repository) Update(ctx context.Context, id int64, member Member) (Member, error) {

	db := database.GetDB(ctx, r.db).WithContext(ctx)

	result := db.Model(&Member{}).
		Where("id = ?", id).
		Updates(member)

	if result.Error != nil {
		return Member{}, fmt.Errorf("updating member:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return Member{}, ErrMemberNotFound
	}

	return member, nil

}

func (r *repository) Delete(ctx context.Context, id int64) error {

	db := database.GetDB(ctx, r.db).WithContext(ctx)
	result := db.Delete(&Member{}, id)

	if result.Error != nil {
		return fmt.Errorf("deleting member:%w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}

	return nil

}
