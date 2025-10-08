package authz

import (
	"context"

	"google.golang.org/grpc"
)

var (
	_ Authorizer = AuthorizerFunc(nil)
)

// Authorizer defines the generic interface for authorization enforcement.
// This interface is designed to work with various policy engines including
// Casbin, OPA (Open Policy Agent), grpc-authz, and custom implementations.
//
// All necessary information for authorization (method, request, subject, etc.)
// should be retrieved from the context using helper functions like GetMethod(),
// GetSubject(), GetRequest(), etc.
//
// Implementations should return true if the request is authorized, false otherwise.
type Authorizer interface {
	// Authorize decides whether the request is permitted based on information
	// stored in the context. The context contains method name, request data,
	// subject (user), and any other authentication/authorization metadata.
	// Returns new context if authorized or an error if the authorization check fails.
	Authorize(ctx context.Context) (context.Context, error)
}

// AuthorizerFunc is a function type that implements the Authorizer interface.
// It allows using ordinary functions as Authorizers without defining a new type.
//
// Example:
//
//	authorizer := authz.AuthorizerFunc(func(ctx context.Context) (context.Context, error) {
//		subject := authz.SubjectFromContext(ctx)
//		if subject == "admin" {
//			return ctx, nil
//		}
//		return nil, status.Errorf(codes.PermissionDenied, "access denied")
//	})
type AuthorizerFunc func(ctx context.Context) (context.Context, error)

// Authorize implements the Authorizer interface by calling the function itself.
func (f AuthorizerFunc) Authorize(ctx context.Context) (context.Context, error) {
	return f(ctx)
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that performs
// authorization checks using the provided Authorizer.
// It enriches the context with method and request information before calling
// the Authorizer.
func UnaryServerInterceptor(authorizer Authorizer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Enrich context with authorization metadata
		ctx = NewMethodContext(ctx, info.FullMethod)
		ctx = NewRequestContext(ctx, req)

		// Perform authorization check
		newCtx, err := authorizer.Authorize(ctx)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that performs
// authorization checks using the provided Authorizer.
// It enriches the context with method information before calling the Authorizer.
func StreamServerInterceptor(authorizer Authorizer) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		// Enrich context with authorization metadata
		ctx = NewMethodContext(ctx, info.FullMethod)

		// Perform authorization check
		newCtx, err := authorizer.Authorize(ctx)
		if err != nil {
			return err
		}
		return handler(srv, &contextServerStream{ctx: newCtx, ServerStream: ss})
	}
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context for the server stream.
func (w *contextServerStream) Context() context.Context {
	return w.ctx
}
