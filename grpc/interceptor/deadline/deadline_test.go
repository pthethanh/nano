package deadline_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/deadline"
	"google.golang.org/grpc"
)

func TestUnaryServerInterceptor_NoOptionsLeavesContextUnchanged(t *testing.T) {
	interceptor := deadline.UnaryServerInterceptor()
	var gotHasDeadline bool
	_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		_, gotHasDeadline = ctx.Deadline()
		return nil, nil
	})
	if gotHasDeadline {
		t.Error("expected no deadline with no options configured")
	}
}

func TestUnaryServerInterceptor_DefaultAppliedWhenCallerSetNone(t *testing.T) {
	interceptor := deadline.UnaryServerInterceptor(deadline.WithDefault(50 * time.Millisecond))
	_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		dl, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected a deadline to be set")
		}
		if remaining := time.Until(dl); remaining <= 0 || remaining > 50*time.Millisecond {
			t.Fatalf("remaining = %v, want (0, 50ms]", remaining)
		}
		return nil, nil
	})
}

func TestUnaryServerInterceptor_DefaultNotAppliedWhenCallerAlreadySetOne(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	want, _ := ctx.Deadline()

	interceptor := deadline.UnaryServerInterceptor(deadline.WithDefault(50 * time.Millisecond))
	_, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(want) {
			t.Fatalf("deadline was overridden: got %v, want unchanged %v", got, want)
		}
		return nil, nil
	})
}

func TestUnaryServerInterceptor_MaxShortensACallerDeadlineThatsTooLong(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	interceptor := deadline.UnaryServerInterceptor(deadline.WithMax(50 * time.Millisecond))
	_, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		dl, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected a deadline")
		}
		if remaining := time.Until(dl); remaining <= 0 || remaining > 50*time.Millisecond {
			t.Fatalf("remaining = %v, want (0, 50ms] (max should have shortened the 1h caller deadline)", remaining)
		}
		return nil, nil
	})
}

func TestUnaryServerInterceptor_MaxDoesNotLengthenAShorterCallerDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	want, _ := ctx.Deadline()

	interceptor := deadline.UnaryServerInterceptor(deadline.WithMax(time.Hour))
	_, _ = interceptor(ctx, nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		got, ok := ctx.Deadline()
		if !ok || !got.Equal(want) {
			t.Fatalf("deadline was changed: got %v, want unchanged %v", got, want)
		}
		return nil, nil
	})
}

func TestUnaryServerInterceptor_DefaultAndMaxTogetherPickTheSmaller(t *testing.T) {
	interceptor := deadline.UnaryServerInterceptor(
		deadline.WithDefault(time.Hour),
		deadline.WithMax(50*time.Millisecond),
	)
	_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		dl, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected a deadline")
		}
		if remaining := time.Until(dl); remaining <= 0 || remaining > 50*time.Millisecond {
			t.Fatalf("remaining = %v, want (0, 50ms]: WithMax should win over a longer WithDefault", remaining)
		}
		return nil, nil
	})
}

func TestStreamServerInterceptor_DefaultAppliedToStreamContext(t *testing.T) {
	interceptor := deadline.StreamServerInterceptor(deadline.WithDefault(50 * time.Millisecond))
	err := interceptor(nil, &fakeServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv any, ss grpc.ServerStream) error {
		if _, ok := ss.Context().Deadline(); !ok {
			t.Fatal("expected a deadline on the wrapped stream context")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakeServerStream) Context() context.Context { return s.ctx }
