package recovery

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDefaultHandler_DoesNotLeakStackToClient(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req any) (any, error) {
		panic("credentials: aws-secret-XYZ")
	}

	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, handler)
	if err == nil {
		t.Fatal("expected an error after panic recovery, got nil")
	}

	msg := err.Error()
	if strings.Contains(msg, "aws-secret-XYZ") {
		t.Errorf("recovered error leaked the panic value to the client: %q", msg)
	}
	if strings.Contains(msg, ".go:") || strings.Contains(msg, "goroutine ") {
		t.Errorf("recovered error leaked a stack trace to the client: %q", msg)
	}

	st := status.Convert(err)
	if st.Code() != codes.Internal {
		t.Errorf("recovered error code = %v, want %v", st.Code(), codes.Internal)
	}
}

func TestDefaultHandler_StreamInterceptorDoesNotLeakStackToClient(t *testing.T) {
	interceptor := StreamServerInterceptor()
	handler := func(srv any, stream grpc.ServerStream) error {
		panic("credentials: aws-secret-XYZ")
	}

	err := interceptor(nil, &fakeServerStream{ctx: context.Background()}, &grpc.StreamServerInfo{FullMethod: "/svc/Method"}, handler)
	if err == nil {
		t.Fatal("expected an error after panic recovery, got nil")
	}
	if strings.Contains(err.Error(), "aws-secret-XYZ") {
		t.Errorf("recovered error leaked the panic value to the client: %q", err.Error())
	}
}

func TestDefaultHandler_LogsPanicAndStackServerSide(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	interceptor := UnaryServerInterceptor()
	handler := func(ctx context.Context, req any) (any, error) {
		panic("boom-marker")
	}
	_, _ = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, handler)

	logged := buf.String()
	if !strings.Contains(logged, "boom-marker") {
		t.Errorf("expected the panic value to be logged server-side, got log output: %q", logged)
	}
}

func TestWithHandler_CustomHandlerStillHonored(t *testing.T) {
	called := false
	custom := WithHandler(func(ctx context.Context, p any) error {
		called = true
		return status.Error(codes.Unavailable, "custom")
	})
	interceptor := UnaryServerInterceptor(custom)
	handler := func(ctx context.Context, req any) (any, error) {
		panic("x")
	}
	_, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, handler)
	if !called {
		t.Fatal("custom handler was not invoked")
	}
	if status.Convert(err).Code() != codes.Unavailable {
		t.Errorf("got code %v, want %v", status.Convert(err).Code(), codes.Unavailable)
	}
}

type fakeServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *fakeServerStream) Context() context.Context { return s.ctx }
