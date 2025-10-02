package recovery

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor returns a new unary server interceptor for panic recovery.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newOpts(opts...)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(ctx, r, o.handler)
			}
		}()

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a new streaming server interceptor for panic recovery.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := newOpts(opts...)
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = recoverFrom(stream.Context(), r, o.handler)
			}
		}()

		return handler(srv, stream)
	}
}

func recoverFrom(ctx context.Context, p any, r Handler) error {
	if r != nil {
		return r(ctx, p)
	}
	return StackHandler(64<<10)(ctx, p)
}
