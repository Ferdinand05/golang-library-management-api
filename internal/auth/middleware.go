package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwtService *JWTService) gin.HandlerFunc {

	return  func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header not found"})
			c.Abort() 
			return
		}

		parts := strings.SplitN(authHeader," ",2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error" : "invalid token format",
			})
			c.Abort()
			return
		} 

		token := parts[1]
		claims,err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error" : "invalid token",
			})
			c.Abort()
			return
		}

		c.Set("userID",claims.Subject)
		c.Set("role",claims.Role)
		c.Set("email",claims.Email)

		c.Next()

	}

}