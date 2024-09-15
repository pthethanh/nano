package nats

import (
	"github.com/nats-io/nats.go"
)

// Codec is an option to provide a custom codec.
func Codec[T any](codec codec) Option[T] {
	return func(opts *Nats[T]) {
		opts.codec = codec
	}
}

// Address is an option to set target addresses of NATS server.
// Multiple addresses are separated by comma.
func Address[T any](addrs string) Option[T] {
	return func(opts *Nats[T]) {
		opts.addrs = addrs
	}
}

// Options is an option to provide additional nats.Option.
func Options[T any](opts ...nats.Option) Option[T] {
	return func(n *Nats[T]) {
		n.opts = append(n.opts, opts...)
	}
}

// Logger is an option to provide custom logger.
func Logger[T any](logger logger) Option[T] {
	return func(opts *Nats[T]) {
		opts.log = logger
	}
}
