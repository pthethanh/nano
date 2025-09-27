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
	JSONCodec struct{}
)

func (m JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (m JSONCodec) Unmarshal(data []byte, v any) error {
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
