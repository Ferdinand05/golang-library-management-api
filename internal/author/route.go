package author

import (
	"ferdinand/library-management-system-api/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {

	authors := api.Group("/authors")

	authors.GET("/", handler.GetAuthors)
	authors.GET("/:id", handler.GetAuthor)

	// admin
	admin := authors.Group("")
	admin.Use(auth.RoleMiddleware("admin"))
	admin.POST("/", handler.CreateAuthor)
	admin.PUT("/:id", handler.UpdateAuthor)
	admin.DELETE("/:id", handler.DeleteAuthor)

}
