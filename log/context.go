package log

import (
	"context"
)

type (
	contextKey      struct{}
	contextKeyAttrs struct{}
)

// NewContext returns a new context with the given logger.
func NewContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext retrieves the logger from the context if present.
func FromContext(ctx context.Context) *Logger {
	if v := ctx.Value(contextKey{}); v != nil {
		if l, ok := v.(*Logger); ok {
			return l
		}
	}
	return nil
}

// NewAttrsContext returns a new context with logging attributes.
func NewAttrsContext(ctx context.Context, attrs ...any) context.Context {
	return context.WithValue(ctx, contextKeyAttrs{}, attrs)
}

// AttrsFromContext retrieves logging attributes from the context if present.
func AttrsFromContext(ctx context.Context) []any {
	if v := ctx.Value(contextKeyAttrs{}); v != nil {
		if attrs, ok := v.([]any); ok {
			return attrs
		}
	}
	return nil
}

// AppendToContext appends logging attributes to the context.
func AppendToContext(ctx context.Context, attrs ...any) context.Context {
	newAttrs := append([]any{}, attrs...)
	if v := ctx.Value(contextKeyAttrs{}); v != nil {
		if existingAttrs, ok := v.([]any); ok {
			newAttrs = append(newAttrs, existingAttrs...)
		}
	}
	return context.WithValue(ctx, contextKeyAttrs{}, newAttrs)
}
