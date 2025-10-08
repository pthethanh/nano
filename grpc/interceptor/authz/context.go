package authz

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	// Metadata keys for authorization context
	metadataKeyMethod  = "x-authz-method"
	metadataKeySubject = "x-authz-subject"
)

type (
	requestKey    struct{}
	anyContextKey struct{}
)

// NewMethodContext adds the gRPC method name to the context via metadata.
// This is automatically called by the interceptor.
func NewMethodContext(ctx context.Context, method string) context.Context {
	md := metadata.Pairs(metadataKeyMethod, method)
	if existingMD, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(existingMD, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}

// MethodFromContext retrieves the gRPC method name from the context metadata.
// Returns empty string if not found.
func MethodFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(metadataKeyMethod)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

// NewRequestContext adds the gRPC request to the context.
// This is automatically called by the unary interceptor.
// Note: Request cannot be stored in metadata as it's not a string,
// so it uses context.Value for storage.
func NewRequestContext(ctx context.Context, req any) context.Context {
	return context.WithValue(ctx, requestKey{}, req)
}

// RequestFromContext retrieves the gRPC request from the context.
// Returns nil if not found.
func RequestFromContext(ctx context.Context) any {
	return ctx.Value(requestKey{})
}

// NewSubjectContext adds the authenticated subject (user/role) to the context via metadata.
// This should typically be called by an authentication interceptor.
func NewSubjectContext(ctx context.Context, subject string) context.Context {
	md := metadata.Pairs(metadataKeySubject, subject)
	if existingMD, ok := metadata.FromIncomingContext(ctx); ok {
		md = metadata.Join(existingMD, md)
	}
	return metadata.NewIncomingContext(ctx, md)
}

// SubjectFromContext retrieves the authenticated subject from the context metadata.
// Returns empty string if not found.
// Also checks legacy "subject" and "user" keys for backward compatibility.
func SubjectFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// Check x-authz-subject first
		if values := md.Get(metadataKeySubject); len(values) > 0 && values[0] != "" {
			return values[0]
		}
		// Backward compatibility: check "subject" key
		if values := md.Get("subject"); len(values) > 0 && values[0] != "" {
			return values[0]
		}
		// Backward compatibility: check "user" key
		if values := md.Get("user"); len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}

	// Fallback to context.Value for backward compatibility
	if subject, ok := ctx.Value("subject").(string); ok && subject != "" {
		return subject
	}
	if user, ok := ctx.Value("user").(string); ok && user != "" {
		return user
	}

	return ""
}

// NewAnyContext adds a value of any type to the context using a type-safe generic key.
// Each unique type T gets its own isolated context key, preventing collisions.
// Example:
//
//	type UserID int
//	ctx = NewAnyContext(ctx, UserID(123))
//	userID := FromAnyContext[UserID](ctx) // returns 123
func NewAnyContext[T any](ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, anyContextKey{}, value)
}

// FromAnyContext retrieves a value of type T from the context.
// Returns the zero value of T if not found or type doesn't match.
// Each unique type T has its own isolated context key, ensuring type safety.
func FromAnyContext[T any](ctx context.Context) T {
	if value, ok := ctx.Value(anyContextKey{}).(T); ok {
		return value
	}
	var zero T
	return zero
}
