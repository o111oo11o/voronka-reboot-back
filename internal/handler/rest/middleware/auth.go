package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/service"
)

const UserIDKey = "userID"

// TokenParser is the subset of AuthService used by the middleware.
type TokenParser interface {
	ParseAccessToken(raw string) (uuid.UUID, error)
}

// Auth validates the Bearer access token and sets UserIDKey in the context.
func Auth(svc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		raw := strings.TrimPrefix(header, "Bearer ")
		if raw == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must use Bearer scheme"})
			return
		}

		userID, err := svc.ParseAccessToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}
