package ratelimit

import (
	"context"
	"errors"

	"google.golang.org/grpc"
)

var ErrLimited = errors.New("rate limit exceeded")

// Limiter reports whether a call is allowed to proceed.
type Limiter interface {
	Allow(context.Context) error
}

// AllowFunc adapts a function into a Limiter.
type AllowFunc func(context.Context) error

func (f AllowFunc) Allow(ctx context.Context) error { return f(ctx) }

// UnaryClientInterceptor returns a client interceptor guarded by the provided limiter.
func UnaryClientInterceptor(limiter Limiter) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := limiter.Allow(ctx); err != nil {
			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a stream client interceptor guarded by the provided limiter.
func StreamClientInterceptor(limiter Limiter) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if err := limiter.Allow(ctx); err != nil {
			return nil, err
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// UnaryServerInterceptor returns a server interceptor guarded by the provided limiter.
func UnaryServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := limiter.Allow(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a stream server interceptor guarded by the provided limiter.
func StreamServerInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := limiter.Allow(stream.Context()); err != nil {
			return err
		}
		return handler(srv, stream)
	}
}
