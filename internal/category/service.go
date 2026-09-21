package category

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type Service interface {
	FindAll(ctx context.Context) ([]Category, error)
	FindByID(ctx context.Context, id int64) (Category, error)
	Create(ctx context.Context, category CreateCategoryRequest) (Category, error)
	Update(ctx context.Context, id int64, category UpdateCategoryRequest) (Category, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo  Repository
	cache redisCache.Cache
}

func NewService(repo Repository, cache redisCache.Cache) *service {
	return &service{repo: repo, cache: cache}
}

func (s *service) FindAll(ctx context.Context) ([]Category, error) {
	cacheKey := "category:all"

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var categories []Category
		if err := json.Unmarshal(val, &categories); err != nil {
			log.Printf("failed to unmarshal categories data from cache: %v", err)
		} else {
			return categories, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch categories data from cache: %v", err)
	}

	categories, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(categories) > 0 {
		categoriesBytes, err := json.Marshal(categories)
		if err != nil {
			log.Printf("failed to marshal categories data: %v", err)
		} else {
			if err := s.cache.Set(ctx, cacheKey, categoriesBytes, 20*time.Minute); err != nil {
				log.Printf("failed to set categories data to cache: %v", err)
			}
		}
	}

	return categories, nil
}

func (s *service) FindByID(ctx context.Context, id int64) (Category, error) {
	cacheKey := fmt.Sprintf("category:%d", id)

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var category Category
		if err := json.Unmarshal(val, &category); err != nil {
			log.Printf("failed to unmarshal category data from cache: %v", err)
		} else {
			return category, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch category data from cache: %v", err)
	}

	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Category{}, err
	}

	categoryBytes, err := json.Marshal(category)
	if err != nil {
		log.Printf("failed to marshal category data: %v", err)
	} else {
		if err := s.cache.Set(ctx, cacheKey, categoryBytes, 30*time.Minute); err != nil {
			log.Printf("failed to set category data to cache: %v", err)
		}
	}

	return category, nil
}

func (s *service) Create(ctx context.Context, category CreateCategoryRequest) (Category, error) {
	categoryRequest := Category{
		Name: category.Name,
	}

	newCategory, err := s.repo.Create(ctx, categoryRequest)
	if err != nil {
		return Category{}, err
	}

	if err := s.cache.Delete(ctx, "category:all"); err != nil {
		log.Printf("failed to invalidate category cache: %v", err)
	}

	return newCategory, nil
}

func (s *service) Update(ctx context.Context, id int64, category UpdateCategoryRequest) (Category, error) {
	categoryRequest := Category{
		Name: category.Name,
	}

	updatedCategory, err := s.repo.Update(ctx, id, categoryRequest)
	if err != nil {
		return Category{}, err
	}

	cacheKey := fmt.Sprintf("category:%d", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("failed to invalidate cache %s: %v", cacheKey, err)
	}

	return updatedCategory, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("category:%d", id)
	if err := s.cache.Delete(ctx, cacheKey, "category:all"); err != nil {
		log.Printf("failed to invalidate caches %s and category:all: %v", cacheKey, err)
	}

	return nil
}
