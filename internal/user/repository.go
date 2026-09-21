package user

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Repository interface {
	FindAll(ctx context.Context) ([]User,error)
	FindByID(ctx context.Context, id int64) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, user User) (User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db}
}

func (r *repository) FindAll(ctx context.Context) ([]User,error) {

	var users []User

	err := r.db.WithContext(ctx).Find(&users).Error

	if err != nil {
		return nil,fmt.Errorf("finding users: %w",err)
	}

	return users,nil
}

func (r *repository) FindByID(ctx context.Context,id int64) (User,error) {

	var user User

	err := r.db.WithContext(ctx).First(&user,id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{},fmt.Errorf("finding user: %w",err)
	}

	return user,nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (User, error) {

	var user User

	err := r.db.WithContext(ctx).Where("email = ?",email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, ErrUserNotFound
		}
		return User{},fmt.Errorf("finding user by email: %w",err)
	}

	return user,nil

}

func (r *repository) Create(ctx context.Context, user User) (User, error) {

	err := r.db.WithContext(ctx).Create(&user).Error
	if err != nil {
		return User{},fmt.Errorf("creating user:%w",err)
	}

	var createdUser User
	err = r.db.WithContext(ctx).First(&createdUser,user.ID).Error
	if err != nil {
		return User{},fmt.Errorf("finding created user:%w",err)
	}

	return createdUser,nil
	
}