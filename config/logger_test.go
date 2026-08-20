package config_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/pthethanh/nano/config"
)

type capturingConfigLogger struct {
	calls int
}

func (l *capturingConfigLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.calls++
}

type testConfig struct {
	Name string `mapstructure:"name"`
}

func TestWithLogger_ReceivesReaderLogCalls(t *testing.T) {
	logger := &capturingConfigLogger{}

	_, err := config.NewReader[testConfig](
		config.WithLogger(logger),
		config.WithPaths("nonexistent", "yaml", "."),
	)
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}

	if logger.calls == 0 {
		t.Error("WithLogger()'s logger was never called: NewReader used the default slog logger instead of the one passed via WithLogger")
	}
}
