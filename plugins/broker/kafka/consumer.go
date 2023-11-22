package kafka

import (
	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
)

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler.
type consumerGroupHandler struct {
	handler  broker.Handler
	opts     broker.SubscribeOptions
	codec    codec
	log      logger
	consumer sarama.ConsumerGroup
}

func (*consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (*consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var m broker.Message
		e := &event{
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
		if e.m.Body == nil {
			e.m.Body = msg.Value
		}
		if m.Header == nil {
			m.Header = make(map[string]string)
		}
		for _, header := range msg.Headers {
			m.Header[string(header.Key)] = string(header.Value)
		}
		if err := h.handler(e); err == nil && h.opts.AutoAck {
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
