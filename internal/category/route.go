package category

import (
	"ferdinand/library-management-system-api/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {

	categories := api.Group("/categories")

	categories.GET("/", handler.GetCategories)
	categories.GET("/:id", handler.GetCategory)

	// admin
	admin := categories.Group("")
	admin.Use(auth.RoleMiddleware("admin"))
	admin.POST("/", handler.CreateCategory)
	admin.PUT("/:id", handler.UpdateCategory)
	admin.DELETE("/:id", handler.DeleteCategory)

}
