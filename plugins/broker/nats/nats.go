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
	Nats struct {
		conn *nats.Conn
		opts []nats.Option
		log  logger

		addrs string
		codec codec
	}

	// Option is an optional configuration.
	Option func(*Nats)
)

var (
	_ broker.Broker = (*Nats)(nil)
)

// New return a new NATs message broker.
func New(opts ...Option) *Nats {
	n := &Nats{
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
func (n *Nats) Open(ctx context.Context) error {
	conn, err := nats.Connect(n.addrs, n.opts...)
	if err != nil {
		return err
	}
	n.conn = conn
	return nil
}

// Publish implements broker.Broker interface.
func (n *Nats) Publish(ctx context.Context, topic string, m *broker.Message, opts ...broker.PublishOption) error {
	b, err := n.codec.Marshal(m)
	if err != nil {
		return err
	}
	return n.conn.Publish(topic, b)
}

// Subscribe implements broker.Broker interface.
func (n *Nats) Subscribe(ctx context.Context, topic string, h broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	op := &broker.SubscribeOptions{
		AutoAck: true,
	}
	op.Apply(opts...)
	msgHandler := func(msg *nats.Msg) {
		m := broker.Message{}
		if err := n.codec.Unmarshal(msg.Data, &m); err != nil {
			h(&event{
				t:      topic,
				m:      &m,
				msg:    msg,
				err:    err,
				reason: broker.ReasonUnmarshalFailure,
			})
			return
		}
		h(&event{
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
func (n *Nats) CheckHealth(ctx context.Context) error {
	if !n.conn.IsConnected() {
		return fmt.Errorf("nats: server status=%d", n.conn.Status())
	}
	return nil
}

// Close flush in-flight messages and close the underlying connection.
func (n *Nats) Close(ctx context.Context) error {
	err := n.conn.FlushWithContext(ctx)
	if err != nil {
		return err
	}
	n.conn.Close()
	return err
}
