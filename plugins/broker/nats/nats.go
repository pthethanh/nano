// Package nats provide a message broker using NATS.
package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/pthethanh/nano/broker"
)

type (
	// Nats is an implementation of broker.Broker using NATS.
	Nats[T any] struct {
		conn *nats.Conn
		opts []nats.Option
		log  logger

		addrs string
		codec codec
	}

	// Option is an optional configuration.
	Option[T any] func(*Nats[T])
)

var (
	_ broker.Broker[any] = (*Nats[any])(nil)
)

// New return a new NATs message broker.
func New[T any](opts ...Option[T]) *Nats[T] {
	n := &Nats[T]{
		log:   slog.Default(),
		codec: JSONCodec{},
		addrs: "127.0.0.1:4222",
	}
	// apply the options.
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// Open connect to target server.
func (n *Nats[T]) Open(ctx context.Context) error {
	conn, err := nats.Connect(n.addrs, n.opts...)
	if err != nil {
		return err
	}
	n.conn = conn
	return nil
}

// Publish implements broker.Broker interface.
func (n *Nats[T]) Publish(ctx context.Context, topic string, m *T, opts ...broker.PublishOption[T]) error {
	b, err := n.codec.Marshal(m)
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, b)
}

// Subscribe implements broker.Broker interface.
func (n *Nats[T]) Subscribe(ctx context.Context, topic string, h func(broker.Event[T]) error, opts ...broker.SubscribeOption[T]) (broker.Subscriber[T], error) {
	op := &broker.SubscribeOptions[T]{
		AutoAck: true,
	}
	op.Apply(opts...)
	msgHandler := func(msg *nats.Msg) {
		var m T
		if err := n.codec.Unmarshal(msg.Data, &m); err != nil {
			h(&event[T]{
				t:      topic,
				m:      &m,
				msg:    msg,
				err:    err,
				reason: broker.ReasonUnmarshalFailure,
			})
			return
		}
		h(&event[T]{
			t:   topic,
			m:   &m,
			msg: msg,
		})
	}
	if op.Queue != "" {
		sub, err := n.conn.QueueSubscribe(topic, op.Queue, msgHandler)
		if err != nil {
			return nil, err
		}
		return &subscriber{
			t: topic,
			s: sub,
		}, nil
	}
	sub, err := n.conn.Subscribe(topic, msgHandler)
	if err != nil {
		return nil, err
	}
	return &subscriber{
		t: topic,
		s: sub,
	}, nil
}

// CheckHealth implements health.Checker.
func (n *Nats[T]) CheckHealth(ctx context.Context) error {
	if !n.conn.IsConnected() {
		return fmt.Errorf("nats: server status=%d", n.conn.Status())
	}
	return nil
}

// Close flush in-flight messages and close the underlying connection.
func (n *Nats[T]) Close(ctx context.Context) error {
	err := n.conn.FlushWithContext(ctx)
	if err != nil {
		return err
	}
	n.conn.Close()
	return err
}
