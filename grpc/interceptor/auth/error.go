package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// These sentinels carry codes.Unauthenticated, matching this package's own
// documented contract for Authenticator implementations, so clients see a
// meaningful status code instead of codes.Unknown.
var (
	// ErrMetadataMissing reports that metadata is missing in the incoming context.
	ErrMetadataMissing = status.Error(codes.Unauthenticated, "auth: could not locate request metadata")
	// ErrAuthorizationMissing reports that authorization metadata is missing in the incoming context.
	ErrAuthorizationMissing = status.Error(codes.Unauthenticated, "auth: could not locate authorization metadata")
	//ErrInvalidToken reports that the token is invalid.
	ErrInvalidToken = status.Error(codes.Unauthenticated, "auth: invalid token")
	// ErrMultipleAuthFound reports that too many authorization entries were found.
	ErrMultipleAuthFound = status.Error(codes.Unauthenticated, "auth: too many authorization entries")
)
