package status

import (
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/durationpb"
)

// FieldViolation describes a single invalid field in a request, used with
// WithBadRequest and returned by BadRequestViolations.
type FieldViolation struct {
	// Field is a (dot-separated, for nested fields) path identifying the
	// invalid field, e.g. "email" or "address.zip_code".
	Field string
	// Description explains why the field is invalid, in a form safe to
	// return to the caller.
	Description string
}

// WithErrorInfo attaches a google.rpc.ErrorInfo detail to s, giving clients
// a stable, machine-readable reason/domain/metadata for the failure instead
// of only a human-readable message they'd have to string-match.
//
//	s, err := status.WithErrorInfo(
//		status.New(codes.PermissionDenied, "quota exceeded"),
//		"QUOTA_EXCEEDED", "myservice.example.com",
//		map[string]string{"limit": "100"},
//	)
func WithErrorInfo(s *Status, reason, domain string, metadata map[string]string) (*Status, error) {
	return s.WithDetails(&errdetails.ErrorInfo{
		Reason:   reason,
		Domain:   domain,
		Metadata: metadata,
	})
}

// WithBadRequest attaches a google.rpc.BadRequest detail listing which
// request fields were invalid and why, so clients can highlight the
// specific fields instead of parsing a single error string.
func WithBadRequest(s *Status, violations ...FieldViolation) (*Status, error) {
	fv := make([]*errdetails.BadRequest_FieldViolation, 0, len(violations))
	for _, v := range violations {
		fv = append(fv, &errdetails.BadRequest_FieldViolation{
			Field:       v.Field,
			Description: v.Description,
		})
	}
	return s.WithDetails(&errdetails.BadRequest{FieldViolations: fv})
}

// WithRetryInfo attaches a google.rpc.RetryInfo detail telling the client
// how long to wait before retrying, instead of leaving retry timing to
// guesswork.
func WithRetryInfo(s *Status, retryDelay time.Duration) (*Status, error) {
	return s.WithDetails(&errdetails.RetryInfo{RetryDelay: durationpb.New(retryDelay)})
}

// ErrorInfo extracts the first google.rpc.ErrorInfo detail from err, if any
// (see WithErrorInfo).
func ErrorInfo(err error) (*errdetails.ErrorInfo, bool) {
	s, ok := FromError(err)
	if !ok {
		return nil, false
	}
	for _, d := range s.Details() {
		if info, ok := d.(*errdetails.ErrorInfo); ok {
			return info, true
		}
	}
	return nil, false
}

// BadRequestViolations extracts the field violations from the first
// google.rpc.BadRequest detail on err, if any (see WithBadRequest).
func BadRequestViolations(err error) ([]FieldViolation, bool) {
	s, ok := FromError(err)
	if !ok {
		return nil, false
	}
	for _, d := range s.Details() {
		br, ok := d.(*errdetails.BadRequest)
		if !ok {
			continue
		}
		out := make([]FieldViolation, 0, len(br.GetFieldViolations()))
		for _, fv := range br.GetFieldViolations() {
			out = append(out, FieldViolation{Field: fv.GetField(), Description: fv.GetDescription()})
		}
		return out, true
	}
	return nil, false
}

// RetryDelay extracts the retry delay from the first google.rpc.RetryInfo
// detail on err, if any (see WithRetryInfo).
func RetryDelay(err error) (time.Duration, bool) {
	s, ok := FromError(err)
	if !ok {
		return 0, false
	}
	for _, d := range s.Details() {
		if ri, ok := d.(*errdetails.RetryInfo); ok {
			return ri.GetRetryDelay().AsDuration(), true
		}
	}
	return 0, false
}
