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

func SetDefault(log *Logger) {
	def.Store(log)
}

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

func Named(name string) *Logger {
	return Default().With("logger", name)
}

// Debug calls Logger.Debug on the default logger.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// DebugContext calls Logger.DebugContext on the default logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

// Info calls Logger.Info on the default logger.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// InfoContext calls Logger.InfoContext on the default logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

// Warn calls Logger.Warn on the default logger.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// WarnContext calls Logger.WarnContext on the default logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

// Error calls Logger.Error on the default logger.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// ErrorContext calls Logger.ErrorContext on the default logger.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// Log calls Logger.Log on the default logger.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Default().Log(ctx, level, msg, args...)
}

// LogAttrs calls Logger.LogAttrs on the default logger.
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
