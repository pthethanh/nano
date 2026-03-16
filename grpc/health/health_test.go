package health

import (
	"context"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckAndUpdateNilCheckerPanics(t *testing.T) {
	s := NewServer()
	defer s.Close()

	s.checkAndUpdate("svc", time.Millisecond, nil)

	rs, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: "svc"})
	if err != nil {
		t.Fatal(err)
	}
	if rs.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("got service status=%v, want %v", rs.Status, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	rs, err = s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if rs.Status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("got overall status=%v, want %v", rs.Status, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
}

func TestCheckAndUpdateRecomputesOverallStatusWithoutSelfEntry(t *testing.T) {
	s := NewServer()
	defer s.Close()

	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)

	s.checkAndUpdate("svc", time.Second, CheckFunc(func(context.Context) error { return nil }))

	rs, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if rs.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("got overall status=%v, want %v", rs.Status, grpc_health_v1.HealthCheckResponse_SERVING)
	}
}

func TestCheckAndUpdateTimeoutDoesNotLeakWorkerGoroutine(t *testing.T) {
	s := NewServer()
	defer s.Close()

	before := goroutineDump(t)
	s.checkAndUpdate("svc", 10*time.Millisecond, CheckFunc(func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}))
	time.Sleep(20 * time.Millisecond)
	after := goroutineDump(t)

	if strings.Count(after, "grpc/health.(*Server).checkAndUpdate.func1") > strings.Count(before, "grpc/health.(*Server).checkAndUpdate.func1") {
		t.Fatal("worker goroutine leaked after timeout")
	}
}

func goroutineDump(t *testing.T) string {
	t.Helper()

	var b strings.Builder
	if err := pprof.Lookup("goroutine").WriteTo(&b, 2); err != nil {
		t.Fatal(err)
	}
	return b.String()
}
