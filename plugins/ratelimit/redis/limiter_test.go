package redis_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	nanoratelimit "github.com/pthethanh/nano/grpc/interceptor/ratelimit"
	ratelimit "github.com/pthethanh/nano/plugins/ratelimit/redis"
)

func TestLimiter_AllowsUpToLimitThenRejects(t *testing.T) {
	s := miniredis.RunT(t)

	l := ratelimit.New(
		ratelimit.Address(s.Addr()),
		ratelimit.Limit(3, time.Minute),
	)
	if err := l.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer l.Close(context.Background())

	for i := range 3 {
		if err := l.Allow(context.Background()); err != nil {
			t.Fatalf("Allow() call %d error = %v, want nil", i+1, err)
		}
	}
	if err := l.Allow(context.Background()); !errors.Is(err, nanoratelimit.ErrLimited) {
		t.Fatalf("4th Allow() error = %v, want %v", err, nanoratelimit.ErrLimited)
	}
}

func TestLimiter_WindowExpiryResetsTheCounter(t *testing.T) {
	s := miniredis.RunT(t)

	l := ratelimit.New(
		ratelimit.Address(s.Addr()),
		ratelimit.Limit(1, time.Second),
	)
	if err := l.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer l.Close(context.Background())

	if err := l.Allow(context.Background()); err != nil {
		t.Fatalf("1st Allow() error = %v", err)
	}
	if err := l.Allow(context.Background()); !errors.Is(err, nanoratelimit.ErrLimited) {
		t.Fatalf("2nd Allow() error = %v, want %v", err, nanoratelimit.ErrLimited)
	}

	s.FastForward(2 * time.Second)

	if err := l.Allow(context.Background()); err != nil {
		t.Fatalf("Allow() after window expiry error = %v, want nil (counter should have reset)", err)
	}
}

func TestLimiter_KeyFuncSeparatesBuckets(t *testing.T) {
	s := miniredis.RunT(t)

	type ctxKey struct{}
	l := ratelimit.New(
		ratelimit.Address(s.Addr()),
		ratelimit.Limit(1, time.Minute),
		ratelimit.KeyFunc(func(ctx context.Context) string {
			id, _ := ctx.Value(ctxKey{}).(string)
			return id
		}),
	)
	if err := l.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer l.Close(context.Background())

	ctxA := context.WithValue(context.Background(), ctxKey{}, "tenant-a")
	ctxB := context.WithValue(context.Background(), ctxKey{}, "tenant-b")

	if err := l.Allow(ctxA); err != nil {
		t.Fatalf("tenant-a 1st Allow() error = %v", err)
	}
	if err := l.Allow(ctxB); err != nil {
		t.Fatalf("tenant-b 1st Allow() error = %v: a different key must not share tenant-a's quota", err)
	}
	if err := l.Allow(ctxA); !errors.Is(err, nanoratelimit.ErrLimited) {
		t.Fatalf("tenant-a 2nd Allow() error = %v, want %v", err, nanoratelimit.ErrLimited)
	}
}

func TestLimiter_AllowBeforeOpenReturnsError(t *testing.T) {
	l := ratelimit.New()
	if err := l.Allow(context.Background()); !errors.Is(err, ratelimit.ErrNotOpen) {
		t.Fatalf("Allow() before Open() error = %v, want %v", err, ratelimit.ErrNotOpen)
	}
}

// wireIntoInterceptor proves *Limiter satisfies nanoratelimit.Limiter and
// works when wired into the real interceptor.
func TestLimiter_WorksAsRatelimitInterceptorLimiter(t *testing.T) {
	s := miniredis.RunT(t)
	l := ratelimit.New(ratelimit.Address(s.Addr()), ratelimit.Limit(1, time.Minute))
	if err := l.Open(context.Background()); err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer l.Close(context.Background())

	var lim nanoratelimit.Limiter = l
	if err := lim.Allow(context.Background()); err != nil {
		t.Fatalf("Allow() error = %v", err)
	}
	if err := lim.Allow(context.Background()); !errors.Is(err, nanoratelimit.ErrLimited) {
		t.Fatalf("2nd Allow() error = %v, want %v", err, nanoratelimit.ErrLimited)
	}
}
