package kafka

import (
	"github.com/IBM/sarama"
)

type Option[T any] func(*Broker[T])

func Config[T any](conf *sarama.Config) Option[T] {
	return func(b *Broker[T]) {
		b.conf = conf
	}
}

func Address[T any](addrs []string) Option[T] {
	return func(b *Broker[T]) {
		b.addrs = addrs
	}
}

func AsyncPublish[T any]() Option[T] {
	return func(b *Broker[T]) {
		b.async = true
	}
}

func OnAsyncPublishFailure[T any](f func(*PublishError)) Option[T] {
	return func(b *Broker[T]) {
		b.async = true
		b.onPublishFailure = f
		b.conf.Producer.Return.Errors = true
	}
}

func OnAsyncPublishSuccess[T any](f func(*T)) Option[T] {
	return func(b *Broker[T]) {
		b.async = true
		b.onPublishSuccess = f
		b.conf.Producer.Return.Successes = true
	}
}

func Codec[T any](c codec) Option[T] {
	return func(b *Broker[T]) {
		b.codec = c
	}
}

func Logger[T any](l logger) Option[T] {
	return func(b *Broker[T]) {
		b.log = l
	}
}

func SASL[T any](user string, pass string) Option[T] {
	return func(b *Broker[T]) {
		if b.conf == nil {
			b.conf = sarama.NewConfig()
		}
		b.conf.Net.SASL.Enable = true
		b.conf.Net.SASL.User = user
		b.conf.Net.SASL.Password = pass
	}
}
