package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/circuitbreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptorTripsThresholdBreaker(t *testing.T) {
	breaker := circuitbreaker.NewThresholdBreaker(2, time.Hour)
	interceptor := circuitbreaker.UnaryClientInterceptor(breaker)
	want := status.Error(codes.Unavailable, "boom")

	for range 2 {
		err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return want
		})
		if status.Code(err) != codes.Unavailable {
			t.Fatalf("got err=%v, want unavailable", err)
		}
	}

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		t.Fatal("invoker should not run while breaker is open")
		return nil
	})
	if !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("got err=%v, want %v", err, circuitbreaker.ErrOpen)
	}
}

func TestThresholdBreakerIgnoresNonTripErrorsByDefault(t *testing.T) {
	breaker := circuitbreaker.NewThresholdBreaker(1, time.Hour)
	interceptor := circuitbreaker.UnaryClientInterceptor(breaker)
	want := status.Error(codes.InvalidArgument, "bad request")

	for range 2 {
		err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return want
		})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("got err=%v, want invalid argument", err)
		}
	}
}

func TestThresholdBreakerTripsConfiguredCodes(t *testing.T) {
	breaker := circuitbreaker.NewThresholdBreaker(1, time.Hour, circuitbreaker.WithTripOnCodes(codes.DeadlineExceeded))
	interceptor := circuitbreaker.UnaryClientInterceptor(breaker)

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return status.Error(codes.DeadlineExceeded, "timeout")
	})
	if status.Code(err) != codes.DeadlineExceeded {
		t.Fatalf("got err=%v, want deadline exceeded", err)
	}

	err = interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		t.Fatal("invoker should not run while breaker is open")
		return nil
	})
	if !errors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("got err=%v, want %v", err, circuitbreaker.ErrOpen)
	}
}
