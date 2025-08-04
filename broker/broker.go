// Package broker defines standard interface for a message broker.
package broker

import (
	"context"
)

type (
	// Broker is an interface used for asynchronous messaging.
	// It defines methods for connecting, publishing, subscribing, and closing the broker.
	//
	// T is the type of message handled by the broker.
	Broker[T any] interface {
		// Open establishes a connection to the target server.
		// Returns an error if the connection fails.
		Open(ctx context.Context) error

		// Publish sends the message m to the specified topic.
		// Additional PublishOption(s) can be provided to customize publishing behavior.
		// Returns an error if publishing fails.
		Publish(ctx context.Context, topic string, m *T, opts ...PublishOption) error

		// Subscribe registers a handler h to consume messages from the specified topic.
		// Additional SubscribeOption(s) can be provided to customize subscription behavior.
		// Returns a Subscriber for managing the subscription and an error if subscription fails.
		Subscribe(ctx context.Context, topic string, h func(Event[T]) error, opts ...SubscribeOption) (Subscriber, error)

		// Close flushes all in-flight messages and closes the underlying connection.
		// The provided context controls the duration of the flush/close operation.
		// If the context does not have a deadline, a default deadline of 5 seconds is applied.
		// Returns an error if closing fails.
		Close(context.Context) error
	}

	// Event is provided to a subscription handler for processing.
	// It represents a message event received from a topic.
	Event[T any] interface {
		// Topic returns the topic name of the event.
		Topic() string

		// Message returns the message payload of the event.
		Message() *T

		// Ack acknowledges successful processing of the event.
		// Returns an error if acknowledgment fails.
		Ack() error

		// Error returns any error encountered during event processing.
		Error() error

		// Reason returns the reason code for the event in case of error.
		Reason() Reason
	}

	// Reason represents the reason code for event errors.
	Reason int

	// Subscriber is a convenience return type for the Subscribe method.
	// It allows management of the subscription.
	Subscriber interface {
		// Topic returns the topic name of the subscription.
		Topic() string

		// Unsubscribe cancels the subscription.
		// Returns an error if unsubscription fails.
		Unsubscribe() error
	}
)

const (
	// ReasonUnmarshalFailure indicates a failure to unmarshal a message.
	ReasonUnmarshalFailure Reason = iota

	// ReasonSubscriptionFailure indicates a failure to subscribe to a topic.
	ReasonSubscriptionFailure
)
