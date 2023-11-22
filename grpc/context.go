package grpc

import (
	"context"
)

type ContextWrapFunc func(context.Context) (context.Context, error)
