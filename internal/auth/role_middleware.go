package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {

		currentRole := c.GetString("role")

		if currentRole == "" {
			c.JSON(http.StatusUnauthorized,gin.H{
				"error" : "unauthorized",
			})
			c.Abort()
			return
		}

		if currentRole != requiredRole {
			c.JSON(http.StatusForbidden,gin.H{
				"error" : "forbidden",
			})
			c.Abort()
			return
		}

		c.Next()

	}
}