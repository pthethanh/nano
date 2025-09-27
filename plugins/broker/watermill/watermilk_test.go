package watermill

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/log"
)

type testMsg struct {
	ID   string
	Data string
}

func TestWatermillBrokerWithKafka(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := "test-watermill-broker"

	pub, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers:   brokers,
			Marshaler: kafka.DefaultMarshaler{},
		}, watermill.NewSlogLogger(log.Default().Logger),
	)
	if err != nil {
		t.Fatalf("failed to create publisher: %v", err)
	}
	defer pub.Close()

	sub, err := kafka.NewSubscriber(
		kafka.SubscriberConfig{
			Brokers:       brokers,
			Unmarshaler:   kafka.DefaultMarshaler{},
			ConsumerGroup: "",
			InitializeTopicDetails: &sarama.TopicDetail{
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		}, watermill.NewSlogLogger(log.Default().Logger))
	if err != nil {
		t.Fatalf("failed to create subscriber: %v", err)
	}
	defer sub.Close()

	b := New[testMsg](pub, sub)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	received := make(chan *testMsg, 1)

	_, err = b.Subscribe(ctx, topic, func(ev broker.Event[testMsg]) error {
		received <- ev.Message()
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	msg := &testMsg{ID: "1", Data: "hello watermill"}
	log.Info("message publishing...", "msg", msg)
	if err := b.Publish(ctx, topic, msg); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	log.Info("message published", "msg", msg)
	c := int64(0)
	select {
	case got := <-received:
		if got.ID != msg.ID || got.Data != msg.Data {
			t.Fatalf("----------------------------->>>>>>received message does not match: got %+v, want %+v", got, msg)
		}
		atomic.AddInt64(&c, 1)
	case <-ctx.Done():
		t.Fatal("----------------------------->>>>>>did not receive message in time")
	}
	log.Info("test completed", "count", c)
}
