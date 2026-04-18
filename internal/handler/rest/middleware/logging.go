package middleware

import (
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"voronka/internal/platform/logger"
)

const (
	// RequestIDHeader is the canonical header used for correlation.
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the Gin context key under which the ID is stored.
	RequestIDKey = "requestID"
)

// RequestID ensures every request has an X-Request-ID; accepts an inbound
// value or generates a UUID. The id is stored on the gin context and echoed
// back to the client.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// Logger attaches a request-scoped slog.Logger to the request context and
// emits one structured line per request. Replaces gin.Logger().
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID, _ := c.Get(RequestIDKey)
		base := slog.Default().With(
			slog.String("request_id", stringify(reqID)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)
		c.Request = c.Request.WithContext(logger.WithContext(c.Request.Context(), base))

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		attrs := []any{
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("bytes", c.Writer.Size()),
		}
		if q := c.Request.URL.RawQuery; q != "" {
			attrs = append(attrs, slog.String("query", q))
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			attrs = append(attrs, slog.String("gin_errors", errs))
		}

		switch {
		case status >= 500:
			base.LogAttrs(c.Request.Context(), slog.LevelError, "request", toAttrs(attrs)...)
		case status >= 400:
			base.LogAttrs(c.Request.Context(), slog.LevelWarn, "request", toAttrs(attrs)...)
		default:
			base.LogAttrs(c.Request.Context(), slog.LevelInfo, "request", toAttrs(attrs)...)
		}
	}
}

// Recovery logs panics with the request-scoped logger and returns 500.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.FromContext(c.Request.Context()).Error("panic recovered",
					slog.Any("panic", rec),
					slog.String("stack", string(debug.Stack())),
				)
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
				} else {
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}

func toAttrs(in []any) []slog.Attr {
	out := make([]slog.Attr, 0, len(in))
	for _, v := range in {
		if a, ok := v.(slog.Attr); ok {
			out = append(out, a)
		}
	}
	return out
}

func stringify(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
