package borrowing

import (
	"ferdinand/library-management-system-api/internal/auth"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api *gin.RouterGroup, handler *Handler) {

	borrowings := api.Group("/borrowings")
	borrowings.GET("/:id", handler.GetBorrowing)
	borrowings.PATCH("/:id/return", handler.ReturnBorrowing)
	borrowings.POST("/", handler.CreateBorrowing)
	
	
	// admin
	admin := borrowings.Group("")
	admin.GET("/member/:id", handler.GetBorrowingByMemberID)
	admin.Use(auth.RoleMiddleware("admin"))
	admin.GET("/", handler.GetBorrowings)

}
