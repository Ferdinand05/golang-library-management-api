package book

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
	redisCache "ferdinand/library-management-system-api/internal/redis"
)

type Service interface {
	FindAll(ctx context.Context) ([]Book, error)
	FindByID(ctx context.Context, id int64) (Book, error)
	FindByAuthorID(ctx context.Context, authorID int64) ([]Book, error)
	FindByCategoryID(ctx context.Context, categoryID int64) ([]Book, error)

	Create(ctx context.Context, book CreateBookRequest) (Book, error)
	Update(ctx context.Context, id int64, book UpdateBookRequest) (Book, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo         Repository
	authorRepo   author.Repository
	categoryRepo category.Repository
	cache        redisCache.Cache
}

func NewService(repo Repository, authorRepo author.Repository, categoryRepo category.Repository, cache redisCache.Cache) *service {
	return &service{
		repo:         repo,
		authorRepo:   authorRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
	}
}

func (s *service) FindAll(ctx context.Context) ([]Book, error) {
	cacheKey := "book:all"

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var books []Book
		if err := json.Unmarshal(val, &books); err != nil {
			log.Printf("failed to unmarshal books data from cache: %v", err)
		} else {
			return books, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch books data from cache: %v", err)
	}

	books, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if len(books) > 0 {
		booksBytes, err := json.Marshal(books)
		if err != nil {
			log.Printf("failed to marshal books data: %v", err)
		} else {
			if err := s.cache.Set(ctx, cacheKey, booksBytes, 5*time.Minute); err != nil {
				log.Printf("failed to set books data to cache: %v", err)
			}
		}
	}

	return books, nil
}

func (s *service) FindByID(ctx context.Context, id int64) (Book, error) {
	cacheKey := fmt.Sprintf("book:%d", id)

	val, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		var book Book
		if err := json.Unmarshal(val, &book); err != nil {
			log.Printf("failed to unmarshal book data from cache: %v", err)
		} else {
			return book, nil
		}
	} else if err != redisCache.ErrCacheMiss {
		log.Printf("failed to fetch book data from cache: %v", err)
	}

	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return Book{}, err
	}

	bookBytes, err := json.Marshal(book)
	if err != nil {
		log.Printf("failed to marshal book data: %v", err)
	} else {
		if err := s.cache.Set(ctx, cacheKey, bookBytes, 10*time.Minute); err != nil {
			log.Printf("failed to set book data to cache: %v", err)
		}
	}

	return book, nil
}

func (s *service) FindByAuthorID(ctx context.Context, authorID int64) ([]Book, error) {
	_, err := s.authorRepo.FindByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, author.ErrAuthorNotFound) {
			return nil, author.ErrAuthorNotFound
		}

		return nil, err
	}

	return s.repo.FindByAuthorID(ctx, authorID)
}

func (s *service) FindByCategoryID(ctx context.Context, categoryID int64) ([]Book, error) {
	_, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return nil, category.ErrCategoryNotFound
		}

		return nil, err
	}

	return s.repo.FindByCategoryID(ctx, categoryID)
}

func (s *service) Create(ctx context.Context, book CreateBookRequest) (Book, error) {
	bookRequest := Book{
		ISBN:        book.ISBN,
		Title:       book.Title,
		Description: book.Description,
		Stock:       book.Stock,
		AuthorID:    book.AuthorID,
		CategoryID:  book.CategoryID,
	}

	_, err := s.authorRepo.FindByID(ctx, bookRequest.AuthorID)
	if err != nil {
		if errors.Is(err, author.ErrAuthorNotFound) {
			return Book{}, author.ErrAuthorNotFound
		}

		return Book{}, err
	}

	_, err = s.categoryRepo.FindByID(ctx, bookRequest.CategoryID)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return Book{}, category.ErrCategoryNotFound
		}

		return Book{}, err
	}

	newBook, err := s.repo.Create(ctx, bookRequest)
	if err != nil {
		return Book{}, err
	}

	if err := s.cache.Delete(ctx, "book:all"); err != nil {
		log.Printf("failed to invalidate book cache: %v", err)
	}

	return newBook, nil
}

func (s *service) Update(ctx context.Context, id int64, book UpdateBookRequest) (Book, error) {
	bookRequest := Book{
		Title:       book.Title,
		Description: book.Description,
		AuthorID:    book.AuthorID,
		CategoryID:  book.CategoryID,
		Stock:       book.Stock,
	}

	_, err := s.authorRepo.FindByID(ctx, bookRequest.AuthorID)
	if err != nil {
		if errors.Is(err, author.ErrAuthorNotFound) {
			return Book{}, author.ErrAuthorNotFound
		}

		return Book{}, err
	}

	_, err = s.categoryRepo.FindByID(ctx, bookRequest.CategoryID)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return Book{}, category.ErrCategoryNotFound
		}

		return Book{}, err
	}

	updatedBook, err := s.repo.Update(ctx, id, bookRequest)
	if err != nil {
		return Book{}, err
	}

	cacheKey := fmt.Sprintf("book:%d", id)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		log.Printf("failed to invalidate cache %s: %v", cacheKey, err)
	}

	return updatedBook, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("book:%d", id)
	if err := s.cache.Delete(ctx, cacheKey, "book:all"); err != nil {
		log.Printf("failed to invalidate caches %s and book:all: %v", cacheKey, err)
	}

	return nil
}