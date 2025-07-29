package client

import (
	"context"

	"google.golang.org/grpc"
)

var (
	WithChainUnaryInterceptor  = grpc.WithChainUnaryInterceptor
	WithChainStreamInterceptor = grpc.WithChainStreamInterceptor
)

func ContextUnaryInterceptor(f func(context.Context) (context.Context, error)) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx, err := f(ctx)
		if err != nil {
			return err
		}
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

func ContextStreamInterceptor(f func(context.Context) (context.Context, error)) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newCtx, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return streamer(newCtx, desc, cc, method, opts...)
	}
}

func DeferContextUnaryInterceptor(f func(context.Context)) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		defer f(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func DeferContextStreamInterceptor(f func(context.Context)) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		defer f(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
