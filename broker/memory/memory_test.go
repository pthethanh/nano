package memory_test

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/pthethanh/nano/broker"
	"github.com/pthethanh/nano/broker/memory"
)

func TestBroker(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}
	b := memory.New(memory.Worker[Person](100, 1000))
	defer b.Close(context.Background())

	topic := "test"
	if err := b.Open(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	if err := b.CheckHealth(context.TODO()); err != nil {
		t.Fatalf("unexpected error: %v\n", err)
	}
	ch := make(chan broker.Event[Person], 100)
	// sub without group
	sub, err := b.Subscribe(context.Background(), topic, func(msg broker.Event[Person]) error {
		if err := msg.Ack(); err != nil {
			t.Error(err)
		}
		ch <- msg
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Unsubscribe()
	if sub.Topic() != topic {
		t.Errorf("got topic=%s, want topic=%s", sub.Topic(), topic)
	}
	// sub on the queue q1
	subGroup1, err := b.Subscribe(context.Background(), topic, func(msg broker.Event[Person]) error {
		ch <- msg
		return nil
	}, broker.Queue("q1"))
	if err != nil {
		t.Fatal(err)
	}
	defer subGroup1.Unsubscribe()
	// sub with the same group as the previous one - queue q1
	subGroup2, err := b.Subscribe(context.Background(), topic, func(msg broker.Event[Person]) error {
		ch <- msg
		return nil
	}, broker.Queue("q1"))
	if err != nil {
		t.Fatal(err)
	}
	defer subGroup2.Unsubscribe()
	want := Person{
		Name: "jack",
		Age:  22,
	}
	// send n messages
	n := 2
	for i := 0; i < n; i++ {
		if err := b.Publish(context.Background(), topic, &want); err != nil {
			t.Fatal(err)
		}
	}
	// send another message to a topic no one subscribe should not impact to the result.
	if err := b.Publish(context.Background(), "other-topic", &want); err != nil {
		t.Fatal(err)
	}
	cnt := 0
	timeout := time.After(50 * time.Millisecond)
	for {
		select {
		case e := <-ch:
			cnt++
			if e.Topic() != topic {
				t.Fatalf("got topic=%s, want topic=test", e.Topic())
			}
			got := e.Message()
			if got == nil || !reflect.DeepEqual(*got, want) {
				t.Errorf("got=%v, want=%v", got, want)
			}
		case <-timeout:
			// should received n*2 messages: sub get n, subGroup1 + subGroup2 = n
			if cnt != n*2 {
				t.Fatalf("got number of messages=%d, want number of messages=%d", cnt, n*2)
			}
			return
		}
	}
}

func runBenchmark(b *testing.B, br broker.Broker[string], topic string, queue string, numBroadcastSubs, numQueueSubs int) {
	ctx := context.Background()
	if err := br.Open(ctx); err != nil {
		b.Fatalf("failed to open broker: %v", err)
	}
	defer br.Close(ctx)

	var wg sync.WaitGroup
	wg.Add(b.N * numBroadcastSubs)
	if numQueueSubs > 0 {
		wg.Add(b.N)
	}
	// Subscribers
	for range numBroadcastSubs {
		_, err := br.Subscribe(ctx, topic, func(evt broker.Event[string]) error {
			wg.Done()
			return nil
		})
		if err != nil {
			b.Fatalf("failed to subscribe: %v", err)
		}
	}
	for range numQueueSubs {
		_, err := br.Subscribe(ctx, topic, func(evt broker.Event[string]) error {
			wg.Done()
			return nil
		}, broker.Queue(queue))
		if err != nil {
			b.Fatalf("failed to subscribe: %v", err)
		}
	}
	// Start benchmark
	b.ResetTimer()
	start := time.Now()

	for i := 0; i < b.N; i++ {
		msg := fmt.Sprintf("msg-%d", i)
		if err := br.Publish(ctx, topic, &msg); err != nil {
			b.Fatalf("publish failed: %v", err)
		}
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		b.Fatalf("timeout waiting for messages")
	}

	elapsed := time.Since(start)
	b.ReportMetric(float64(b.N)/elapsed.Seconds(), "msg/sec")
	b.ReportMetric(float64(elapsed.Microseconds())/float64(b.N), "µs/msg")
}

func BenchmarkBroker_Queue_Mixed_100(b *testing.B) {
	// Queue semantics: 1 subscriber → each message consumed once
	broker := memory.New[string]()
	runBenchmark(b, broker, "queue_topic", "q1", 50, 50)
}

func BenchmarkBroker_FanOut_10(b *testing.B) {
	// Fan-out semantics: 10 subscribers → each message delivered to all subscribers
	broker := memory.New[string]()
	runBenchmark(b, broker, "fan_out_topic", "", 10, 0)
}

func BenchmarkBroker_FanOut_50(b *testing.B) {
	// Stress test with 50 subscribers
	broker := memory.New[string]()
	runBenchmark(b, broker, "fan_out_topic_50", "", 50, 0)
}

func BenchmarkBroker_FanOut_100(b *testing.B) {
	// Stress test with 50 subscribers
	broker := memory.New[string]()
	runBenchmark(b, broker, "fan_out_topic_100", "", 100, 0)
}
