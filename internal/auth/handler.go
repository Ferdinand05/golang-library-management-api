package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *gin.Context) {

	ctx,cancel := context.WithTimeout(c.Request.Context(),3*time.Second	)
	defer cancel()

	var loginRequest LoginRequest

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{
			"error" : "invalid request body",
		})
		return
	}

	token,err := h.service.Login(ctx,loginRequest)
	if err != nil {

		if errors.Is(err,ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized,gin.H{
				"error" : "invalid credentials",
			})
			return
		}

		c.JSON(http.StatusInternalServerError,gin.H{
			"error" : "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK,gin.H{
		"token" : token,
	})

}