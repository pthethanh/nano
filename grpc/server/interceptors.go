package server

import (
	"context"

	"google.golang.org/grpc"
)

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

var (
	ChainUnaryInterceptor  = grpc.ChainUnaryInterceptor
	ChainStreamInterceptor = grpc.ChainStreamInterceptor
	UnaryInterceptor       = grpc.UnaryInterceptor
	StreamInterceptor      = grpc.StreamInterceptor
)

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// newWrapServerStream returns a ServerStream that has the ability to overwrite context.
func newWrapServerStream(_ context.Context, stream grpc.ServerStream) grpc.ServerStream {
	if existing, ok := stream.(*wrappedServerStream); ok {
		return existing
	}
	return &wrappedServerStream{
		ServerStream: stream,
		ctx:          stream.Context(),
	}
}

func ContextUnaryInterceptor(f func(context.Context) (context.Context, error)) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		newCtx, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

func ContextStreamInterceptor(f func(context.Context) (context.Context, error)) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := f(ss.Context())
		if err != nil {
			return err
		}
		return handler(srv, newWrapServerStream(newCtx, ss))
	}
}
