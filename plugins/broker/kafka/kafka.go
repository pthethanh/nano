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

type Broker[T any] struct {
	addrs            []string
	conf             *sarama.Config
	codec            codec
	log              logger
	async            bool
	onPublishFailure func(*PublishError)
	onPublishSuccess func(*T)

	client         sarama.Client
	syncProducer   sarama.SyncProducer
	asyncProducer  sarama.AsyncProducer
	consumerGroups []sarama.ConsumerGroup
	connected      bool
	mu             sync.Mutex
}

var (
	_ broker.Broker[any] = (*Broker[any])(nil)
)

func New[T any](opts ...Option[T]) *Broker[T] {
	k := &Broker[T]{
		conf:  sarama.NewConfig(),
		log:   slog.Default(),
		codec: JSONCodec{},
		addrs: []string{"127.0.0.1:9092"},
	}
	for _, o := range opts {
		o(k)
	}
	return k
}

func (k *Broker[T]) Open(context.Context) error {
	if k.connected {
		return nil
	}
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
						pErr := &PublishError{
							Error: err.Err,
						}
						if err.Msg.Metadata != nil {
							if v, ok := err.Msg.Metadata.(*broker.Message); ok {
								pErr.Message = v
							}
						}
						k.onPublishFailure(pErr)
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
	k.connected = true
	k.log.Log(context.Background(), slog.LevelInfo, "connected", "address", k.addrs, "async", k.async)
	return nil
}

func (k *Broker[T]) Publish(ctx context.Context, topic string, msg *T, opts ...broker.PublishOption) error {
	b, err := k.codec.Marshal(msg)
	if err != nil {
		return err
	}
	m := &sarama.ProducerMessage{
		Topic:    topic,
		Value:    sarama.ByteEncoder(b),
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
						err:    err,
						reason: broker.ReasonSubscriptionFailure,
					})
				}
			default:
				err := consumer.Consume(ctx, topics, consumerHandler)
				switch err {
				case nil:
					continue
				case sarama.ErrClosedConsumerGroup:
					return
				default:
					handler(&event[T]{
						err:    err,
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

func (k *Broker[T]) String() string {
	return "kafka"
}

func (k *Broker[T]) Close(ctx context.Context) error {
	k.log.Log(ctx, slog.LevelInfo, "closing")
	k.mu.Lock()
	defer k.mu.Unlock()
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
	k.connected = false
	return nil
}
