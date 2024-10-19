package log

import (
	"context"
	"log/slog"
)

type (
	Logger struct {
		*slog.Logger
		ctxs []ContextRetriever
	}

	Option           = func(*Logger)
	ContextRetriever = func(context.Context) []any
)

func New(logger *slog.Logger, opts ...Option) *Logger {
	l := &Logger{
		Logger: logger,
	}
	return l.apply(opts...)
}

func (logger *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).DebugContext(ctx, msg, args...)
}

func (logger *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).ErrorContext(ctx, msg, args...)
}

func (logger *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).InfoContext(ctx, msg, args...)
}

func (logger *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	logger.context(ctx).Log(ctx, level, msg, args...)
}

func (logger *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	logger.context(ctx).LogAttrs(ctx, level, msg, attrs...)
}

func (logger *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	logger.context(ctx).WarnContext(ctx, msg, args...)
}

func (logger *Logger) With(args ...any) *Logger {
	return &Logger{
		Logger: logger.Logger.With(args...),
		ctxs:   logger.ctxs,
	}
}
func (logger *Logger) WithGroup(name string) *Logger {
	return &Logger{
		Logger: logger.Logger.WithGroup(name),
		ctxs:   logger.ctxs,
	}
}

func (logger *Logger) context(ctx context.Context) *slog.Logger {
	log := logger.Logger
	if l := FromContext(ctx); l != nil {
		log = l.Logger
	}
	var attrs []any
	// retrieve context data
	if vs := AttrsFromContext(ctx); len(vs) > 0 {
		attrs = append(attrs, vs...)
	}
	// retrieve more context data from custom resolver
	for _, rs := range logger.ctxs {
		attrs = append(attrs, rs(ctx)...)
	}
	if len(attrs) > 0 {
		return log.With(attrs...)
	}
	return log
}

func (logger *Logger) apply(opts ...Option) *Logger {
	for _, opt := range opts {
		opt(logger)
	}
	return logger
}
