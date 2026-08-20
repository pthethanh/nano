package watermill_test

import (
	"context"
	"testing"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/plugins/broker/watermill"
)

type capturingPublisher struct {
	published []*message.Message
}

func (p *capturingPublisher) Publish(topic string, messages ...*message.Message) error {
	p.published = append(p.published, messages...)
	return nil
}
func (p *capturingPublisher) Close() error { return nil }

func TestPublish_HeadersBecomeMessageMetadata(t *testing.T) {
	pub := &capturingPublisher{}
	b := watermill.New[testMsg](pub, newFakeSubscriber())

	err := b.Publish(context.Background(), "topic", &testMsg{ID: "1"}, broker.Header("trace-id", "abc"))
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if len(pub.published) != 1 {
		t.Fatalf("got %d published messages, want 1", len(pub.published))
	}
	if got := pub.published[0].Metadata.Get("trace-id"); got != "abc" {
		t.Errorf("Metadata[trace-id] = %q, want %q", got, "abc")
	}
}
