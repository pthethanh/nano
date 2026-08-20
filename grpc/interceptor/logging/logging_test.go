package logging

import (
	"context"
	"log/slog"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type capturingLogger struct {
	calls []call
}

type call struct {
	msg   string
	attrs []any
}

func (l *capturingLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.calls = append(l.calls, call{msg: msg, attrs: args})
}

func (l *capturingLogger) hasKey(key string) bool {
	for _, c := range l.calls {
		for i := 0; i+1 < len(c.attrs); i += 2 {
			if k, ok := c.attrs[i].(string); ok && k == key {
				return true
			}
		}
	}
	return false
}

func TestLogResponse_DefaultOptionsDoNotLogResponsePayload(t *testing.T) {
	logger := &capturingLogger{}
	o := newOpts() // no Response() option: logResponse defaults to false

	logResponse(logger, context.Background(), "sent grpc response", o, "/svc/Method", "SENSITIVE_PAYLOAD", nil, 0)

	if logger.hasKey("grpc.response") {
		t.Fatalf("logResponse() logged grpc.response with default options, want it withheld unless Response() is enabled")
	}
}

func TestLogResponse_ResponseOptionEnablesPayloadLogging(t *testing.T) {
	logger := &capturingLogger{}
	o := newOpts(Response())

	logResponse(logger, context.Background(), "sent grpc response", o, "/svc/Method", "PAYLOAD", nil, 0)

	if !logger.hasKey("grpc.response") {
		t.Fatalf("logResponse() did not log grpc.response even though Response() was enabled")
	}
}

func TestUnaryServerInterceptor_DefaultOptionsDoNotLeakResponse(t *testing.T) {
	logger := &capturingLogger{}
	interceptor := UnaryServerInterceptor(logger)

	handler := func(ctx context.Context, req any) (any, error) {
		return "SENSITIVE_PAYLOAD", nil
	}
	_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/svc/Method"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logger.hasKey("grpc.response") {
		t.Fatalf("UnaryServerInterceptor() with default options logged grpc.response; response logging must be opt-in via Response()")
	}
}

func TestClientContextLogger_HeaderLookupIsCaseInsensitive(t *testing.T) {
	fn := ClientContextLogger(func(ctx context.Context, attrs ...any) context.Context {
		return context.WithValue(ctx, attrs[0], attrs[1])
	}, map[string]any{"X-Request-Id": ""})

	md := metadata.Pairs("x-request-id", "abc-123") // gRPC metadata is always stored lowercase
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	newCtx, err := fn(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := newCtx.Value("X-Request-Id").(string)
	if got != "abc-123" {
		t.Errorf("ClientContextLogger() lookup for %q = %q, want %q (lookup must be case-insensitive like ServerContextLogger)", "X-Request-Id", got, "abc-123")
	}
}
