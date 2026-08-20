package tracing_test

import (
	"context"
	"testing"
	"time"

	nanotracing "github.com/pthethanh/nano/grpc/interceptor/tracing"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
)

// TestStreamClientInterceptorEndsSpanWhenStreamIsAbandoned proves that a
// stream that's simply dropped (its context cancelled, without ever hitting
// a terminal error on Header/SendMsg/RecvMsg/CloseSend) still gets its span
// ended, instead of leaking one span per abandoned stream forever.
func TestStreamClientInterceptorEndsSpanWhenStreamIsAbandoned(t *testing.T) {
	tp := newTracerProvider()
	interceptor := nanotracing.StreamClientInterceptor(
		nanotracing.WithTracerProvider(tp),
		nanotracing.WithPropagator(propagation.TraceContext{}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := interceptor(ctx, &grpc.StreamDesc{ClientStreams: true, ServerStreams: true}, nil, "/svc/method", func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return &testClientStream{}, nil // recvErr is nil: no method ever returns an error
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = stream

	if got := len(tp.Ended()); got != 0 {
		t.Fatalf("got %d ended spans before cancellation, want 0", got)
	}

	// Simulate the caller giving up on the stream (e.g. its own deferred
	// cancel() firing) without ever seeing a terminal error.
	cancel()

	deadline := time.Now().Add(time.Second)
	for len(tp.Ended()) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("span was never ended after the stream's context was cancelled: an abandoned stream leaks its span")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
