package auth_test

import (
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSentinelErrors_CarryUnauthenticatedStatusCode(t *testing.T) {
	sentinels := map[string]error{
		"ErrMetadataMissing":      auth.ErrMetadataMissing,
		"ErrAuthorizationMissing": auth.ErrAuthorizationMissing,
		"ErrInvalidToken":         auth.ErrInvalidToken,
		"ErrMultipleAuthFound":    auth.ErrMultipleAuthFound,
	}
	for name, err := range sentinels {
		t.Run(name, func(t *testing.T) {
			got := status.Code(err)
			if got != codes.Unauthenticated {
				t.Errorf("status.Code(%s) = %v, want %v: this package's own doc says implementations should return codes.Unauthenticated, but a plain errors.New sentinel surfaces as codes.Unknown", name, got, codes.Unauthenticated)
			}
		})
	}
}
