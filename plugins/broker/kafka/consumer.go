package kafka

import (
	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
)

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler.
type consumerGroupHandler[T any] struct {
	handler  func(broker.Event[T]) error
	opts     broker.SubscribeOptions
	codec    broker.Codec
	log      logger
	consumer sarama.ConsumerGroup
}

func (*consumerGroupHandler[T]) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*consumerGroupHandler[T]) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m T
		e := &event[T]{
			m:        &m,
			topic:    msg.Topic,
			msg:      msg,
			consumer: h.consumer,
			session:  session,
		}
		if err := h.codec.Unmarshal(msg.Value, &m); err != nil {
			e.err = err
			e.reason = broker.ReasonUnmarshalFailure
		}
		if err := h.handler(e); err == nil && h.opts.AutoAck {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
