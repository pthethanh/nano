package ratelimit_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/ratelimit"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptorRejectsWhenLimited(t *testing.T) {
	want := errors.New("limited")
	interceptor := ratelimit.UnaryServerInterceptor(ratelimit.AllowFunc(func(context.Context) error {
		return want
	}))

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not run")
		return nil, nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}

func TestTokenBucketAllowsBurstThenRejects(t *testing.T) {
	limiter := ratelimit.NewTokenBucket(rate.Every(time.Hour), 2)

	if err := limiter.Allow(context.Background()); err != nil {
		t.Fatalf("first allow failed: %v", err)
	}
	if err := limiter.Allow(context.Background()); err != nil {
		t.Fatalf("second allow failed: %v", err)
	}
	if err := limiter.Allow(context.Background()); !errors.Is(err, ratelimit.ErrLimited) {
		t.Fatalf("got err=%v, want %v", err, ratelimit.ErrLimited)
	}
}

func TestUnaryServerInterceptorUsesTokenBucket(t *testing.T) {
	interceptor := ratelimit.UnaryServerInterceptor(ratelimit.NewTokenBucket(rate.Every(time.Hour), 1))

	if _, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		return "ok", nil
	}); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		t.Fatal("handler should not run after burst is exhausted")
		return nil, nil
	})
	if !errors.Is(err, ratelimit.ErrLimited) {
		t.Fatalf("got err=%v, want %v", err, ratelimit.ErrLimited)
	}
}
