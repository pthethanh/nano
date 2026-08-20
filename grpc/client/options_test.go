package client_test

import (
	"testing"
	"time"

	"github.com/pthethanh/nano/grpc/client"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/keepalive"
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

func TestWithKeepaliveReturnsDialOption(t *testing.T) {
	opt := client.WithKeepalive(keepalive.ClientParameters{Time: 30 * time.Second, Timeout: 10 * time.Second})
	if opt == nil {
		t.Fatal("expected dial option")
	}
}

func TestWithKeepaliveDefaultsReturnsDialOption(t *testing.T) {
	if opt := client.WithKeepaliveDefaults(); opt == nil {
		t.Fatal("expected dial option")
	}
}
