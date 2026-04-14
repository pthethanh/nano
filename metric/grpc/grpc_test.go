package grpc_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	grpcmetric "github.com/pthethanh/nano/metric/grpc"
	"github.com/pthethanh/nano/metric/memory"
	gogrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryServerInterceptorRecordsMetrics(t *testing.T) {
	reporter := memory.New()
	interceptor := grpcmetric.UnaryServerInterceptor(reporter)
	_, err := interceptor(context.Background(), nil, &gogrpc.UnaryServerInfo{FullMethod: "/svc/method"}, func(ctx context.Context, req any) (any, error) {
		return nil, status.Error(codes.NotFound, "missing")
	})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("got err=%v, want not found", err)
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	reporter.ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "grpc_requests_total") {
		t.Fatalf("expected grpc_requests_total in metrics output, got: %s", body)
	}
	if !strings.Contains(body, "grpc_request_duration_seconds") {
		t.Fatalf("expected grpc_request_duration_seconds in metrics output, got: %s", body)
	}
}

func TestStreamClientInterceptorRecordsMetricsOnStreamCompletion(t *testing.T) {
	reporter := memory.New()
	interceptor := grpcmetric.StreamClientInterceptor(reporter)
	stream := &testClientStream{recvErr: io.EOF}

	got, err := interceptor(context.Background(), &gogrpc.StreamDesc{ClientStreams: true, ServerStreams: true}, nil, "/svc/method", func(ctx context.Context, desc *gogrpc.StreamDesc, cc *gogrpc.ClientConn, method string, opts ...gogrpc.CallOption) (gogrpc.ClientStream, error) {
		return stream, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	before := scrapeMetrics(t, reporter)
	if strings.Contains(before, "grpc_client_requests_total") {
		t.Fatalf("did not expect client stream metrics before stream completion, got: %s", before)
	}

	if err := got.RecvMsg(nil); err != io.EOF {
		t.Fatalf("got err=%v, want EOF", err)
	}

	after := scrapeMetrics(t, reporter)
	if !strings.Contains(after, "grpc_client_requests_total") {
		t.Fatalf("expected grpc_client_requests_total in metrics output, got: %s", after)
	}
	if !strings.Contains(after, "grpc_client_request_duration_seconds") {
		t.Fatalf("expected grpc_client_request_duration_seconds in metrics output, got: %s", after)
	}
}

func scrapeMetrics(t *testing.T, reporter *memory.Reporter) string {
	t.Helper()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	reporter.ServeHTTP(rec, req)
	return rec.Body.String()
}

type testClientStream struct {
	recvErr error
}

func (s *testClientStream) Header() (metadata.MD, error) { return nil, nil }
func (s *testClientStream) Trailer() metadata.MD         { return nil }
func (s *testClientStream) CloseSend() error             { return nil }
func (s *testClientStream) Context() context.Context     { return context.Background() }
func (s *testClientStream) SendMsg(any) error            { return nil }
func (s *testClientStream) RecvMsg(any) error            { return s.recvErr }
