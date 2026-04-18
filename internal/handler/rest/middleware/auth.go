package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/platform/logger"
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
			logger.FromContext(c.Request.Context()).Debug("auth: reject token", slog.String("err", err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, userID)
		ctx := logger.WithContext(c.Request.Context(),
			logger.FromContext(c.Request.Context()).With(slog.String("user_id", userID.String())))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
