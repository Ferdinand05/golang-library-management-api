package user

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"ferdinand/library-management-system-api/internal/crypto"
	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type Service interface {
	FindAll(ctx context.Context) ([]User, error)
	FindByID(ctx context.Context, id int64) (User, error)
	Create(ctx context.Context, request CreateUserRequest) (User, error)
}

type service struct {
	repo  Repository
	cache redisCache.Cache
}

func NewService(repo Repository, cache redisCache.Cache) *service {
	return &service{repo: repo, cache: cache}
}

func (s *service) FindAll(ctx context.Context) ([]User, error) {
	cacheKey := "user:all"

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var users []User
		if err := json.Unmarshal(val, &users); err != nil {
			log.Printf("failed to unmarshal users data from cache: %v", err)
		} else {
			return users, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch users data from cache: %v", err)
	}

	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(users) > 0 {
		usersBytes, err := json.Marshal(users)
		if err != nil {
			log.Printf("failed to marshal users data: %v", err)
		} else {
			if err := s.cache.Set(ctx, cacheKey, usersBytes, 5*time.Minute); err != nil {
				log.Printf("failed to set users data to cache: %v", err)
			}
		}
	}

	return users, nil
}

func (s *service) FindByID(ctx context.Context, id int64) (User, error) {
	cacheKey := fmt.Sprintf("user:%d", id)

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var user User
		if err := json.Unmarshal(val, &user); err != nil {
			log.Printf("failed to unmarshal user data from cache: %v", err)
		} else {
			return user, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch user data from cache: %v", err)
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return User{}, err
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Printf("failed to marshal user data: %v", err)
	} else {
		if err := s.cache.Set(ctx, cacheKey, userBytes, 5*time.Minute); err != nil {
			log.Printf("failed to set user data to cache: %v", err)
		}
	}

	return user, nil
}

func (s *service) Create(ctx context.Context, request CreateUserRequest) (User, error) {
	hashedPassword, err := crypto.HashPassword(request.Password)
	if err != nil {
		return User{}, err
	}

	userRequest := User{
		Email:    request.Email,
		Password: hashedPassword,
		Role:     request.Role,
	}

	createdUser, err := s.repo.Create(ctx, userRequest)
	if err != nil {
		return User{}, err
	}

	if err := s.cache.Delete(ctx, "user:all"); err != nil {
		log.Printf("failed to invalidate user cache: %v", err)
	}

	return createdUser, nil
}
