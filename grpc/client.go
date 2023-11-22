package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "google.golang.org/grpc/health" // enable health check
)

// Dial creates a client connection to the given target with health check enabled
// and some others default configurations.
func Dial(ctx context.Context, address string, options ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append([]grpc.DialOption{}, options...)
	if len(opts) == 0 {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// MustDial same as Dial but panic if error.
func MustDial(ctx context.Context, address string, options ...grpc.DialOption) *grpc.ClientConn {
	conn, err := Dial(ctx, address, options...)
	if err != nil {
		panic(err)
	}
	return conn
}

func ContextUnaryClientInterceptor(f ContextWrapFunc) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx, err := f(ctx)
		if err != nil {
			return err
		}
		return invoker(newCtx, method, req, reply, cc, opts...)
	}
}

func ContextStreamClientInterceptor(f ContextWrapFunc) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newCtx, err := f(ctx)
		if err != nil {
			return nil, err
		}
		return streamer(newCtx, desc, cc, method, opts...)
	}
}
