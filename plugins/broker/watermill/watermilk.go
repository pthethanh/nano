package watermill

import (
	"context"
	"errors"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pthethanh/nano/broker"
)

// Broker implements nano's broker plugin using Watermill.
type Broker[T any] struct {
	pub    message.Publisher
	sub    message.Subscriber
	codec  broker.Codec
	logger logger
}

// Ensure Broker implements broker.Broker interface
var (
	_ broker.Broker[any] = (*Broker[any])(nil)
)

type Option[T any] func(*Broker[T])

// Option to set a custom codec.
func Codec[T any](c broker.Codec) Option[T] {
	return func(b *Broker[T]) {
		b.codec = c
	}
}

// Option to set a custom logger.
func Logger[T any](l logger) Option[T] {
	return func(b *Broker[T]) {
		b.logger = l
	}
}

func New[T any](pub message.Publisher, sub message.Subscriber, opts ...Option[T]) *Broker[T] {
	b := &Broker[T]{
		pub:    pub,
		sub:    sub,
		codec:  jsonCodec{}, // Use JSONCodec as default
		logger: defaultLogger{},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Broker[T]) Open(ctx context.Context) error {
	b.logger.Log(ctx, slog.LevelDebug, "broker opened")
	return nil
}

func (b *Broker[T]) Close(ctx context.Context) error {
	b.logger.Log(ctx, slog.LevelDebug, "broker closing")
	var errPub, errSub error
	if b.pub != nil {
		errPub = b.pub.Close()
	}
	if b.sub != nil {
		errSub = b.sub.Close()
	}
	err := errors.Join(errPub, errSub)
	if err != nil {
		b.logger.Log(ctx, slog.LevelError, "broker close error", "error", err)
		return err
	}
	b.logger.Log(ctx, slog.LevelDebug, "broker closed successfully")
	return nil
}

func (b *Broker[T]) Publish(ctx context.Context, topic string, m *T, opts ...broker.PublishOption) error {
	if b.pub == nil {
		b.logger.Log(ctx, slog.LevelError, "publish failed: publisher not initialized")
		return errors.New("publisher not initialized")
	}
	data, err := b.codec.Marshal(m)
	if err != nil {
		b.logger.Log(ctx, slog.LevelError, "publish failed: marshal error", "error", err)
		return err
	}
	msg := message.NewMessage(watermill.NewUUID(), data)
	b.logger.Log(ctx, slog.LevelDebug, "publishing message", "topic", topic, "msg_id", msg.UUID)
	err = b.pub.Publish(topic, msg)
	if err != nil {
		b.logger.Log(ctx, slog.LevelError, "publish failed", "topic", topic, "error", err)
	} else {
		b.logger.Log(ctx, slog.LevelDebug, "publish succeeded", "topic", topic, "msg_id", msg.UUID)
	}
	return err
}

func (b *Broker[T]) Subscribe(ctx context.Context, topic string, handler func(broker.Event[T]) error, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	if b.sub == nil {
		b.logger.Log(ctx, slog.LevelError, "subscribe failed: subscriber not initialized")
		return nil, errors.New("subscriber not initialized")
	}
	opt := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   "", // queue is not supported in watermill, so we ignore it for now.
	}
	opt.Apply(opts...)
	newCtx, cancel := context.WithCancel(ctx)
	messages, err := b.sub.Subscribe(newCtx, topic)
	if err != nil {
		b.logger.Log(ctx, slog.LevelError, "subscribe failed", "topic", topic, "error", err)
		cancel()
		return nil, err
	}
	sub := &subscriber[T]{topic: topic, cancel: cancel}
	b.logger.Log(ctx, slog.LevelDebug, "subscribed to topic", "topic", topic)
	go func() {
		for {
			select {
			case msg := <-messages:
				if msg == nil {
					b.logger.Log(ctx, slog.LevelDebug, "message channel closed", "topic", topic)
					return
				}
				var v T
				if err := b.codec.Unmarshal(msg.Payload, &v); err != nil {
					if opt.AutoAck {
						msg.Nack()
					}
					b.logger.Log(ctx, slog.LevelError, "failed to unmarshal message", "topic", topic, "error", err)
					_ = handler(&event[T]{
						topic:  topic,
						err:    err,
						reason: broker.ReasonUnmarshalFailure,
					})
					continue
				}
				b.logger.Log(ctx, slog.LevelDebug, "received message", "topic", topic, "msg_id", msg.UUID)
				_ = handler(&event[T]{
					topic:   topic,
					payload: &v,
					raw:     msg,
				})
				if opt.AutoAck {
					msg.Ack()
					b.logger.Log(ctx, slog.LevelDebug, "message acknowledged", "topic", topic, "msg_id", msg.UUID)
				}
			case <-ctx.Done():
				b.logger.Log(ctx, slog.LevelDebug, "context done, stopping subscription", "topic", topic)
			}
		}
	}()
	return sub, nil
}

// defaultLogger is a simple logger that uses slog.Default().
type defaultLogger struct{}

func (defaultLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	slog.Log(ctx, level, msg, args...)
}
