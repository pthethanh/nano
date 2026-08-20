package recovery

import (
	"context"
	"log/slog"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	// Error carries the recovered panic value and stack trace for a custom
	// Handler's own use (e.g. logging). Its Error() message deliberately
	// omits both, since a Handler's returned error is what the RPC caller
	// sees on the wire.
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

// StackHandler builds a Handler that captures a stack trace of up to stackSize
// bytes, logs the recovered panic and stack server-side via slog, and returns
// a generic codes.Internal error to the RPC caller. The panic value and stack
// trace are never included in the error returned to the caller.
func StackHandler(stackSize int) Handler {
	return func(ctx context.Context, p any) error {
		stack := make([]byte, stackSize)
		stack = stack[:runtime.Stack(stack, false)]
		slog.ErrorContext(ctx, "panic recovered", "panic", p, "stack", string(stack))
		return status.Error(codes.Internal, "internal error")
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

// Error implements the error interface without including the panic value or
// stack trace, since this message can end up on the wire as an RPC status
// message. Use the Err and Stack fields directly for server-side logging.
func (e *Error) Error() string {
	return "panic recovered"
}
