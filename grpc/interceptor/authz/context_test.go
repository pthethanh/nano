package authz_test

import (
	"context"
	"testing"

	"github.com/pthethanh/nano/grpc/interceptor/authz"
)

func TestMethodContext(t *testing.T) {
	ctx := context.Background()
	method := "/test.Service/Method"

	ctx = authz.NewMethodContext(ctx, method)
	got := authz.MethodFromContext(ctx)

	if got != method {
		t.Errorf("MethodFromContext() = %q, want %q", got, method)
	}
}

type testRequest struct{ Name string }

func TestRequestContext(t *testing.T) {
	ctx := context.Background()
	req := &testRequest{Name: "test"}

	ctx = authz.NewRequestContext(ctx, req)
	got, ok := authz.RequestFromContext[*testRequest](ctx)

	if !ok || got != req {
		t.Errorf("RequestFromContext() = (%v, %v), want (%v, true)", got, ok, req)
	}
}

func TestRequestContext_WrongType(t *testing.T) {
	ctx := context.Background()
	ctx = authz.NewRequestContext(ctx, &testRequest{Name: "test"})

	got, ok := authz.RequestFromContext[string](ctx)
	if ok || got != "" {
		t.Errorf("RequestFromContext[string]() = (%q, %v), want (\"\", false)", got, ok)
	}
}

func TestRequestContext_NotFound(t *testing.T) {
	ctx := context.Background()

	got, ok := authz.RequestFromContext[*testRequest](ctx)
	if ok || got != nil {
		t.Errorf("RequestFromContext() = (%v, %v), want (nil, false)", got, ok)
	}
}

func TestSubjectContext(t *testing.T) {
	ctx := context.Background()
	subject := "user123"

	ctx = authz.NewSubjectContext(ctx, subject)
	got := authz.SubjectFromContext(ctx)

	if got != subject {
		t.Errorf("SubjectFromContext() = %q, want %q", got, subject)
	}
}

func TestSubjectContext_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func(context.Context) context.Context
		want     string
	}{
		{
			name: "subject key",
			setupCtx: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, "subject", "user-from-subject")
			},
			want: "user-from-subject",
		},
		{
			name: "user key",
			setupCtx: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, "user", "user-from-user")
			},
			want: "user-from-user",
		},
		{
			name: "authz key takes precedence",
			setupCtx: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, "subject", "old-subject")
				ctx = context.WithValue(ctx, "user", "old-user")
				ctx = authz.NewSubjectContext(ctx, "new-subject")
				return ctx
			},
			want: "new-subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx(context.Background())
			got := authz.SubjectFromContext(ctx)
			if got != tt.want {
				t.Errorf("SubjectFromContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAnyContext(t *testing.T) {
	t.Run("string type", func(t *testing.T) {
		ctx := context.Background()
		want := "test-value"

		ctx = authz.NewAnyContext(ctx, want)
		got := authz.FromAnyContext[string](ctx)

		if got != want {
			t.Errorf("FromAnyContext[string]() = %q, want %q", got, want)
		}
	})

	t.Run("int type", func(t *testing.T) {
		ctx := context.Background()
		want := 42

		ctx = authz.NewAnyContext(ctx, want)
		got := authz.FromAnyContext[int](ctx)

		if got != want {
			t.Errorf("FromAnyContext[int]() = %d, want %d", got, want)
		}
	})

	t.Run("struct type", func(t *testing.T) {
		type User struct {
			ID   int
			Name string
		}

		ctx := context.Background()
		want := User{ID: 123, Name: "Alice"}

		ctx = authz.NewAnyContext(ctx, want)
		got := authz.FromAnyContext[User](ctx)

		if got != want {
			t.Errorf("FromAnyContext[User]() = %+v, want %+v", got, want)
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		type Config struct {
			Debug bool
		}

		ctx := context.Background()
		want := &Config{Debug: true}

		ctx = authz.NewAnyContext(ctx, want)
		got := authz.FromAnyContext[*Config](ctx)

		if got != want {
			t.Errorf("FromAnyContext[*Config]() = %p, want %p", got, want)
		}
	})

	t.Run("zero value when not found", func(t *testing.T) {
		ctx := context.Background()
		got := authz.FromAnyContext[string](ctx)

		if got != "" {
			t.Errorf("FromAnyContext[string]() = %q, want empty string", got)
		}

		gotInt := authz.FromAnyContext[int](ctx)
		if gotInt != 0 {
			t.Errorf("FromAnyContext[int]() = %d, want 0", gotInt)
		}
	})
}

// TestNoContextCollisions verifies that different context keys don't interfere with each other
func TestNoContextCollisions(t *testing.T) {
	ctx := context.Background()

	// Define custom types for testing
	type UserID int

	// Set up various context values
	ctx = authz.NewMethodContext(ctx, "/test.Service/Method")
	ctx = authz.NewRequestContext(ctx, "test-request")
	ctx = authz.NewSubjectContext(ctx, "user123")
	ctx = authz.NewAnyContext(ctx, UserID(456))

	// Verify all values are independently retrievable
	t.Run("method not affected", func(t *testing.T) {
		got := authz.MethodFromContext(ctx)
		if got != "/test.Service/Method" {
			t.Errorf("MethodFromContext() = %q, want %q", got, "/test.Service/Method")
		}
	})

	t.Run("request not affected", func(t *testing.T) {
		got, ok := authz.RequestFromContext[string](ctx)
		if !ok || got != "test-request" {
			t.Errorf("RequestFromContext() = (%q, %v), want (%q, true)", got, ok, "test-request")
		}
	})

	t.Run("subject not affected", func(t *testing.T) {
		got := authz.SubjectFromContext(ctx)
		if got != "user123" {
			t.Errorf("SubjectFromContext() = %q, want %q", got, "user123")
		}
	})

	t.Run("UserID not affected", func(t *testing.T) {
		got := authz.FromAnyContext[UserID](ctx)
		if got != UserID(456) {
			t.Errorf("FromAnyContext[UserID]() = %d, want 456", got)
		}
	})
}

// TestOverwritingSameType verifies that setting the same type twice overwrites the previous value
func TestOverwritingSameType(t *testing.T) {
	type UserID int

	ctx := context.Background()
	ctx = authz.NewAnyContext(ctx, UserID(100))
	ctx = authz.NewAnyContext(ctx, UserID(200))

	got := authz.FromAnyContext[UserID](ctx)
	if got != 200 {
		t.Errorf("FromAnyContext[UserID]() = %d, want 200 (should be overwritten)", got)
	}
}

// TestAnyContext_DifferentTypesDoNotCollide verifies that two different types stored via
// NewAnyContext on the same context chain each get their own isolated key, instead of the
// second value shadowing the first.
func TestAnyContext_DifferentTypesDoNotCollide(t *testing.T) {
	type UserID int
	type SessionID string

	ctx := context.Background()
	ctx = authz.NewAnyContext(ctx, UserID(1))
	ctx = authz.NewAnyContext(ctx, SessionID("abc"))

	if got := authz.FromAnyContext[UserID](ctx); got != UserID(1) {
		t.Errorf("FromAnyContext[UserID]() = %d, want 1", got)
	}
	if got := authz.FromAnyContext[SessionID](ctx); got != SessionID("abc") {
		t.Errorf("FromAnyContext[SessionID]() = %q, want %q", got, "abc")
	}
}
