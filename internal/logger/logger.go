package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Config struct {
	Level       slog.Level
	Format      string
	Output      io.Writer
	ServiceName string
	Version     string
	Environment string
}

func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	var handler slog.Handler

	attrs := []slog.Attr{
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.Version),
		slog.String("env", cfg.Environment),
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: true,
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	}

	handler = handler.WithAttrs(attrs)

	return slog.New(handler)
}

type loggerKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
