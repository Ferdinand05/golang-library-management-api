package member

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

func (h *Handler) GetMembers(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	members, err := h.service.FindAll(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	var responses = make([]MemberResponse, 0, len(members))

	for _, response := range members {
		responses = append(responses, toMemberResponse(response))
	}

	c.JSON(http.StatusOK, gin.H{
		"members": responses,
	})

}

func (h *Handler) GetMember(c *gin.Context) {

	ctx := c.Request.Context()
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid member id",
		})
		return
	}

	member, err := h.service.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "member not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"member": toMemberResponse(member),
	})

}

func (h *Handler) CreateMember(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	var memberRequest CreateMemberRequest
	if err := c.ShouldBindJSON(&memberRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	newMember, err := h.service.Create(ctx, memberRequest)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"member": toMemberResponse(newMember),
	})

}

func (h *Handler) UpdateMember(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid member id",
		})
		return
	}

	var memberUpdateRequest UpdateMemberRequest

	if err := c.ShouldBindJSON(&memberUpdateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	updatedMember, err := h.service.Update(ctx, id, memberUpdateRequest)
	if err != nil {

		if errors.Is(err, ErrMemberNotFound) {

			c.JSON(http.StatusNotFound, gin.H{
				"error": "member not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"member": toMemberResponse(updatedMember),
	})

}

func (h *Handler) DeleteMember(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid member id",
		})
		return
	}

	err = h.service.Delete(ctx, id)
	if err != nil {

		if errors.Is(err, ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "member not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "data member deleted successfully",
	})

}

func toMemberResponse(m Member) MemberResponse {

	var address string
	if m.Address != nil {
		address = *m.Address
	}

	return MemberResponse{
		ID:         m.ID,
		Name:       m.Name,
		MemberCode: m.MemberCode,
		Email:      m.Email,
		Phone:      m.Phone,
		Address:    address,
	}
}
