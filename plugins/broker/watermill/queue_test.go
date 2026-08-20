package watermill_test

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/plugins/broker/watermill"
)

type warnCountingLogger struct {
	warnCount atomic.Int64
}

func (l *warnCountingLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if level >= slog.LevelWarn && strings.Contains(strings.ToLower(msg), "queue") {
		l.warnCount.Add(1)
	}
}

func TestSubscribe_QueueOptionWarnsSinceItIsUnsupported(t *testing.T) {
	sub := newFakeSubscriber()
	logger := &warnCountingLogger{}
	b := watermill.New[testMsg](fakePublisher{}, sub, watermill.Logger[testMsg](logger))

	_, err := b.Subscribe(context.Background(), "topic", func(ev broker.Event[testMsg]) error { return nil }, broker.Queue("q1"))
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if logger.warnCount.Load() == 0 {
		t.Error("Subscribe() with broker.Queue(...) did not warn that watermill does not support queue groups; the option is silently ignored")
	}
}
