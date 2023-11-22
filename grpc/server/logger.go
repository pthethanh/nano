package server

import (
	"context"
	"log/slog"
)

type (
	logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}
)
