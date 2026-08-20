package watermill_test

import (
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/plugins/broker/watermill"
)

// fakeSubscriber hands back a channel that never produces a message and
// never closes on its own, so the only way the consume loop advances is via
// context cancellation (i.e. Unsubscribe()).
type fakeSubscriber struct {
	ch chan *message.Message
}

func newFakeSubscriber() *fakeSubscriber {
	return &fakeSubscriber{ch: make(chan *message.Message)}
}

func (f *fakeSubscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return f.ch, nil
}

func (f *fakeSubscriber) Close() error { return nil }

type fakePublisher struct{}

func (fakePublisher) Publish(topic string, messages ...*message.Message) error { return nil }
func (fakePublisher) Close() error                                             { return nil }

type countingLogger struct {
	contextDoneCount atomic.Int64
}

func (l *countingLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if strings.Contains(msg, "context done") {
		l.contextDoneCount.Add(1)
	}
}

func TestSubscribe_UnsubscribeStopsConsumeLoop(t *testing.T) {
	sub := newFakeSubscriber()
	logger := &countingLogger{}
	b := watermill.New[testMsg](fakePublisher{}, sub, watermill.Logger[testMsg](logger))

	subscription, err := b.Subscribe(context.Background(), "topic", func(ev broker.Event[testMsg]) error { return nil })
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	if err := subscription.Unsubscribe(); err != nil {
		t.Fatalf("Unsubscribe() error = %v", err)
	}

	// The consume goroutine must notice the cancellation that Unsubscribe()
	// triggers. If it's watching the wrong context, it never will, and this
	// deadline trips.
	deadline := time.Now().Add(500 * time.Millisecond)
	for logger.contextDoneCount.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("consume goroutine never noticed Unsubscribe(): the cancellation Unsubscribe() triggers must stop it, but no ctx.Done() branch fired within 500ms")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Once it does notice, it must return rather than loop back into
	// select forever (ctx.Done() stays permanently ready once cancelled).
	count1 := logger.contextDoneCount.Load()
	time.Sleep(50 * time.Millisecond)
	count2 := logger.contextDoneCount.Load()
	if count2 != count1 {
		t.Fatalf("consume goroutine kept looping after noticing cancellation instead of returning: log count grew from %d to %d in 50ms", count1, count2)
	}
}
