package health_test

import (
	"context"
	"runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestAddNilCheckerMarksServiceNotServing(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	s.Add(health.Service{
		Name:     "svc",
		Delay:    health.NoDelay,
		Interval: time.Hour,
		Timeout:  time.Millisecond,
		Checker:  nil,
	})

	waitForStatus(t, s, "svc", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	waitForStatus(t, s, "", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
}

func TestAddRecomputesOverallStatusWithoutSelfEntry(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	s.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	s.Add(health.Service{
		Name:     "svc",
		Delay:    health.NoDelay,
		Interval: time.Hour,
		Timeout:  time.Second,
		Checker:  health.CheckFunc(func(context.Context) error { return nil }),
	})

	waitForStatus(t, s, "", grpc_health_v1.HealthCheckResponse_SERVING)
}

func TestAddTimeoutDoesNotLeakWorkerGoroutine(t *testing.T) {
	s := health.NewServer()
	defer s.Close()

	before := goroutineDump(t)
	s.Add(health.Service{
		Name:     "svc",
		Delay:    health.NoDelay,
		Interval: time.Hour,
		Timeout:  10 * time.Millisecond,
		Checker: health.CheckFunc(func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}),
	})
	waitForStatus(t, s, "svc", grpc_health_v1.HealthCheckResponse_UNKNOWN)
	time.Sleep(20 * time.Millisecond)
	after := goroutineDump(t)

	if strings.Count(after, "grpc/health.(*Server).checkAndUpdate.func1") > strings.Count(before, "grpc/health.(*Server).checkAndUpdate.func1") {
		t.Fatal("worker goroutine leaked after timeout")
	}
}

func waitForStatus(t *testing.T, s *health.Server, service string, want grpc_health_v1.HealthCheckResponse_ServingStatus) {
	t.Helper()

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		rs, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: service})
		if err == nil && rs.Status == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	rs, err := s.Check(context.Background(), &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		t.Fatalf("final check failed for service %q: %v", service, err)
	}
	t.Fatalf("got service %q status=%v, want %v", service, rs.Status, want)
}

func goroutineDump(t *testing.T) string {
	t.Helper()

	var b strings.Builder
	if err := pprof.Lookup("goroutine").WriteTo(&b, 2); err != nil {
		t.Fatal(err)
	}
	return b.String()
}
