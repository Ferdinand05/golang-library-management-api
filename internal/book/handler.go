package book

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/category"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetBooks(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	books, err := h.service.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	var bookResponses = make([]BookResponse, 0, len(books))

	for _, book := range books {
		bookResponses = append(bookResponses, toBookResponse(book))
	}

	c.JSON(http.StatusOK, gin.H{
		"books": bookResponses,
	})

}

func (h *Handler) GetBook(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	book, err := h.service.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "book not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return

	}

	c.JSON(http.StatusOK, gin.H{
		"book": toBookResponse(book),
	})

}
func (h *Handler) CreateBook(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var bookRequest CreateBookRequest

	if err := c.ShouldBindJSON(&bookRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	book, err := h.service.Create(ctx, bookRequest)
	if err != nil {

		if errors.Is(err, author.ErrAuthorNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "author not found",
			})
			return
		}

		if errors.Is(err, category.ErrCategoryNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"book": toBookResponse(book),
	})

}
func (h *Handler) UpdateBook(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	var bookUpdateRequest UpdateBookRequest
	if err := c.ShouldBindJSON(&bookUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	updatedBook, err := h.service.Update(ctx, id, bookUpdateRequest)
	if err != nil {

		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "book not found",
			})
			return
		}

		if errors.Is(err, author.ErrAuthorNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "author not found",
			})
			return
		}

		if errors.Is(err, category.ErrCategoryNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"book": toBookResponse(updatedBook),
	})

}
func (h *Handler) DeleteBook(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {

		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "book not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "data deleted successfully",
	})

}

func (h *Handler) GetBooksByAuthor(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	books, err := h.service.FindByAuthorID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	bookResponses := make([]BookResponse, 0, len(books))
	for _, response := range books {
		bookResponses = append(bookResponses, toBookResponse(response))
	}

	c.JSON(http.StatusOK, gin.H{
		"books": bookResponses,
	})

}

func (h *Handler) GetBooksByCategory(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid book id",
		})
		return
	}

	books, err := h.service.FindByCategoryID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	bookResponses := make([]BookResponse, 0, len(books))
	for _, response := range books {
		bookResponses = append(bookResponses, toBookResponse(response))
	}

	c.JSON(http.StatusOK, gin.H{
		"books": bookResponses,
	})
}

func toBookResponse(book Book) BookResponse {
	return BookResponse{
		ID:          book.ID,
		Title:       book.Title,
		ISBN:        book.ISBN,
		Description: book.Description,
		Stock:       book.Stock,
		Author: author.AuthorResponse{
			ID:   book.Author.ID,
			Name: book.Author.Name,
		},
		Category: category.CategoryResponse{
			ID:   book.Category.ID,
			Name: book.Category.Name,
		},
	}
}
