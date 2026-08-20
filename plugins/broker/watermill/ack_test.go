package watermill_test

import (
	"context"
	"testing"
	"time"

	wm "github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/plugins/broker/watermill"
)

func TestSubscribe_FailedUnmarshalEventCanStillBeAcked(t *testing.T) {
	sub := newFakeSubscriber()
	b := watermill.New[testMsg](fakePublisher{}, sub)

	received := make(chan broker.Event[testMsg], 1)
	_, err := b.Subscribe(context.Background(), "topic", func(ev broker.Event[testMsg]) error {
		received <- ev
		return nil
	}, broker.DisableAutoAck())
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	badMsg := message.NewMessage(wm.NewUUID(), []byte("not-valid-json"))
	sub.ch <- badMsg

	var ev broker.Event[testMsg]
	select {
	case ev = <-received:
	case <-time.After(time.Second):
		t.Fatal("handler was not invoked for the unmarshal-failure message")
	}

	if ev.Error() == nil {
		t.Fatal("expected the event to carry the unmarshal error")
	}
	if err := ev.Ack(); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	select {
	case <-badMsg.Acked():
		// good: Ack() on the failure event reached the underlying message.
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Ack() on an unmarshal-failure event did not ack the underlying watermill message (event.raw was never set on the failure path, so Ack() silently no-ops)")
	}
}
