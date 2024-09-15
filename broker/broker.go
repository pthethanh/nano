// Package broker defines standard interface for a message broker.
package broker

import (
	"context"
)

type (
	// Broker is an interface used for asynchronous messaging.
	Broker[T any] interface {
		// Open establish connection to the target server.
		Open(ctx context.Context) error
		// Publish publish the message to the target topic.
		Publish(ctx context.Context, topic string, m *T, opts ...PublishOption[T]) error
		// Subscribe subscribe to the topic to consume messages.
		Subscribe(ctx context.Context, topic string, h func(Event[T]) error, opts ...SubscribeOption[T]) (Subscriber[T], error)
		// Close flush all in-flight messages and close underlying connection.
		// Close allows a context to control the duration
		// of a flush/close call. This context should be non-nil.
		// If a deadline is not set, a default deadline of 5s will be applied.
		Close(context.Context) error
	}

	// Event is given to a subscription handler for processing
	Event[T any] interface {
		Topic() string
		Message() *T
		Ack() error
		Error() error
		Reason() Reason
	}

	Reason int

	// Subscriber is a convenience return type for the Subscribe method
	Subscriber[T any] interface {
		Topic() string
		Unsubscribe() error
	}
)

const (
	ReasonUnmarshalFailure Reason = iota
	ReasonSubscriptionFailure
)
