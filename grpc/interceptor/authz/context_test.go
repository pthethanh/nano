package authz

import (
	"context"
	"testing"
)

func TestMethodContext(t *testing.T) {
	ctx := context.Background()
	method := "/test.Service/Method"

	ctx = NewMethodContext(ctx, method)
	got := MethodFromContext(ctx)

	if got != method {
		t.Errorf("MethodFromContext() = %q, want %q", got, method)
	}
}

func TestRequestContext(t *testing.T) {
	ctx := context.Background()
	req := &struct{ Name string }{Name: "test"}

	ctx = NewRequestContext(ctx, req)
	got := RequestFromContext(ctx)

	if got != req {
		t.Errorf("RequestFromContext() = %v, want %v", got, req)
	}
}

func TestSubjectContext(t *testing.T) {
	ctx := context.Background()
	subject := "user123"

	ctx = NewSubjectContext(ctx, subject)
	got := SubjectFromContext(ctx)

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
				ctx = NewSubjectContext(ctx, "new-subject")
				return ctx
			},
			want: "new-subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupCtx(context.Background())
			got := SubjectFromContext(ctx)
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

		ctx = NewAnyContext(ctx, want)
		got := FromAnyContext[string](ctx)

		if got != want {
			t.Errorf("FromAnyContext[string]() = %q, want %q", got, want)
		}
	})

	t.Run("int type", func(t *testing.T) {
		ctx := context.Background()
		want := 42

		ctx = NewAnyContext(ctx, want)
		got := FromAnyContext[int](ctx)

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

		ctx = NewAnyContext(ctx, want)
		got := FromAnyContext[User](ctx)

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

		ctx = NewAnyContext(ctx, want)
		got := FromAnyContext[*Config](ctx)

		if got != want {
			t.Errorf("FromAnyContext[*Config]() = %p, want %p", got, want)
		}
	})

	t.Run("zero value when not found", func(t *testing.T) {
		ctx := context.Background()
		got := FromAnyContext[string](ctx)

		if got != "" {
			t.Errorf("FromAnyContext[string]() = %q, want empty string", got)
		}

		gotInt := FromAnyContext[int](ctx)
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
	ctx = NewMethodContext(ctx, "/test.Service/Method")
	ctx = NewRequestContext(ctx, "test-request")
	ctx = NewSubjectContext(ctx, "user123")
	ctx = NewAnyContext(ctx, UserID(456))

	// Verify all values are independently retrievable
	t.Run("method not affected", func(t *testing.T) {
		got := MethodFromContext(ctx)
		if got != "/test.Service/Method" {
			t.Errorf("MethodFromContext() = %q, want %q", got, "/test.Service/Method")
		}
	})

	t.Run("request not affected", func(t *testing.T) {
		got := RequestFromContext(ctx)
		if got != "test-request" {
			t.Errorf("RequestFromContext() = %v, want %q", got, "test-request")
		}
	})

	t.Run("subject not affected", func(t *testing.T) {
		got := SubjectFromContext(ctx)
		if got != "user123" {
			t.Errorf("SubjectFromContext() = %q, want %q", got, "user123")
		}
	})

	t.Run("UserID not affected", func(t *testing.T) {
		got := FromAnyContext[UserID](ctx)
		if got != UserID(456) {
			t.Errorf("FromAnyContext[UserID]() = %d, want 456", got)
		}
	})
}

// TestOverwritingSameType verifies that setting the same type twice overwrites the previous value
func TestOverwritingSameType(t *testing.T) {
	type UserID int

	ctx := context.Background()
	ctx = NewAnyContext(ctx, UserID(100))
	ctx = NewAnyContext(ctx, UserID(200))

	got := FromAnyContext[UserID](ctx)
	if got != 200 {
		t.Errorf("FromAnyContext[UserID]() = %d, want 200 (should be overwritten)", got)
	}
}
