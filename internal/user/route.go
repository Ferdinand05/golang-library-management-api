package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {
	users := api.Group("/users")

	users.GET("", handler.GetUsers)
	users.GET("/:id", handler.GetUser)
	users.POST("", handler.CreateUser)
}
