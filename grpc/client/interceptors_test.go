package client_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/pthethanh/nano/grpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestOutgoingMetadataSetsAndOverridesOutgoingValues(t *testing.T) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-old",
		"x-tenant-id", "tenant-1",
	))

	ctx = client.OutgoingMetadata(ctx, "X-Request-Id", "req-new", "authorization", "Bearer token")

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got, want := outgoing.Get("x-request-id"), []string{"req-new"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("authorization"), []string{"Bearer token"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got authorization=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("x-tenant-id"), []string{"tenant-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-tenant-id=%v, want %v", got, want)
	}
}

func TestOutgoingMetadataReturnsSameContextWithoutPairs(t *testing.T) {
	ctx := context.Background()
	if got := client.OutgoingMetadata(ctx); got != ctx {
		t.Fatal("expected original context")
	}
}

func TestAppendOutgoingMetadataAppendsOutgoingValues(t *testing.T) {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
	))

	ctx = client.AppendOutgoingMetadata(ctx, "X-Request-Id", "req-2", "authorization", "Bearer token")

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got, want := outgoing.Get("x-request-id"), []string{"req-1", "req-2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("authorization"), []string{"Bearer token"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got authorization=%v, want %v", got, want)
	}
}

func TestAppendOutgoingMetadataReturnsSameContextWithoutPairs(t *testing.T) {
	ctx := context.Background()
	if got := client.AppendOutgoingMetadata(ctx); got != ctx {
		t.Fatal("expected original context")
	}
}

func TestForwardIncomingMetadataCopiesAllWhenNoKeysProvided(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
		"authorization", "Bearer token",
	))

	ctx = client.ForwardIncomingMetadata(ctx)

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got, want := outgoing.Get("x-request-id"), []string{"req-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("authorization"), []string{"Bearer token"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got authorization=%v, want %v", got, want)
	}
}

func TestForwardIncomingMetadataReturnsSameContextWithoutMatchingKeys(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
	))
	if got := client.ForwardIncomingMetadata(ctx, "authorization"); got != ctx {
		t.Fatal("expected original context")
	}
}

func TestForwardIncomingMetadataCopiesOnlySelectedKeys(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
		"authorization", "Bearer token",
	))

	ctx = client.ForwardIncomingMetadata(ctx, "X-Request-Id")

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got, want := outgoing.Get("x-request-id"), []string{"req-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
	if got := outgoing.Get("authorization"); len(got) != 0 {
		t.Fatalf("got authorization=%v, want none", got)
	}
}

func TestForwardIncomingMetadataOverridesMatchingOutgoingKeys(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-incoming",
		"authorization", "Bearer incoming",
	))
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(
		"x-request-id", "req-outgoing",
		"x-tenant-id", "tenant-1",
	))

	ctx = client.ForwardIncomingMetadata(ctx, "x-request-id", "authorization")

	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}
	if got, want := outgoing.Get("x-request-id"), []string{"req-incoming"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-request-id=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("authorization"), []string{"Bearer incoming"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got authorization=%v, want %v", got, want)
	}
	if got, want := outgoing.Get("x-tenant-id"), []string{"tenant-1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got x-tenant-id=%v, want %v", got, want)
	}
}

func TestForwardIncomingMetadataReturnsSameContextWithoutIncomingMetadata(t *testing.T) {
	ctx := context.Background()
	if got := client.ForwardIncomingMetadata(ctx, "x-request-id"); got != ctx {
		t.Fatal("expected original context")
	}
}

func TestForwardMetadataUnaryInterceptor(t *testing.T) {
	interceptor := client.ForwardMetadataUnaryInterceptor("x-request-id")
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
		"authorization", "Bearer token",
	))

	err := interceptor(ctx, "/svc/method", nil, nil, nil, func(callCtx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		outgoing, ok := metadata.FromOutgoingContext(callCtx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		if got, want := outgoing.Get("x-request-id"), []string{"req-1"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got x-request-id=%v, want %v", got, want)
		}
		if got := outgoing.Get("authorization"); len(got) != 0 {
			t.Fatalf("got authorization=%v, want none", got)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestForwardMetadataStreamInterceptor(t *testing.T) {
	interceptor := client.ForwardMetadataStreamInterceptor("authorization")
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-request-id", "req-1",
		"authorization", "Bearer token",
	))

	stream, err := interceptor(ctx, &grpc.StreamDesc{}, nil, "/svc/method", func(callCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		outgoing, ok := metadata.FromOutgoingContext(callCtx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		if got, want := outgoing.Get("authorization"), []string{"Bearer token"}; !reflect.DeepEqual(got, want) {
			t.Fatalf("got authorization=%v, want %v", got, want)
		}
		if got := outgoing.Get("x-request-id"); len(got) != 0 {
			t.Fatalf("got x-request-id=%v, want none", got)
		}
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if stream != nil {
		t.Fatal("expected nil client stream")
	}
}

func TestForwardMetadataUnaryInterceptorPreservesWrappedErrors(t *testing.T) {
	want := errors.New("boom")
	interceptor := client.ForwardMetadataUnaryInterceptor("x-request-id")

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(callCtx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}

func TestContextUnaryInterceptorPreservesWrappedErrors(t *testing.T) {
	want := errors.New("boom")
	interceptor := client.ContextUnaryInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, want
	})

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(callCtx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
}

func TestContextStreamInterceptorPreservesWrappedErrors(t *testing.T) {
	want := errors.New("boom")
	interceptor := client.ContextStreamInterceptor(func(ctx context.Context) (context.Context, error) {
		return nil, want
	})

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/svc/method", func(callCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, nil
	})
	if !errors.Is(err, want) {
		t.Fatalf("got err=%v, want %v", err, want)
	}
	if stream != nil {
		t.Fatal("expected nil client stream")
	}
}

func TestDeferContextUnaryInterceptorRunsAfterInvoker(t *testing.T) {
	var (
		called     bool
		invokerRan bool
	)
	interceptor := client.DeferContextUnaryInterceptor(func(ctx context.Context) {
		called = true
		if !invokerRan {
			t.Fatal("expected invoker to run before deferred callback")
		}
	})

	err := interceptor(context.Background(), "/svc/method", nil, nil, nil, func(callCtx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		invokerRan = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("expected deferred callback to run")
	}
}

func TestDeferContextStreamInterceptorRunsAfterStreamer(t *testing.T) {
	var (
		called      bool
		streamerRan bool
	)
	interceptor := client.DeferContextStreamInterceptor(func(ctx context.Context) {
		called = true
		if !streamerRan {
			t.Fatal("expected streamer to run before deferred callback")
		}
	})

	stream, err := interceptor(context.Background(), &grpc.StreamDesc{}, nil, "/svc/method", func(callCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		streamerRan = true
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if stream != nil {
		t.Fatal("expected nil client stream")
	}
	if !called {
		t.Fatal("expected deferred callback to run")
	}
}
