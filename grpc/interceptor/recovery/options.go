package recovery

import (
	"context"
	"fmt"
	"runtime"
)

type (

	// Option customizes the behavior of the recovery interceptor.
	Option func(*options)

	// Handler is a function that recovers from the panic `p` by returning an `error`.
	// The context can be used to extract request scoped metadata and context values.
	Handler func(ctx context.Context, p any) (err error)

	options struct {
		handler Handler
	}

	// Error is the error type returned by the default recovery handler.
	Error struct {
		Err   any
		Stack []byte
	}
)

// WithHandler customizes the function for recovering from a panic.
func WithHandler(f Handler) Option {
	return func(o *options) {
		o.handler = f
	}
}

// StackHandler customizes the size of stack trace to be captured.
func StackHandler(stackSize int) Handler {
	return func(ctx context.Context, p any) error {
		stack := make([]byte, stackSize)
		stack = stack[:runtime.Stack(stack, false)]
		return &Error{Err: p, Stack: stack}
	}
}

func newOpts(opts ...Option) *options {
	opt := &options{
		handler: nil,
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func (e *Error) Error() string {
	return fmt.Sprintf("panic recovered: %v\n\n%s", e.Err, e.Stack)
}
