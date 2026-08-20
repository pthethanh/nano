package broker

type (
	// PublishOptions holds configuration for publishing messages.
	PublishOptions struct {
		// Headers are transport-level key/value pairs attached to the
		// message, e.g. Kafka record headers, NATS message headers, or
		// Watermill message metadata. Support is transport-dependent; a
		// Broker implementation with no wire-level header concept (such as
		// the in-memory broker) ignores this field.
		Headers map[string]string
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

// Header sets a single transport-level header on the published message,
// alongside any already set by earlier options.
func Header(key, value string) PublishOption {
	return func(o *PublishOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string, 1)
		}
		o.Headers[key] = value
	}
}

// Headers sets multiple transport-level headers on the published message,
// merging with (and overriding on key conflict) any already set by earlier
// options.
func Headers(headers map[string]string) PublishOption {
	return func(o *PublishOptions) {
		if o.Headers == nil {
			o.Headers = make(map[string]string, len(headers))
		}
		for k, v := range headers {
			o.Headers[k] = v
		}
	}
}

// Apply applies a list of PublishOption functions to the PublishOptions receiver.
func (op *PublishOptions) Apply(opts ...PublishOption) {
	for _, f := range opts {
		f(op)
	}
}
