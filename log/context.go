package log

import (
	"context"
)

type (
	contextKey      struct{}
	contextKeyAttrs struct{}
)

// NewContext return new context with the given logger inside.
func NewContext(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext return logger from the given context if exist
func FromContext(ctx context.Context) *Logger {
	if v := ctx.Value(contextKey{}); v != nil {
		if l, ok := v.(*Logger); ok {
			return l
		}
	}
	return nil
}

// NewAttrsContext return new logging attributes context.
func NewAttrsContext(ctx context.Context, attrs ...any) context.Context {
	return context.WithValue(ctx, contextKeyAttrs{}, attrs)
}

// AttrsFromContext retrieve log attributes from context if any.
func AttrsFromContext(ctx context.Context) []any {
	if v := ctx.Value(contextKeyAttrs{}); v != nil {
		if attrs, ok := v.([]any); ok {
			return attrs
		}
	}
	return nil
}

// AppendToContext append logging attributes to the given context.
func AppendToContext(ctx context.Context, attrs ...any) context.Context {
	newAttrs := append([]any{}, attrs...)
	if v := ctx.Value(contextKeyAttrs{}); v != nil {
		if existingAttrs, ok := v.([]any); ok {
			newAttrs = append(newAttrs, existingAttrs...)
		}
	}
	return context.WithValue(ctx, contextKeyAttrs{}, newAttrs)
}
