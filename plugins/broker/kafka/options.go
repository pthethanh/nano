package kafka

import (
	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
)

type Option func(*Broker)

func Config(conf *sarama.Config) Option {
	return func(b *Broker) {
		b.conf = conf
	}
}

func Address(addrs []string) Option {
	return func(b *Broker) {
		b.addrs = addrs
	}
}

func AsyncPublish() Option {
	return func(b *Broker) {
		b.async = true
	}
}

func OnAsyncPublishFailure(f func(*PublishError)) Option {
	return func(b *Broker) {
		b.async = true
		b.onPublishFailure = f
		b.conf.Producer.Return.Errors = true
	}
}

func OnAsyncPublishSuccess(f func(*broker.Message)) Option {
	return func(b *Broker) {
		b.async = true
		b.onPublishSuccess = f
		b.conf.Producer.Return.Successes = true
	}
}

func Codec(c codec) Option {
	return func(b *Broker) {
		b.codec = c
	}
}

func Logger(l logger) Option {
	return func(b *Broker) {
		b.log = l
	}
}
