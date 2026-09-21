package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup,handler *Handler) {
	
	auth := api.Group("/auth")

	auth.POST("/login",handler.Login)

}