package server

import (
	"context"

	"google.golang.org/grpc"
)

// ContextUnaryInterceptor injects context modifications into unary server calls.
func ContextUnaryInterceptor(f func(context.Context) (context.Context, error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		newCtx, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

// ContextStreamInterceptor injects context modifications into stream server calls.
func ContextStreamInterceptor(f func(context.Context) (context.Context, error)) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := f(ss.Context())
		if err != nil {
			return err
		}
		return handler(srv, NewContextServerStream(newCtx, ss))
	}
}

// DeferContextUnaryInterceptor defers a context function after unary server calls.
func DeferContextUnaryInterceptor(f func(context.Context)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer f(ctx)
		return handler(ctx, req)
	}
}

// DeferContextStreamInterceptor defers a context function after stream server calls.
func DeferContextStreamInterceptor(f func(context.Context)) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		defer f(ss.Context())
		return handler(srv, NewContextServerStream(ss.Context(), ss))
	}
}
