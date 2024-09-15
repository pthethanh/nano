package broker

type (
	// PublishOptions is a configuration holder for publish options.
	PublishOptions[T any] struct {
	}

	// SubscribeOptions is a configuration holder for subscriptions.
	SubscribeOptions[T any] struct {
		// AutoAck defaults to true. When a handler returns
		// with a nil error the message is acked.
		AutoAck bool
		// Subscribers with the same queue name
		// will create a shared subscription where each
		// receives a subset of messages.
		Queue string
	}

	// PublishOption is a func for config publish options.
	PublishOption[T any] func(*PublishOptions[T])

	// SubscribeOption is a func for config subscription.
	SubscribeOption[T any] func(*SubscribeOptions[T])
)

// Queue sets the name of the queue to share messages on
func Queue[T any](name string) SubscribeOption[T] {
	return func(o *SubscribeOptions[T]) {
		o.Queue = name
	}
}

// DisableAutoAck will disable auto ack of messages
// after they have been handled.
func DisableAutoAck[T any]() SubscribeOption[T] {
	return func(o *SubscribeOptions[T]) {
		o.AutoAck = false
	}
}

// Apply apply the options.
func (op *SubscribeOptions[T]) Apply(opts ...SubscribeOption[T]) {
	for _, f := range opts {
		f(op)
	}
}
