package router

import (
	"ferdinand/library-management-system-api/internal/auth"
	"ferdinand/library-management-system-api/internal/author"
	"ferdinand/library-management-system-api/internal/book"
	"ferdinand/library-management-system-api/internal/borrowing"
	"ferdinand/library-management-system-api/internal/category"
	"ferdinand/library-management-system-api/internal/member"
	"ferdinand/library-management-system-api/internal/user"

	"github.com/gin-gonic/gin"
)

type RouteHandlers struct {
	BookHandler      *book.Handler
	CategoryHandler  *category.Handler
	AuthorHandler    *author.Handler
	MemberHandler    *member.Handler
		BorrowingHandler *borrowing.Handler
		UserHandler      *user.Handler
		AuthHandler 	*auth.Handler
	}

	func New(handlers RouteHandlers, jwtService *auth.JWTService) *gin.Engine {

		r := gin.Default()
		v1 := r.Group("/api/v1")

		
		// public
		auth.RegisterRoutes(v1,handlers.AuthHandler)
		
		// auth protected 
		protected := v1.Group("")
		protected.Use(auth.AuthMiddleware(jwtService))
		book.RegisterRoutes(protected, handlers.BookHandler)
		borrowing.RegisterRoutes(protected, handlers.BorrowingHandler)
		category.RegisterRoutes(protected, handlers.CategoryHandler)
		author.RegisterRoutes(protected, handlers.AuthorHandler)

		user.RegisterRoutes(protected, handlers.UserHandler)
		
		// admin
		admin := protected.Group("")
		admin.Use(auth.RoleMiddleware("admin"))
		member.RegisterRoutes(admin, handlers.MemberHandler)



		return r
	
	}

