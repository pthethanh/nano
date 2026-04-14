package circuitbreaker

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

var ErrOpen = errors.New("circuit breaker is open")

type options struct {
	shouldTrip func(error) bool
}

// Option customizes ThresholdBreaker behavior.
type Option func(*options)

// WithShouldTrip overrides which errors count toward opening the breaker.
func WithShouldTrip(fn func(error) bool) Option {
	return func(o *options) {
		if fn != nil {
			o.shouldTrip = fn
		}
	}
}

// WithTripOnCodes marks only the provided gRPC codes as breaker-worthy failures.
func WithTripOnCodes(tripCodes ...codes.Code) Option {
	return WithShouldTrip(TripOnCodes(tripCodes...))
}

// Breaker is the minimal interface required by the interceptors.
type Breaker interface {
	Allow() error
	Success()
	Failure(error)
}

// UnaryClientInterceptor returns a client interceptor guarded by the provided breaker.
func UnaryClientInterceptor(b Breaker) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if err := b.Allow(); err != nil {
			return err
		}
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			b.Failure(err)
			return err
		}
		b.Success()
		return nil
	}
}

// StreamClientInterceptor returns a stream client interceptor guarded by the provided breaker.
func StreamClientInterceptor(b Breaker) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if err := b.Allow(); err != nil {
			return nil, err
		}
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			b.Failure(err)
			return nil, err
		}
		b.Success()
		return stream, nil
	}
}

// TripOnCodes returns a classifier that counts only the provided gRPC codes.
func TripOnCodes(tripCodes ...codes.Code) func(error) bool {
	allowed := make(map[codes.Code]struct{}, len(tripCodes))
	for _, code := range tripCodes {
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

func defaultShouldTrip(err error) bool {
	return TripOnCodes(codes.Unavailable, codes.ResourceExhausted)(err)
}
