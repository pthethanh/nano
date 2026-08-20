// Package deadline provides gRPC interceptors that enforce a default and/or
// maximum request deadline server-side, instead of trusting every caller to
// set one.
package deadline

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

type options struct {
	def time.Duration // applied when the caller set no deadline; 0 disables
	max time.Duration // caps any deadline, including a caller-set one; 0 disables
}

// Option customizes deadline enforcement.
type Option func(*options)

// WithDefault applies d as the deadline when the incoming context has none.
func WithDefault(d time.Duration) Option {
	return func(o *options) { o.def = d }
}

// WithMax caps the effective deadline at d from now, shortening it if the
// caller set (or WithDefault applied) a longer one. It also applies when the
// caller set no deadline at all, unless WithDefault already set a shorter one.
func WithMax(d time.Duration) Option {
	return func(o *options) { o.max = d }
}

// UnaryServerInterceptor returns a unary server interceptor that enforces
// the configured default and/or maximum deadline on the request context.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	o := newOptions(opts...)
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		newCtx, cancel := apply(ctx, o)
		if cancel != nil {
			defer cancel()
		}
		return handler(newCtx, req)
	}
}

// StreamServerInterceptor returns a stream server interceptor that enforces
// the configured default and/or maximum deadline on the stream context.
func StreamServerInterceptor(opts ...Option) grpc.StreamServerInterceptor {
	o := newOptions(opts...)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx, cancel := apply(ss.Context(), o)
		if cancel != nil {
			defer cancel()
		}
		return handler(srv, &contextServerStream{ServerStream: ss, ctx: newCtx})
	}
}

func newOptions(opts ...Option) *options {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func apply(ctx context.Context, o *options) (context.Context, context.CancelFunc) {
	deadline, hasDeadline := ctx.Deadline()
	if !hasDeadline {
		d := o.def
		if o.max > 0 && (d == 0 || o.max < d) {
			d = o.max
		}
		if d > 0 {
			return context.WithTimeout(ctx, d)
		}
		return ctx, nil
	}
	if o.max > 0 {
		if maxDeadline := time.Now().Add(o.max); deadline.After(maxDeadline) {
			return context.WithDeadline(ctx, maxDeadline)
		}
	}
	return ctx, nil
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context { return s.ctx }
