package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
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
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid admin token"})
			return
		}
		c.Next()
	}
}
