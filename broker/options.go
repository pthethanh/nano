package broker

type (
	// PublishOptions holds configuration for publishing messages.
	PublishOptions struct {
	}

	// SubscribeOptions holds configuration for subscribing to messages.
	// Fields:
	//   AutoAck: If true (default), messages are automatically acknowledged when the handler returns nil error.
	//   Queue: Subscribers with the same queue name will share the subscription and receive a subset of messages.
	SubscribeOptions struct {
		AutoAck bool   // If true, automatically ack messages on successful handler execution.
		Queue   string // Name of the queue for shared subscriptions.
	}

	// PublishOption defines a function that configures PublishOptions.
	PublishOption func(*PublishOptions)

	// SubscribeOption defines a function that configures SubscribeOptions.
	SubscribeOption func(*SubscribeOptions)
)

// Queue sets the queue name for shared subscriptions.
// Subscribers with the same queue name will receive a subset of messages.
func Queue(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Queue = name
	}
}

// DisableAutoAck disables automatic acknowledgment of messages
// after they have been handled by the subscriber.
func DisableAutoAck() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = false
	}
}

// Apply applies a list of SubscribeOption functions to the SubscribeOptions receiver.
func (op *SubscribeOptions) Apply(opts ...SubscribeOption) {
	for _, f := range opts {
		f(op)
	}
}
