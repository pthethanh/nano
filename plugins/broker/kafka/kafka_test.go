package kafka_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/log"
	"github.com/pthethanh/nano/plugins/broker/kafka"
)

type testMsg struct {
	ID   string
	Data string
}

func TestKafkaBroker_PublishSubscribe(t *testing.T) {
	if os.Getenv("KAFKA_TEST") == "" {
		t.Skip("Set KAFKA_TEST=1 to run this test (requires local Kafka on :9092)")
	}
	log.Info("starting kafka broker test")
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.Timeout = 10 * time.Second
	cfg.Consumer.Group.Rebalance.Retry.Max = 3
	cfg.Consumer.Group.Rebalance.Retry.Backoff = 100 * time.Millisecond
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	log.Info("kafka config", "cfg", cfg)
	b := kafka.New(
		kafka.Address[testMsg]([]string{"localhost:9092"}),
		kafka.AsyncPublish[testMsg](),
		kafka.Config[testMsg](cfg),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.Open(ctx); err != nil {
		t.Fatalf("failed to open broker: %v", err)
	}
	defer b.Close(ctx)
	topic := "test-kafka-broker"
	received := make(chan *testMsg, 1)

	_, err := b.Subscribe(ctx, topic, func(ev broker.Event[testMsg]) error {
		if ev.Error() != nil {
			t.Errorf("event error: %v, reason: %v", ev.Error(), ev.Reason())
			return ev.Error()
		}
		received <- ev.Message()
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe failed: %v", err)
	}
	log.Info("subscribed to topic", "topic", topic)
	time.Sleep(5 * time.Second) // wait for consumer to be ready
	msg := &testMsg{ID: "1", Data: "hello kafka"}
	if err := b.Publish(ctx, topic, msg); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	log.Info("message published", "msg", msg)
	select {
	case got := <-received:
		log.Info("message received", "msg", got)
		if got.ID != msg.ID || got.Data != msg.Data {
			t.Fatalf("received message does not match: got %+v, want %+v", got, msg)
		}
	case <-ctx.Done():
		t.Fatal("did not receive message in time")
	}
}
