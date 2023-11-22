package nats

import (
	"github.com/nats-io/nats.go"
)

// Codec is an option to provide a custom codec.
func Codec(codec codec) Option {
	return func(opts *Nats) {
		opts.codec = codec
	}
}

// Address is an option to set target addresses of NATS server.
// Multiple addresses are separated by comma.
func Address(addrs string) Option {
	return func(opts *Nats) {
		opts.addrs = addrs
	}
}

// Options is an option to provide additional nats.Option.
func Options(opts ...nats.Option) Option {
	return func(n *Nats) {
		n.opts = append(n.opts, opts...)
	}
}

// Logger is an option to provide custom logger.
func Logger(logger logger) Option {
	return func(opts *Nats) {
		opts.log = logger
	}
}
