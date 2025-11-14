package middleware

import (
	"net/http"
	"soa-video-streaming/pkg/cookie"
	"soa-video-streaming/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
)

func JWTAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := cookie.GetAccessToken(c)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		claims, err := authService.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token payload",
			})
			return
		}

		c.Set("user_id", userID)

		c.Next()
	}
}
