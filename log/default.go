package log

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pthethanh/nano/log/zap"
)

var (
	def  atomic.Pointer[Logger]
	once sync.Once
)

// SetDefault sets the default logger.
func SetDefault(log *Logger) {
	def.Store(log)
}

// Default returns the default logger, creating one if needed.
func Default() *Logger {
	once.Do(func() {
		if def.Load() != nil {
			return
		}
		log := &Logger{
			Logger: slog.New(zap.NewHandler(zap.Config{
				Name:             getEnv("LOG_NAME", ""),
				Format:           getEnv("LOG_FORMAT", "json"),
				AddSource:        getEnv("LOG_ADDSOURCE", "false") == "true",
				Level:            getEnv("LOG_LEVEL", slog.LevelDebug.String()),
				OutputPaths:      strings.Split(getEnv("LOG_OUTPUTPATHS", "stderr"), ","),
				ErrorOutputPaths: strings.Split(getEnv("LOG_ERROROUTPUTPATHS", "stderr"), ","),
			})),
		}
		def.Store(log)
	})
	return def.Load()
}

// Named returns a logger with the given name.
func Named(name string) *Logger {
	return Default().With("logger", name)
}

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// DebugContext logs a debug message with context using the default logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

// Info logs an info message using the default logger.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// InfoContext logs an info message with context using the default logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// WarnContext logs a warning message with context using the default logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

// Error logs an error message using the default logger.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// ErrorContext logs an error message with context using the default logger.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// Log logs a message at the specified level with context using the default logger.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Default().Log(ctx, level, msg, args...)
}

// LogAttrs logs a message with attributes at the specified level using the default logger.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Default().LogAttrs(ctx, level, msg, attrs...)
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v != "" {
		return v
	}
	return def
}
