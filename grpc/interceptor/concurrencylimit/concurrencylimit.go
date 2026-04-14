package concurrencylimit

import (
	"context"

	"google.golang.org/grpc"
)

// Limiter acquires and releases concurrency slots.
type Limiter interface {
	Acquire(context.Context) error
	Release()
}

// UnaryServerInterceptor returns a server interceptor that enforces concurrency limits.
func UnaryServerInterceptor(limiter Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if err := limiter.Acquire(ctx); err != nil {
			return nil, err
		}
		defer limiter.Release()
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a stream server interceptor that enforces concurrency limits.
func StreamServerInterceptor(limiter Limiter) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if err := limiter.Acquire(stream.Context()); err != nil {
			return err
		}
		defer limiter.Release()
		return handler(srv, stream)
	}
}
