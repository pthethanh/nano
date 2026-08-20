package circuitbreaker_test

import (
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/circuitbreaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestErrOpen_CarriesResourceExhaustedStatusCode(t *testing.T) {
	got := status.Code(circuitbreaker.ErrOpen)
	if got != codes.ResourceExhausted {
		t.Errorf("status.Code(ErrOpen) = %v, want %v: a plain errors.New sentinel surfaces to clients as codes.Unknown, which breaks client-side logic that branches on status code (e.g. retry policies)", got, codes.ResourceExhausted)
	}
}
