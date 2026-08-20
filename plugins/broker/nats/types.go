package nats

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/nano/broker"
)

type (
	event[T any] struct {
		t      string
		m      *T
		msg    *nats.Msg
		err    error
		reason broker.Reason
	}
	subscriber struct {
		t string
		s *nats.Subscription
	}
	logger interface {
		Log(ctx context.Context, level slog.Level, msg string, args ...any)
	}
	JSONCodec[T any] struct{}
)

func (m JSONCodec[T]) Marshal(v *T) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONCodec[T]) Unmarshal(data []byte, v *T) error {
	return json.Unmarshal(data, v)
}

// natsHeaderFrom converts broker.PublishOptions.Headers into NATS message headers.
func natsHeaderFrom(headers map[string]string) nats.Header {
	if len(headers) == 0 {
		return nil
	}
	h := make(nats.Header, len(headers))
	for k, v := range headers {
		h.Set(k, v)
	}
	return h
}

func (e *event[T]) Topic() string {
	return e.t
}

func (e *event[T]) Message() *T {
	return e.m
}

// Ack is a no-op: this broker subscribes via plain core NATS
// (Subscribe/QueueSubscribe), which has no application-level ack or
// redelivery concept. The underlying nats.Msg.Ack is a JetStream-only
// operation that would return an error (or, if the message happens to have
// a Reply subject set for an unrelated request-reply exchange, send a
// misleading response there) for messages delivered this way.
func (e *event[T]) Ack() error {
	return nil
}

func (e *event[T]) Error() error {
	return e.err
}

func (e *event[T]) Reason() broker.Reason {
	return e.reason
}

func (s *subscriber) Topic() string {
	return s.t
}

func (s *subscriber) Unsubscribe() error {
	return s.s.Unsubscribe()
}
