package client

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/keepalive"
)

// WithRoundRobin enables the standard gRPC round_robin balancing policy.
func WithRoundRobin() grpc.DialOption {
	return grpc.WithDefaultServiceConfig(`{"loadBalancingConfig":[{"round_robin":{}}]}`)
}

// WithConnectParams configures client connection backoff and minimum connect timeout.
func WithConnectParams(cfg backoff.Config, minConnectTimeout time.Duration) grpc.DialOption {
	return grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           cfg,
		MinConnectTimeout: minConnectTimeout,
	})
}

// WithKeepalive configures client-side keepalive ping parameters, so idle
// connections behind L4 load balancers and NAT gateways aren't silently
// dropped for inactivity.
func WithKeepalive(params keepalive.ClientParameters) grpc.DialOption {
	return grpc.WithKeepaliveParams(params)
}

// WithKeepaliveDefaults applies production-sensible keepalive settings:
// ping every 30s (10s timeout to react), sent even without active RPCs so a
// fully-idle connection is still probed rather than silently dropped.
func WithKeepaliveDefaults() grpc.DialOption {
	return WithKeepalive(keepalive.ClientParameters{
		Time:                30 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	})
}
