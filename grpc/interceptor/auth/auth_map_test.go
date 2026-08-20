package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/auth"
	"google.golang.org/grpc/metadata"
)

func TestMapAuthenticator_MalformedHeaderIsRejected(t *testing.T) {
	// A permissive default/catch-all authenticator registered under "" is a
	// realistic setup (e.g. accepting a bare token with no scheme prefix).
	m := auth.MapAuthenticator{
		"": auth.AuthenticatorFunc(func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		}),
	}

	tests := []struct {
		name   string
		header string
	}{
		{name: "all whitespace", header: "   "},
		{name: "three words", header: "Bearer abc extra"},
		{name: "four words", header: "a b c d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := metadata.Pairs(auth.AuthorizationMD, tt.header)
			ctx := metadata.NewIncomingContext(context.Background(), md)

			_, err := m.Authenticate(ctx)
			if !errors.Is(err, auth.ErrInvalidToken) {
				t.Errorf("Authenticate() error = %v, want %v: a malformed Authorization header must not silently authenticate via a catch-all authenticator", err, auth.ErrInvalidToken)
			}
		})
	}
}

func TestMapAuthenticator_WellFormedHeadersStillWork(t *testing.T) {
	called := false
	m := auth.MapAuthenticator{
		"Bearer": auth.AuthenticatorFunc(func(ctx context.Context) (context.Context, error) {
			called = true
			return ctx, nil
		}),
	}
	md := metadata.Pairs(auth.AuthorizationMD, "Bearer sometoken")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if _, err := m.Authenticate(ctx); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !called {
		t.Error("expected the \"Bearer\" authenticator to be invoked for a well-formed two-word header")
	}
}

func TestMapAuthenticator_BareTokenStillWorks(t *testing.T) {
	called := false
	m := auth.MapAuthenticator{
		"": auth.AuthenticatorFunc(func(ctx context.Context) (context.Context, error) {
			called = true
			return ctx, nil
		}),
	}
	md := metadata.Pairs(auth.AuthorizationMD, "sometoken")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	if _, err := m.Authenticate(ctx); err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if !called {
		t.Error("expected the default (\"\") authenticator to be invoked for a well-formed single-word header")
	}
}
