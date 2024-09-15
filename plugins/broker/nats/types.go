package nats

import (
	"context"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/nano/broker"

	json "github.com/bytedance/sonic"
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

	// Codec defines the interface gRPC uses to encode and decode messages.  Note
	// that implementations of this interface must be thread safe; a Codec's
	// methods can be called from concurrent goroutines.
	codec interface {
		// Marshal returns the wire format of v.
		Marshal(v any) ([]byte, error)
		// Unmarshal parses the wire format into v.
		Unmarshal(data []byte, v any) error
	}
	JSONCodec struct{}
)

func (m JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (e *event[T]) Topic() string {
	return e.t
}

func (e *event[T]) Message() *T {
	return e.m
}

func (e *event[T]) Ack() error {
	return e.msg.Ack()
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
