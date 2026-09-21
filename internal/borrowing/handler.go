package borrowing

import (
	"context"
	"errors"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/member"
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

func (h *Handler) GetBorrowings(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	borrowings, err := h.service.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	var responses = make([]BorrowingResponse, 0, len(borrowings))
	for _, response := range borrowings {
		responses = append(responses, toBorrowingResponse(response))
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowings": responses,
	})
}

func (h *Handler) GetBorrowing(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid borrowing id",
		})
		return
	}

	borrowing, err := h.service.FindByID(ctx, id)
	if err != nil {

		if errors.Is(err, ErrBorrowingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "borrowing not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowing": toBorrowingResponse(borrowing),
	})
}

func (h *Handler) GetBorrowingByMemberID(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid member id",
		})
		return
	}

	borrowings, err := h.service.FindByMemberID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	var responses = make([]BorrowingResponse, 0, len(borrowings))
	for _, response := range borrowings {
		responses = append(responses, toBorrowingResponse(response))
	}

	c.JSON(http.StatusOK, gin.H{
		"borrowings": responses,
	})

}

func (h *Handler) CreateBorrowing(c *gin.Context) {

	var createRequest CreateBorrowingRequest

	if err := c.ShouldBindJSON(&createRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	ctx, cancel := context.
		WithTimeout(
			c.Request.Context(),
			3*time.Second)
	defer cancel()

	newBorrowing, err := h.service.Create(ctx, createRequest)
	if err != nil {

		if errors.Is(err, member.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "member not found",
			})
			return
		}

		if errors.Is(err, book.ErrBookOutOfStock) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "book out of stock",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"borrowing": toBorrowingResponse(newBorrowing),
	})

}

func (h *Handler) ReturnBorrowing(c *gin.Context) {
	borrowingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid borrowing id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	err = h.service.Return(ctx, borrowingID)
	if err != nil {
		if errors.Is(err, ErrBorrowingNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "borrowing not found",
			})
			return
		}

		if errors.Is(err, ErrBorrowingAlreadyReturned) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "borrowing already returned",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "book returned successfully",
	})

}

func toBorrowingResponse(b Borrowing) BorrowingResponse {
	return BorrowingResponse{
		ID: int64(b.ID),
		Member: MemberResponse{
			ID:   b.Member.ID,
			Code: b.Member.MemberCode,
			Name: b.Member.Name,
		},
		Book: BookResponse{
			ID:    b.Book.ID,
			ISBN:  b.Book.ISBN,
			Title: b.Book.Title,
		},
		BorrowedAt: b.BorrowedAt,
		DueAt:      b.DueAt,
		ReturnedAt: b.ReturnedAt,
	}
}
