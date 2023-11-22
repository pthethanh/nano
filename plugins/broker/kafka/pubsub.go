package kafka

import (
	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
)

type subscriber struct {
	broker   *Broker
	consumer sarama.ConsumerGroup
	t        string
	opts     broker.SubscribeOptions
}

type event struct {
	topic    string
	err      error
	consumer sarama.ConsumerGroup
	msg      *sarama.ConsumerMessage
	m        *broker.Message
	session  sarama.ConsumerGroupSession
	reason   broker.Reason
}

func (p *event) Topic() string {
	return p.topic
}

func (p *event) Message() *broker.Message {
	return p.m
}

func (p *event) Ack() error {
	p.session.MarkMessage(p.msg, "")
	return nil
}

func (p *event) Error() error {
	return p.err
}

func (p *event) Reason() broker.Reason {
	return p.reason
}

func (s *subscriber) Options() broker.SubscribeOptions {
	return s.opts
}

func (s *subscriber) Topic() string {
	return s.t
}

func (s *subscriber) Unsubscribe() error {
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
