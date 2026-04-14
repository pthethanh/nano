package retry

import (
	rand "math/rand/v2"
	"time"

	grpcbackoff "google.golang.org/grpc/backoff"
)

// Backoff returns a retry backoff function using the same delay progression and
// jitter formula as gRPC's exponential backoff strategy.
//
// Attempt numbering starts at 1 for the first retry after the initial call.
func Backoff(cfg grpcbackoff.Config) func(attempt int) time.Duration {
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = grpcbackoff.DefaultConfig.BaseDelay
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = grpcbackoff.DefaultConfig.Multiplier
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = grpcbackoff.DefaultConfig.MaxDelay
	}
	if cfg.Jitter < 0 {
		cfg.Jitter = 0
	}

	return func(attempt int) time.Duration {
		retries := attempt - 1
		if retries < 0 {
			retries = 0
		}
		delay, max := float64(cfg.BaseDelay), float64(cfg.MaxDelay)
		for delay < max && retries > 0 {
			delay *= cfg.Multiplier
			retries--
		}
		if delay > max {
			delay = max
		}
		return jitter(time.Duration(delay), cfg.Jitter)
	}
}

func jitter(delay time.Duration, factor float64) time.Duration {
	if factor == 0 {
		return delay
	}
	jittered := float64(delay) * (1 + factor*(rand.Float64()*2-1))
	if jittered < 0 {
		return 0
	}
	return time.Duration(jittered)
}
