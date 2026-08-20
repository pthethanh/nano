package grpc_test

import (
	"context"
	"strings"
	"testing"
	"time"

	grpcmetric "github.com/pthethanh/nano/metric/grpc"
	"github.com/pthethanh/nano/metric/memory"
	gogrpc "google.golang.org/grpc"
)

// TestStreamClientInterceptorRecordsMetricsWhenStreamIsAbandoned proves that
// a stream that's simply dropped (its context cancelled, without ever
// hitting a terminal error on Header/SendMsg/RecvMsg/CloseSend) still gets
// its metrics recorded, instead of leaking one un-recorded stream forever.
func TestStreamClientInterceptorRecordsMetricsWhenStreamIsAbandoned(t *testing.T) {
	reporter := memory.New()
	interceptor := grpcmetric.StreamClientInterceptor(reporter)

	ctx, cancel := context.WithCancel(context.Background())
	stream, err := interceptor(ctx, &gogrpc.StreamDesc{ClientStreams: true, ServerStreams: true}, nil, "/svc/method", func(ctx context.Context, desc *gogrpc.StreamDesc, cc *gogrpc.ClientConn, method string, opts ...gogrpc.CallOption) (gogrpc.ClientStream, error) {
		return &testClientStream{}, nil // recvErr is nil: no method ever returns an error
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = stream

	before := scrapeMetrics(t, reporter)
	if strings.Contains(before, "grpc_client_requests_total") {
		t.Fatalf("did not expect client stream metrics before cancellation, got: %s", before)
	}

	// Simulate the caller giving up on the stream (e.g. its own deferred
	// cancel() firing) without ever seeing a terminal error.
	cancel()

	deadline := time.Now().Add(time.Second)
	for {
		after := scrapeMetrics(t, reporter)
		if strings.Contains(after, "grpc_client_requests_total") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("metrics were never recorded after the stream's context was cancelled: an abandoned stream leaks its metric")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
