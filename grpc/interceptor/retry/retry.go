package retry

import (
	"context"
	"time"

	"google.golang.org/grpc"
	grpcbackoff "google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type options struct {
	maxAttempts int
	backoff     func(attempt int) time.Duration
	shouldRetry func(error) bool
}

// Option customizes retry behavior.
type Option func(*options)

// WithMaxAttempts sets the total number of attempts including the first call.
func WithMaxAttempts(attempts int) Option {
	return func(o *options) {
		if attempts > 0 {
			o.maxAttempts = attempts
		}
	}
}

// WithBackoff sets the backoff between attempts.
func WithBackoff(backoff func(attempt int) time.Duration) Option {
	return func(o *options) {
		if backoff != nil {
			o.backoff = backoff
		}
	}
}

// WithBackoffConfig sets retry backoff behavior using gRPC backoff semantics.
func WithBackoffConfig(cfg grpcbackoff.Config) Option {
	return WithBackoff(Backoff(cfg))
}

// WithShouldRetry overrides the retry classifier.
func WithShouldRetry(fn func(error) bool) Option {
	return func(o *options) {
		if fn != nil {
			o.shouldRetry = fn
		}
	}
}

// WithRetryableCodes retries only when the returned gRPC status code matches
// one of the provided codes.
func WithRetryableCodes(retryableCodes ...codes.Code) Option {
	return WithShouldRetry(RetryableCodes(retryableCodes...))
}

// UnaryClientInterceptor returns a client interceptor that retries failed unary calls.
func UnaryClientInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	o := newOptions(opts...)
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		var err error
		for attempt := 1; attempt <= o.maxAttempts; attempt++ {
			err = invoker(ctx, method, req, reply, cc, callOpts...)
			if !o.shouldRetry(err) || attempt == o.maxAttempts {
				return err
			}
			if err := sleep(ctx, o.backoff(attempt)); err != nil {
				return err
			}
		}
		return err
	}
}

// StreamClientInterceptor returns a client interceptor that retries stream setup failures.
func StreamClientInterceptor(opts ...Option) grpc.StreamClientInterceptor {
	o := newOptions(opts...)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, callOpts ...grpc.CallOption) (grpc.ClientStream, error) {
		var (
			stream grpc.ClientStream
			err    error
		)
		for attempt := 1; attempt <= o.maxAttempts; attempt++ {
			stream, err = streamer(ctx, desc, cc, method, callOpts...)
			if !o.shouldRetry(err) || attempt == o.maxAttempts {
				return stream, err
			}
			if err := sleep(ctx, o.backoff(attempt)); err != nil {
				return nil, err
			}
		}
		return stream, err
	}
}

func newOptions(opts ...Option) *options {
	o := &options{
		maxAttempts: 3,
		backoff:     defaultBackoff,
		shouldRetry: defaultShouldRetry,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func defaultBackoff(attempt int) time.Duration {
	return Backoff(grpcbackoff.DefaultConfig)(attempt)
}

// RetryableCodes returns a classifier that retries only the provided gRPC codes.
func RetryableCodes(retryableCodes ...codes.Code) func(error) bool {
	allowed := make(map[codes.Code]struct{}, len(retryableCodes))
	for _, code := range retryableCodes {
		allowed[code] = struct{}{}
	}
	return func(err error) bool {
		if err == nil || len(allowed) == 0 {
			return false
		}
		_, ok := allowed[grpcstatus.Code(err)]
		return ok
	}
}

func defaultShouldRetry(err error) bool {
	return false
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
