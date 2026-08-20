package concurrencylimit_test

import (
	"context"
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/concurrencylimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSemaphore_AcquireTimeoutCarriesAGRPCStatusCode(t *testing.T) {
	sem := concurrencylimit.NewSemaphore(1)
	if err := sem.Acquire(context.Background()); err != nil {
		t.Fatalf("first Acquire() error = %v", err)
	}
	// Slot is full; a context that's already done forces the ctx.Done() branch.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sem.Acquire(ctx)
	if err == nil {
		t.Fatal("expected an error when the semaphore is full and ctx is done")
	}
	got := status.Code(err)
	if got != codes.Canceled {
		t.Errorf("status.Code(err) = %v, want %v: raw ctx.Err() surfaces to clients as codes.Unknown instead of a code that reflects what actually happened", got, codes.Canceled)
	}
}
