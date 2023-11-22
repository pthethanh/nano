package grpc

import (
	"context"

	"google.golang.org/grpc"
)

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapper's WrappedContext, overwriting the nested grpc.ServerStream.Context()
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// newWrapServerStream returns a ServerStream that has the ability to overwrite context.
func newWrapServerStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	if existing, ok := stream.(*wrappedServerStream); ok {
		return existing
	}
	return &wrappedServerStream{
		ServerStream: stream,
		ctx:          stream.Context(),
	}
}

func ContextUnaryServerInterceptor(f ContextWrapFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		newCtx, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return handler(newCtx, req)
	}
}

func ContextStreamServerInterceptor(f ContextWrapFunc) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, err := f(ss.Context())
		if err != nil {
			return err
		}
		return handler(srv, newWrapServerStream(newCtx, ss))
	}
}
