package book

import (
	"ferdinand/library-management-system-api/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {

	books := api.Group("/books")
	books.GET("/", handler.GetBooks)
	books.GET("/:id", handler.GetBook)
	api.GET("/authors/:id/books", handler.GetBooksByAuthor)
	api.GET("/categories/:id/books", handler.GetBooksByCategory)

	// admin
	admin := books.Group("")
	admin.Use(auth.RoleMiddleware("admin"))
	admin.POST("/", handler.CreateBook)
	admin.PUT("/:id", handler.UpdateBook)
	admin.DELETE("/:id", handler.DeleteBook)

}
