package author

import (
	"context"
	"errors"
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

func (h *Handler) GetAuthors(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	authors, err := h.service.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authors": authors,
	})
}

func (h *Handler) GetAuthor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid author id",
		})
		return
	}

	author, err := h.service.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "author not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"author": author,
	})
}

func (h *Handler) CreateAuthor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var authorRequest CreateAuthorRequest
	if err := c.ShouldBindJSON(&authorRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	author, err := h.service.Create(ctx, authorRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"author": author,
	})
}

func (h *Handler) UpdateAuthor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid author id",
		})
		return
	}

	var authorUpdateRequest UpdateAuthorRequest
	if err := c.ShouldBindJSON(&authorUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	updatedAuthor, err := h.service.Update(ctx, id, authorUpdateRequest)
	if err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "author not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"author": updatedAuthor,
	})
}

func (h *Handler) DeleteAuthor(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid author id",
		})
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrAuthorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "author not found",
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