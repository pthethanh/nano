package ratelimit_test

import (
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrLimited_CarriesResourceExhaustedStatusCode(t *testing.T) {
	got := status.Code(ratelimit.ErrLimited)
	if got != codes.ResourceExhausted {
		t.Errorf("status.Code(ErrLimited) = %v, want %v: a plain errors.New sentinel surfaces to clients as codes.Unknown", got, codes.ResourceExhausted)
	}
}
