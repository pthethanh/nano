package auth

import (
	"context"

	"google.golang.org/grpc"
)

const (
	// AuthorizationMD authorization metadata name.
	AuthorizationMD = "authorization"
)

// Authenticator defines the interface to perform the actual
// authentication of the request. Implementations should fetch
// the required data from the context.Context object. GRPC specific
// data like `metadata` and `peer` is available on the context.
// Should return a new `context.Context` that is a child of `ctx`
// or `codes.Unauthenticated` when auth is lacking or
// `codes.PermissionDenied` when lacking permissions.
type Authenticator interface {
	Authenticate(ctx context.Context) (context.Context, error)
}

// AuthenticatorFunc defines a pluggable function to perform authentication
// of requests. Should return a new `context.Context` that is a
// child of `ctx` or `codes.Unauthenticated` when auth is lacking or
// `codes.PermissionDenied` when lacking permissions.
type AuthenticatorFunc func(ctx context.Context) (context.Context, error)

// Authenticate implements the Authenticator interface
func (f AuthenticatorFunc) Authenticate(ctx context.Context) (context.Context, error) {
	return f(ctx)
}

// StreamInterceptor returns a grpc.StreamServerInterceptor that performs
// an authentication check for each request by using
// Authenticator.Authenticate(ctx context.Context).
func StreamInterceptor(auth Authenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var newCtx context.Context
		var err error
		a := auth
		// if server overrides Authenticator, use it instead.
		if srvAuth, ok := srv.(Authenticator); ok {
			a = srvAuth
		}
		newCtx, err = a.Authenticate(ss.Context())
		if err != nil {
			return err
		}
		return handler(srv, newContextServerStream(newCtx, ss))
	}
}

// UnaryInterceptor returns a grpc.UnaryServerInterceptor that performs
// an authentication check for each request by using
// Authenticator.Authenticate(ctx context.Context).
func UnaryInterceptor(auth Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var newCtx context.Context
		var err error
		a := auth
		// if server override Authenticator, use it instead.
		if srvAuth, ok := info.Server.(Authenticator); ok {
			a = srvAuth
		}
		newCtx, err = a.Authenticate(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
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

// newContextServerStream returns a ServerStream with the new context.
// If the stream is already a ContextServerStream, it returns the existing one.
// This is useful for interceptors that need to modify the context.
func newContextServerStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	if existing, ok := stream.(*contextServerStream); ok {
		return existing
	}
	return &contextServerStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}
