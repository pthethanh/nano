package retry_test

import (
	"context"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/interceptor/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// TestUnaryClientInterceptor_ReusedReplyDoesNotLeakStaleDataAcrossAttempts is
// a regression guard, not a bug fix: it was written to prove a suspected bug
// (the interceptor reuses the same reply pointer across attempts without
// resetting it, so a failed attempt's partial writes could survive into the
// final result). It turns out not to be a bug in practice: grpc-go's default
// codec (google.golang.org/grpc/encoding/proto) unmarshals responses via
// plain proto.Unmarshal, which resets the destination message before
// decoding unless Merge is explicitly set — so each attempt's decode already
// starts from a clean message regardless of what this interceptor does. This
// test pins that behavior down so a future codec change (or a custom codec
// that sets Merge: true) can't silently reintroduce the leak.
func TestUnaryClientInterceptor_ReusedReplyDoesNotLeakStaleDataAcrossAttempts(t *testing.T) {
	// What attempt 1 (which fails) writes into reply: a non-zero Status.
	staleBytes, err := proto.Marshal(&grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING})
	if err != nil {
		t.Fatalf("marshal stale: %v", err)
	}
	// What the real, successful attempt 2 response is: the zero value,
	// which proto3 does not encode on the wire.
	freshBytes, err := proto.Marshal(&grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_UNKNOWN})
	if err != nil {
		t.Fatalf("marshal fresh: %v", err)
	}

	attempt := 0
	invoker := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		attempt++
		msg := reply.(proto.Message)
		if attempt == 1 {
			if err := proto.Unmarshal(staleBytes, msg); err != nil {
				t.Fatalf("unmarshal stale: %v", err)
			}
			return status.Error(codes.Unavailable, "transient")
		}
		if err := proto.Unmarshal(freshBytes, msg); err != nil {
			t.Fatalf("unmarshal fresh: %v", err)
		}
		return nil
	}

	interceptor := retry.UnaryClientInterceptor(
		retry.WithRetryableCodes(codes.Unavailable),
		retry.WithBackoff(func(int) time.Duration { return 0 }),
	)

	reply := &grpc_health_v1.HealthCheckResponse{}
	if err := interceptor(context.Background(), "/svc/Method", nil, reply, nil, invoker); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reply.Status != grpc_health_v1.HealthCheckResponse_UNKNOWN {
		t.Errorf("reply.Status = %v after successful retry, want %v: stale data from the failed first attempt leaked through because reply was reused without being reset between attempts", reply.Status, grpc_health_v1.HealthCheckResponse_UNKNOWN)
	}
}
