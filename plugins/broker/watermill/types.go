package watermill

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pthethanh/nano/broker"
)

// logger interface for pluggable logging, similar to plugins/kafka.
type logger interface {
	Log(ctx context.Context, level slog.Level, msg string, args ...any)
}

// jsonCodec implements broker.Codec using encoding/json.
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

type event[T any] struct {
	topic   string
	payload *T
	raw     *message.Message
	err     error
	reason  broker.Reason
}

func (e *event[T]) Topic() string {
	return e.topic
}

func (e *event[T]) Message() *T {
	return e.payload
}

func (e *event[T]) RawMessage() *message.Message {
	return e.raw
}

func (e *event[T]) Ack() error {
	if e.raw != nil {
		e.raw.Ack()
	}
	return nil
}

func (e *event[T]) Error() error {
	return e.err
}

func (e *event[T]) Reason() broker.Reason {
	return e.reason
}

type subscriber[T any] struct {
	topic  string
	cancel func()
}

func (s *subscriber[T]) Unsubscribe() error {
	s.cancel()
	return nil
}

func (s *subscriber[T]) Topic() string {
	return s.topic
}
