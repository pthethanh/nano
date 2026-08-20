package nats

import (
	"testing"

	natsgo "github.com/nats-io/nats.go"
)

// White-box test (same package): plain core-NATS messages (as delivered by
// Subscribe/QueueSubscribe, which is all this broker uses) have no
// JetStream ack/redelivery concept. e.msg.Ack() forwards to nats.go's
// JetStream-only Ack, which errors on such a message.
func TestEvent_AckIsANoOpOnPlainCoreNATSMessages(t *testing.T) {
	msg := &natsgo.Msg{Subject: "topic", Data: []byte("hi")} // no Sub/Reply: what Subscribe() actually delivers
	e := &event[string]{t: "topic", msg: msg}

	if err := e.Ack(); err != nil {
		t.Fatalf("Ack() error = %v, want nil: plain core NATS has no ack/redelivery concept, so Ack() must be a safe no-op instead of surfacing a confusing JetStream-only error", err)
	}
}
