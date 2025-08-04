package log

import (
	"context"
	"log/slog"
)

type (
	Logger struct {
		*slog.Logger
	}
)

// New creates a new Logger from slog.Logger.
func New(logger *slog.Logger) *Logger {
	return &Logger{
		Logger: logger,
	}
}

// DebugContext logs a debug message with context.
func (logger *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).DebugContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context.
func (logger *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).ErrorContext(ctx, msg, args...)
}

// InfoContext logs an info message with context.
func (logger *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).InfoContext(ctx, msg, args...)
}

// Log logs a message at the specified level with context.
func (logger *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger.context(ctx).Log(ctx, level, msg, args...)
}

// LogAttrs logs a message with attributes at the specified level and context.
func (logger *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logger.context(ctx).LogAttrs(ctx, level, msg, attrs...)
}

// WarnContext logs a warning message with context.
func (logger *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).WarnContext(ctx, msg, args...)
}

// With returns a logger with additional attributes.
func (logger *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: logger.Logger.With(args...),
	}
}

// WithGroup returns a logger with the specified group name.
func (logger *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: logger.Logger.WithGroup(name),
	}
}

func (logger *Logger) context(ctx context.Context) *slog.Logger {
	log := logger.Logger
	if l := FromContext(ctx); l != nil {
		log = l.Logger
	}
	if attrs := AttrsFromContext(ctx); len(attrs) > 0 {
		return log.With(attrs...)
	}
	return log
}
