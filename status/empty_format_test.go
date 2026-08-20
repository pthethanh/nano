package status_test

import (
	"strings"
	"testing"

	"github.com/pthethanh/nano/status"
	"google.golang.org/grpc/codes"
)

func TestNew_EmptyFormatWithArgsFallsBackToCodeString(t *testing.T) {
	emptyFormat := "" // non-literal so `go vet`'s printf check doesn't flag this deliberate edge case
	s := status.New(codes.NotFound, emptyFormat, 42)
	got := s.Message()
	if strings.Contains(got, "EXTRA") || strings.Contains(got, "%!") {
		t.Fatalf("New(code, \"\", 42).Message() = %q, want a clean fallback to code.String() (got a garbled fmt.Sprintf artifact instead)", got)
	}
	if got != codes.NotFound.String() {
		t.Errorf("New(code, \"\", 42).Message() = %q, want %q", got, codes.NotFound.String())
	}
}

func TestError_EmptyFormatWithArgsFallsBackToCodeString(t *testing.T) {
	emptyFormat := "" // non-literal so `go vet`'s printf check doesn't flag this deliberate edge case
	err := status.Error(codes.Internal, emptyFormat, "unused-arg")
	if strings.Contains(err.Error(), "EXTRA") || strings.Contains(err.Error(), "%!") {
		t.Fatalf("Error(code, \"\", ...).Error() = %q, want a clean fallback to code.String()", err.Error())
	}
}
