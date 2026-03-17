package server

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

type testServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *testServerStream) Context() context.Context {
	return s.ctx
}

func TestIncomingMetadataReturnsCopy(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
		"authorization", "Bearer token",
	))

	md := IncomingMetadata(ctx)
	if got, want := md.Get("x-request-id"), []string{"req-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}

	md.Set("x-request-id", "req-2")
	if got, want := IncomingMetadataValue(ctx, "x-request-id"), []string{"req-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
}

func TestIncomingMetadataValueReturnsNilWithoutMetadata(t *testing.T) {
	if got := IncomingMetadataValue(context.Background(), "x-request-id"); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestNewContextServerStreamUpdatesExistingWrapperContext(t *testing.T) {
	ctx1 := context.WithValue(context.Background(), contextKey("step"), "one")
	ctx2 := context.WithValue(context.Background(), contextKey("step"), "two")

	wrapped := NewContextServerStream(ctx1, &testServerStream{ctx: context.Background()})
	updated := NewContextServerStream(ctx2, wrapped)

	if wrapped != updated {
		t.Fatal("expected existing wrapped stream to be reused")
	}
	if got := updated.Context().Value(contextKey("step")); got != "two" {
		t.Fatalf("got context value=%v, want %v", got, "two")
	}
}

func TestContextUnaryInterceptorPassesModifiedContext(t *testing.T) {
	interceptor := ContextUnaryInterceptor(func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, contextKey("request-id"), "req-1"), nil
	})

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		if got := ctx.Value(contextKey("request-id")); got != "req-1" {
			t.Fatalf("got context value=%v, want %v", got, "req-1")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestContextStreamInterceptorPassesModifiedContext(t *testing.T) {
	interceptor := ContextStreamInterceptor(func(ctx context.Context) (context.Context, error) {
		return context.WithValue(ctx, contextKey("request-id"), "req-1"), nil
	})

	err := interceptor(nil, &testServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		if got := stream.Context().Value(contextKey("request-id")); got != "req-1" {
			t.Fatalf("got context value=%v, want %v", got, "req-1")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeferContextUnaryInterceptorRunsAfterHandler(t *testing.T) {
	var (
		called     bool
		handlerRan bool
	)
	interceptor := DeferContextUnaryInterceptor(func(ctx context.Context) {
		called = true
		if !handlerRan {
			t.Fatal("expected handler to run before deferred callback")
		}
	})

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		handlerRan = true
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected deferred callback to run")
	}
}

func TestDeferContextStreamInterceptorRunsAfterHandler(t *testing.T) {
	var (
		called     bool
		handlerRan bool
	)
	interceptor := DeferContextStreamInterceptor(func(ctx context.Context) {
		called = true
		if !handlerRan {
			t.Fatal("expected handler to run before deferred callback")
		}
	})

	err := interceptor(nil, &testServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		handlerRan = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected deferred callback to run")
	}
}

func TestContextUnaryInterceptorPreservesWrappedErrors(t *testing.T) {
	want := errors.New("boom")
	interceptor := ContextUnaryInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, want
	})

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, func(ctx context.Context, req any) (any, error) {
		return nil, nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}

func TestContextStreamInterceptorPreservesWrappedErrors(t *testing.T) {
	want := errors.New("boom")
	interceptor := ContextStreamInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, want
	})

	err := interceptor(nil, &testServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{}, func(srv any, stream grpc.ServerStream) error {
		return nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}
