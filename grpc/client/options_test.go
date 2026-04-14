package client_test

import (
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/client"
	"google.golang.org/grpc/backoff"
)

func TestWithRoundRobinReturnsDialOption(t *testing.T) {
	if opt := client.WithRoundRobin(); opt == nil {
		t.Fatal("expected dial option")
	}
}

func TestWithConnectParamsReturnsDialOption(t *testing.T) {
	cfg := backoff.Config{
		BaseDelay:  10 * time.Millisecond,
		Multiplier: 1.6,
		Jitter:     0.2,
		MaxDelay:   time.Second,
	}
	if opt := client.WithConnectParams(cfg, 3*time.Second); opt == nil {
		t.Fatal("expected dial option")
	}
}
