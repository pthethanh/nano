package health_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestRemove_StopsCheckerGoroutineWithoutClosingServer(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	before := goroutineDump(t)
	s.Add(health.Service{
		Name:     "svc",
		Delay:    health.NoDelay,
		Interval: 5 * time.Millisecond,
		Timeout:  time.Second,
		Checker:  health.CheckFunc(func(context.Context) error { return nil }),
	})
	waitForStatus(t, s, "svc", grpc_health_v1.HealthCheckResponse_SERVING)

	s.Remove("svc")

	// The checker goroutine must be gone...
	deadline := time.Now().Add(time.Second)
	for {
		after := goroutineDump(t)
		if strings.Count(after, "grpc/health.(*Server).Add.func1") <= strings.Count(before, "grpc/health.(*Server).Add.func1") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("checker goroutine for \"svc\" was still running 1s after Remove()")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// ...but the server itself must still be usable (Close() was not implied).
	if _, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: ""}); err != nil {
		t.Errorf("server unusable after Remove(): Check() error = %v", err)
	}
}

func TestRemove_UnknownServiceIsANoOp(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	s.Remove("never-added") // must not panic
}
