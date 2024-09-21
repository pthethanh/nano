package kafka

import (
	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
)

type subscriber[T any] struct {
	broker   *Broker[T]
	consumer sarama.ConsumerGroup
	t        string
	opts     broker.SubscribeOptions
}

type event[T any] struct {
	topic    string
	err      error
	consumer sarama.ConsumerGroup
	msg      *sarama.ConsumerMessage
	m        *T
	session  sarama.ConsumerGroupSession
	reason   broker.Reason
}

func (p *event[T]) Topic() string {
	return p.topic
}

func (p *event[T]) Message() *T {
	return p.m
}

func (p *event[T]) Ack() error {
	p.session.MarkMessage(p.msg, "")
	return nil
}

func (p *event[T]) Error() error {
	return p.err
}

func (p *event[T]) Reason() broker.Reason {
	return p.reason
}

func (s *subscriber[T]) Options() broker.SubscribeOptions {
	return s.opts
}

func (s *subscriber[T]) Topic() string {
	return s.t
}

func (s *subscriber[T]) Unsubscribe() error {
	if err := s.consumer.Close(); err != nil {
		return err
	}

	k := s.broker
	k.mu.Lock()
	defer k.mu.Unlock()

	for i, cg := range k.consumerGroups {
		if cg == s.consumer {
			k.consumerGroups = append(k.consumerGroups[:i], k.consumerGroups[i+1:]...)
			return nil
		}
	}
	return nil
}
