package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/retry"
	"google.golang.org/grpc"
	grpcbackoff "google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUnaryClientInterceptorRetriesRetryableErrors(t *testing.T) {
	var attempts int
	interceptor := retry.UnaryClientInterceptor(
		retry.WithMaxAttempts(3),
		retry.WithRetryableCodes(codes.Unavailable),
		retry.WithBackoff(func(int) time.Duration { return 0 }),
	)
	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "try again")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("got %d attempts, want 3", attempts)
	}
}

func TestUnaryClientInterceptorStopsOnNonRetryableError(t *testing.T) {
	var attempts int
	interceptor := retry.UnaryClientInterceptor(
		retry.WithMaxAttempts(3),
		retry.WithRetryableCodes(codes.Unavailable),
		retry.WithBackoff(func(int) time.Duration { return 0 }),
	)
	want := status.Error(codes.InvalidArgument, "bad request")
	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		return want
	})
	if !errors.Is(err, want) && status.Code(err) != codes.InvalidArgument {
		t.Fatalf("got err=%v, want invalid argument", err)
	}
	if attempts != 1 {
		t.Fatalf("got %d attempts, want 1", attempts)
	}
}

func TestUnaryClientInterceptorDoesNotRetryByDefault(t *testing.T) {
	var attempts int
	interceptor := retry.UnaryClientInterceptor(
		retry.WithMaxAttempts(3),
		retry.WithBackoff(func(int) time.Duration { return 0 }),
	)

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		return status.Error(codes.Unavailable, "try again")
	})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("got err=%v, want unavailable", err)
	}
	if attempts != 1 {
		t.Fatalf("got %d attempts, want 1", attempts)
	}
}

func TestBackoffUsesGRPCConfigSemantics(t *testing.T) {
	backoff := retry.Backoff(grpcbackoff.Config{
		BaseDelay:  100 * time.Millisecond,
		Multiplier: 2,
		Jitter:     0,
		MaxDelay:   800 * time.Millisecond,
	})

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: 100 * time.Millisecond},
		{attempt: 2, want: 200 * time.Millisecond},
		{attempt: 3, want: 400 * time.Millisecond},
		{attempt: 4, want: 800 * time.Millisecond},
		{attempt: 5, want: 800 * time.Millisecond},
	}

	for _, tt := range tests {
		if got := backoff(tt.attempt); got != tt.want {
			t.Fatalf("attempt %d: got %v, want %v", tt.attempt, got, tt.want)
		}
	}
}

func TestWithBackoffConfigRetries(t *testing.T) {
	var attempts int
	interceptor := retry.UnaryClientInterceptor(
		retry.WithMaxAttempts(3),
		retry.WithBackoffConfig(grpcbackoff.Config{
			BaseDelay:  time.Nanosecond,
			Multiplier: 2,
			Jitter:     0,
			MaxDelay:   time.Nanosecond,
		}),
		retry.WithRetryableCodes(codes.Unavailable),
	)

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "try again")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("got %d attempts, want 3", attempts)
	}
}
