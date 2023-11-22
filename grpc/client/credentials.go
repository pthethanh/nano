package client

import (
	"context"

	"google.golang.org/grpc/credentials"
)

type tokenCredentials string

// NewTokenCredentials returns a PerRPCCredentials using the provided token.
func NewTokenCredentials(token string) credentials.PerRPCCredentials {
	return tokenCredentials(token)
}

func (tok tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": string(tok),
	}, nil
}

func (tok tokenCredentials) RequireTransportSecurity() bool {
	return false
}
