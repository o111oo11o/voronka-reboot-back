package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"voronka/internal/platform/logger"
)

// Admin checks the X-Admin-Token header against the configured token.
// Uses constant-time comparison to prevent timing attacks.
func Admin(token string) gin.HandlerFunc {
	expected := []byte(token)
	return func(c *gin.Context) {
		got := c.GetHeader("X-Admin-Token")
		if got == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-Admin-Token header"})
			return
		}
		if subtle.ConstantTimeCompare([]byte(got), expected) != 1 {
			logger.FromContext(c.Request.Context()).Warn("admin: rejected token",
				slog.String("client_ip", c.ClientIP()),
				slog.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid admin token"})
			return
		}
		c.Next()
	}
}
