package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("x-user-id")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: missing identity header",
			})
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}
