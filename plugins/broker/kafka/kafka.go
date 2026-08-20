// Package kafka provides a kafka broker using sarama cluster
package kafka

import (
	"context"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/pthethanh/nano/broker"
)

// Broker is an implementation of broker.Broker using Kafka (via sarama).
type Broker[T any] struct {
	addrs            []string
	conf             *sarama.Config
	codec            broker.Codec[T]
	log              logger
	async            bool
	onPublishFailure func(*PublishError[T])
	onPublishSuccess func(*T)

	client         sarama.Client
	syncProducer   sarama.SyncProducer
	asyncProducer  sarama.AsyncProducer
	consumerGroups []sarama.ConsumerGroup
	mu             sync.Mutex
}

var (
	_ broker.Broker[any] = (*Broker[any])(nil)
)

// New returns a new Kafka message broker.
func New[T any](opts ...Option[T]) *Broker[T] {
	k := &Broker[T]{
		conf:  sarama.NewConfig(),
		log:   slog.Default(),
		codec: JSONCodec[T]{},
		addrs: []string{"127.0.0.1:9092"},
	}
	for _, o := range opts {
		o(k)
	}
	return k
}

// Open connects to the target Kafka cluster and starts a sync or async
// producer, depending on the AsyncPublish option.
func (k *Broker[T]) Open(context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.client != nil {
		return nil
	}
	if k.async {
		c, err := sarama.NewClient(k.addrs, k.conf)
		if err != nil {
			return err
		}
		p, err := sarama.NewAsyncProducerFromClient(c)
		if err != nil {
			return err
		}
		if k.conf.Producer.Return.Errors {
			go func() {
				f := func(*sarama.ProducerError) {}
				if k.onPublishFailure != nil {
					f = func(err *sarama.ProducerError) {
						k.onPublishFailure(publishErrorFrom[T](err))
					}
				}
				for err := range p.Errors() {
					f(err)
				}
			}()
		}
		if k.conf.Producer.Return.Successes {
			go func() {
				f := func(*sarama.ProducerMessage) {}
				if k.onPublishSuccess != nil {
					f = func(pm *sarama.ProducerMessage) {
						if pm.Metadata != nil {
							if msg, ok := pm.Metadata.(*T); ok {
								k.onPublishSuccess(msg)
							}
						}
					}
				}
				for m := range p.Successes() {
					f(m)
				}
			}()
		}
		k.client = c
		k.asyncProducer = p
	} else {
		// SyncProducer requires errors & successes are set to true
		k.conf.Producer.Return.Successes = true
		k.conf.Producer.Return.Errors = true
		c, err := sarama.NewClient(k.addrs, k.conf)
		if err != nil {
			return err
		}
		p, err := sarama.NewSyncProducerFromClient(c)
		if err != nil {
			return err
		}
		k.client = c
		k.syncProducer = p
	}
	k.consumerGroups = make([]sarama.ConsumerGroup, 0)
	k.log.Log(context.Background(), slog.LevelInfo, "connected", "address", k.addrs, "async", k.async)
	return nil
}

// Publish implements broker.Broker interface.
func (k *Broker[T]) Publish(ctx context.Context, topic string, msg *T, opts ...broker.PublishOption) error {
	var popts broker.PublishOptions
	popts.Apply(opts...)

	b, err := k.codec.Marshal(msg)
	if err != nil {
		return err
	}
	m := &sarama.ProducerMessage{
		Topic:    topic,
		Value:    sarama.ByteEncoder(b),
		Headers:  recordHeadersFrom(popts.Headers),
		Metadata: msg,
	}
	if k.async {
		k.asyncProducer.Input() <- m
		return nil
	} else {
		_, _, err = k.syncProducer.SendMessage(m)
		return err
	}
}

// Subscribe implements broker.Broker interface. Each call creates its own
// consumer group (defaulting to a random group ID via broker.Queue).
func (k *Broker[T]) Subscribe(ctx context.Context, topic string, handler func(broker.Event[T]) error, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	opt := broker.SubscribeOptions{
		AutoAck: true,
		Queue:   uuid.New().String(),
	}
	opt.Apply(opts...)
	consumer, err := k.newConsumerGroup(opt.Queue)
	if err != nil {
		return nil, err
	}
	consumerHandler := &consumerGroupHandler[T]{
		handler:  handler,
		opts:     opt,
		codec:    k.codec,
		log:      k.log,
		consumer: consumer,
	}
	topics := []string{topic}
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				if err != nil {
					handler(&event[T]{
						topic:  topic,
						err:    err,
						reason: broker.ReasonSubscriptionFailure,
					})
				}
			default:
				err := consumer.Consume(ctx, topics, consumerHandler)
				switch err {
				case nil:
					// everything is ok, continue
					continue
				case sarama.ErrClosedConsumerGroup:
					return
				default:
					// report error to handler
					handler(&event[T]{
						err:    err,
						topic:  topic,
						reason: broker.ReasonSubscriptionFailure,
					})
				}
			}
		}
	}()
	k.log.Log(ctx, slog.LevelInfo, "subscribed successfully", "topic", topic, "queue", opt.Queue)
	return &subscriber[T]{
		broker:   k,
		consumer: consumer,
		opts:     opt,
		t:        topic,
	}, nil
}

func (k *Broker[T]) newConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	cg, err := sarama.NewConsumerGroup(k.addrs, groupID, k.conf)
	if err != nil {
		return nil, err
	}
	k.mu.Lock()
	defer k.mu.Unlock()
	k.consumerGroups = append(k.consumerGroups, cg)
	return cg, nil
}

// String returns the broker name.
func (k *Broker[T]) String() string {
	return "kafka"
}

// Close flushes and closes producers and consumer groups, and closes the
// underlying client connection.
func (k *Broker[T]) Close(ctx context.Context) error {
	k.log.Log(ctx, slog.LevelInfo, "closing")
	k.mu.Lock()
	defer k.mu.Unlock()
	if k.client == nil {
		// Open() was never called (or never succeeded): nothing to close.
		return nil
	}
	for _, consumer := range k.consumerGroups {
		consumer.Close()
	}
	k.consumerGroups = nil
	if k.syncProducer != nil {
		k.syncProducer.Close()
	}
	if k.asyncProducer != nil {
		k.asyncProducer.Close()
	}
	if err := k.client.Close(); err != nil {
		return err
	}
	return nil
}
