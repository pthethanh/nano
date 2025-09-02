package memory_test

import (
	"context"
	"fmt"
	"reflect"
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

func BenchmarkBroker(b *testing.B) {
	br := memory.New[string]()
	if err := br.Open(context.TODO()); err != nil {
		b.Fatal(err)
	}
	for j := range 3 {
		for i := 0; i < 3; i++ {
			sub, err := br.Subscribe(context.Background(), fmt.Sprintf("topic%d", j), func(e broker.Event[string]) error {
				return nil
			}, broker.Queue(fmt.Sprintf("queue%d", j)))
			if err != nil {
				b.Fatal(err)
			}
			defer sub.Unsubscribe()
		}
		for i := 0; i < 3; i++ {
			sub, err := br.Subscribe(context.Background(), fmt.Sprintf("topic%d", j), func(e broker.Event[string]) error {
				return nil
			})
			if err != nil {
				b.Fatal(err)
			}
			defer sub.Unsubscribe()
		}
	}
	m := "ha ha"
	b.Run("publish", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			if err := br.Publish(context.Background(), "topic1", &m); err != nil {
				b.Fatal(err)
			}
			if err := br.Publish(context.Background(), "topic2", &m); err != nil {
				b.Fatal(err)
			}
			if err := br.Publish(context.Background(), "topic3", &m); err != nil {
				b.Fatal(err)
			}
		}
	})
}
