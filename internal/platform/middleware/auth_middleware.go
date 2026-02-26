package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Carlvalencia1/streamhub-backend/internal/platform/security"
)

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		var tokenString string

		// 🔥 1️⃣ Intentar Authorization Header
		authHeader := c.GetHeader("Authorization")

		if authHeader != "" {

			parts := strings.Split(authHeader, " ")

			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// 🔥 2️⃣ Si no hay header → intentar query param (WebSocket)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing token",
			})
			c.Abort()
			return
		}

		claims, err := security.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			c.Abort()
			return
		}

		// 🔥 CLAVES CORRECTAS
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}