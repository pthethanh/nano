package client

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
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
