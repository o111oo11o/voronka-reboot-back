// Package logger configures the process-wide slog.Default() logger and
// exposes helpers for request-scoped logging.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

// ctxKey is the unexported type used to store a *slog.Logger on a context.
type ctxKey struct{}

// Config controls logger initialization.
type Config struct {
	Level  string // "debug" | "info" | "warn" | "error"
	Format string // "json" | "text"
}

// Init installs a configured slog.Logger as the default. Safe to call once at startup.
func Init(cfg Config) *slog.Logger {
	return initTo(os.Stdout, cfg)
}

func initTo(w io.Writer, cfg Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: true,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "text":
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithContext returns ctx carrying the given logger.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	if l == nil {
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext returns the logger attached to ctx, or slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if l, ok := ctx.Value(ctxKey{}).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
