package status_test

import (
	"testing"
	"time"

	"github.com/pthethanh/nano/status"
	"google.golang.org/grpc/codes"
)

func TestErrorInfo_RoundTrip(t *testing.T) {
	s, err := status.WithErrorInfo(
		status.New(codes.PermissionDenied, "quota exceeded"),
		"QUOTA_EXCEEDED", "myservice.example.com",
		map[string]string{"limit": "100"},
	)
	if err != nil {
		t.Fatalf("WithErrorInfo() error = %v", err)
	}

	info, ok := status.ErrorInfo(s.Err())
	if !ok {
		t.Fatal("expected ErrorInfo to be found")
	}
	if info.Reason != "QUOTA_EXCEEDED" || info.Domain != "myservice.example.com" || info.Metadata["limit"] != "100" {
		t.Errorf("got ErrorInfo=%+v, want Reason=QUOTA_EXCEEDED Domain=myservice.example.com Metadata[limit]=100", info)
	}
}

func TestErrorInfo_NotPresent(t *testing.T) {
	if _, ok := status.ErrorInfo(status.Error(codes.Internal, "boom")); ok {
		t.Error("expected no ErrorInfo on a plain error")
	}
}

func TestBadRequestViolations_RoundTrip(t *testing.T) {
	s, err := status.WithBadRequest(
		status.New(codes.InvalidArgument, "invalid request"),
		status.FieldViolation{Field: "email", Description: "must be a valid email address"},
		status.FieldViolation{Field: "age", Description: "must be non-negative"},
	)
	if err != nil {
		t.Fatalf("WithBadRequest() error = %v", err)
	}

	violations, ok := status.BadRequestViolations(s.Err())
	if !ok {
		t.Fatal("expected violations to be found")
	}
	if len(violations) != 2 {
		t.Fatalf("got %d violations, want 2", len(violations))
	}
	if violations[0].Field != "email" || violations[1].Field != "age" {
		t.Errorf("got violations=%+v", violations)
	}
}

func TestRetryDelay_RoundTrip(t *testing.T) {
	s, err := status.WithRetryInfo(status.New(codes.Unavailable, "try again"), 5*time.Second)
	if err != nil {
		t.Fatalf("WithRetryInfo() error = %v", err)
	}

	delay, ok := status.RetryDelay(s.Err())
	if !ok {
		t.Fatal("expected a retry delay to be found")
	}
	if delay != 5*time.Second {
		t.Errorf("got delay=%v, want 5s", delay)
	}
}

func TestRetryDelay_NotPresent(t *testing.T) {
	if _, ok := status.RetryDelay(status.Error(codes.Internal, "boom")); ok {
		t.Error("expected no retry delay on a plain error")
	}
}

func TestErrorDetails_MultipleDetailsOnOneStatus(t *testing.T) {
	s, err := status.WithErrorInfo(status.New(codes.InvalidArgument, "bad input"), "BAD_INPUT", "myservice.example.com", nil)
	if err != nil {
		t.Fatalf("WithErrorInfo() error = %v", err)
	}
	s, err = status.WithBadRequest(s, status.FieldViolation{Field: "name", Description: "required"})
	if err != nil {
		t.Fatalf("WithBadRequest() error = %v", err)
	}

	if _, ok := status.ErrorInfo(s.Err()); !ok {
		t.Error("expected ErrorInfo to survive combining with BadRequest")
	}
	if violations, ok := status.BadRequestViolations(s.Err()); !ok || len(violations) != 1 {
		t.Errorf("expected BadRequest to survive combining with ErrorInfo, got violations=%v ok=%v", violations, ok)
	}
}
