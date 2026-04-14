package grpc

import (
	"context"
	"io"
	"time"

	"github.com/pthethanh/nano/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type options struct {
	prefix  string
	buckets []float64
}

// Option customizes metric names and histogram buckets.
type Option func(*options)

// WithPrefix sets the metric name prefix.
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// WithDurationBuckets overrides the histogram buckets used for duration metrics.
func WithDurationBuckets(buckets []float64) Option {
	return func(o *options) {
		if len(buckets) > 0 {
			o.buckets = append([]float64(nil), buckets...)
		}
	}
}

// UnaryServerInterceptor records request counts and durations for unary server calls.
func UnaryServerInterceptor(reporter metric.Reporter, opts ...Option) grpc.UnaryServerInterceptor {
	o := newOptions(opts...)
	counter := reporter.Counter(o.prefix+"requests_total", "method", "code", "kind")
	histogram := reporter.Histogram(o.prefix+"request_duration_seconds", o.buckets, "method", "code", "kind")

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := grpcCode(err).String()
		counter.With("method", info.FullMethod, "code", code, "kind", "unary").Add(1)
		histogram.With("method", info.FullMethod, "code", code, "kind", "unary").Record(time.Since(start).Seconds())
		return resp, err
	}
}

// StreamServerInterceptor records accepted stream counts and setup durations.
func StreamServerInterceptor(reporter metric.Reporter, opts ...Option) grpc.StreamServerInterceptor {
	o := newOptions(opts...)
	counter := reporter.Counter(o.prefix+"requests_total", "method", "code", "kind")
	histogram := reporter.Histogram(o.prefix+"request_duration_seconds", o.buckets, "method", "code", "kind")

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, stream)
		code := grpcCode(err).String()
		counter.With("method", info.FullMethod, "code", code, "kind", "stream").Add(1)
		histogram.With("method", info.FullMethod, "code", code, "kind", "stream").Record(time.Since(start).Seconds())
		return err
	}
}

// UnaryClientInterceptor records request counts and durations for unary client calls.
func UnaryClientInterceptor(reporter metric.Reporter, opts ...Option) grpc.UnaryClientInterceptor {
	o := newOptions(opts...)
	counter := reporter.Counter(o.prefix+"client_requests_total", "method", "code", "kind")
	histogram := reporter.Histogram(o.prefix+"client_request_duration_seconds", o.buckets, "method", "code", "kind")

	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, callOpts...)
		code := grpcCode(err).String()
		counter.With("method", method, "code", code, "kind", "unary").Add(1)
		histogram.With("method", method, "code", code, "kind", "unary").Record(time.Since(start).Seconds())
		return err
	}
}

// StreamClientInterceptor records stream counts and durations for client stream
// lifecycles.
func StreamClientInterceptor(reporter metric.Reporter, opts ...Option) grpc.StreamClientInterceptor {
	o := newOptions(opts...)
	counter := reporter.Counter(o.prefix+"client_requests_total", "method", "code", "kind")
	histogram := reporter.Histogram(o.prefix+"client_request_duration_seconds", o.buckets, "method", "code", "kind")

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()
		stream, err := streamer(ctx, desc, cc, method, callOpts...)
		if err != nil {
			code := grpcCode(err).String()
			counter.With("method", method, "code", code, "kind", "stream").Add(1)
			histogram.With("method", method, "code", code, "kind", "stream").Record(time.Since(start).Seconds())
			return nil, err
		}
		return &clientStream{
			ClientStream: stream,
			method:       method,
			startedAt:    start,
			counter:      counter,
			histogram:    histogram,
		}, nil
	}
}

func newOptions(opts ...Option) *options {
	o := &options{
		prefix: "grpc_",
		buckets: []float64{
			0.005,
			0.01,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
			2.5,
			5,
			10,
		},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func grpcCode(err error) codes.Code {
	if err == nil || err == io.EOF {
		return codes.OK
	}
	return grpcstatus.Code(err)
}
