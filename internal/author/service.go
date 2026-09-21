package author

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type Service interface {
	FindAll(ctx context.Context) ([]Author, error)
	FindByID(ctx context.Context, id int64) (Author, error)
	Create(ctx context.Context, author CreateAuthorRequest) (Author, error)
	Update(ctx context.Context, id int64, author UpdateAuthorRequest) (Author, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo  Repository
	cache redisCache.Cache
}

func NewService(repo Repository, cache redisCache.Cache) *service {
	return &service{repo: repo, cache: cache}
}

func (s *service) FindAll(ctx context.Context) ([]Author, error) {
	cacheKey := "author:all"

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var authors []Author
		if err := json.Unmarshal(val, &authors); err != nil {
			log.Printf("failed to unmarshal authors data from cache: %v", err)
		} else {
			return authors, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch authors data from cache: %v", err)
	}

	authors, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(authors) > 0 {
		authorsBytes, err := json.Marshal(authors)
		if err != nil {
			log.Printf("failed to marshal authors data: %v", err)
		} else {
			if err := s.cache.Set(ctx, cacheKey, authorsBytes, 20*time.Minute); err != nil {
				log.Printf("failed to set authors data to cache: %v", err)
			}
		}
	}

	return authors, nil
}

func (s *service) FindByID(ctx context.Context, id int64) (Author, error) {
	cacheKey := fmt.Sprintf("author:%d", id)

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var author Author
		if err := json.Unmarshal(val, &author); err != nil {
			log.Printf("failed to unmarshal author data from cache: %v", err)
		} else {
			return author, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch author data from cache: %v", err)
	}

	author, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Author{}, err
	}

	authorBytes, err := json.Marshal(author)
	if err != nil {
		log.Printf("failed to marshal author data: %v", err)
	} else {
		if err := s.cache.Set(ctx, cacheKey, authorBytes, 30*time.Minute); err != nil {
			log.Printf("failed to set author data to cache: %v", err)
		}
	}

	return author, nil
}

func (s *service) Create(ctx context.Context, author CreateAuthorRequest) (Author, error) {
	authorRequest := Author{
		Name: author.Name,
	}

	newAuthor, err := s.repo.Create(ctx, authorRequest)
	if err != nil {
		return Author{}, err
	}

	if err := s.cache.Delete(ctx, "author:all"); err != nil {
		log.Printf("failed to invalidate author cache: %v", err)
	}

	return newAuthor, nil
}

func (s *service) Update(ctx context.Context, id int64, author UpdateAuthorRequest) (Author, error) {
	authorRequest := Author{
		Name: author.Name,
	}

	updatedAuthor, err := s.repo.Update(ctx, id, authorRequest)
	if err != nil {
		return Author{}, err
	}

	cacheKey := fmt.Sprintf("author:%d", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("failed to invalidate cache %s: %v", cacheKey, err)
	}

	return updatedAuthor, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("author:%d", id)
	if err := s.cache.Delete(ctx, cacheKey, "author:all"); err != nil {
		log.Printf("failed to invalidate caches %s and author:all: %v", cacheKey, err)
	}

	return nil
}
