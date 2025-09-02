// Package memory provides a message broker using memory.
package memory

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/pthethanh/nano/broker"
)

type (
	topic = string
	queue = string
	id    = string
	// Broker is a memory message broker.
	Broker[T any] struct {
		subs   map[topic]map[queue]map[id]*subscriber[T]
		mu     *sync.RWMutex
		ch     chan func() error
		worker int
		buf    int
		wg     *sync.WaitGroup
		opened bool
	}

	subscriber[T any] struct {
		id     string
		t      string
		h      func(broker.Event[T]) error
		opts   *broker.SubscribeOptions
		closed int32
		close  func()
	}

	event[T any] struct {
		t      string
		msg    *T
		err    error
		reason broker.Reason
	}

	Option[T any] func(*Broker[T])
)

var (
	_ broker.Broker[any] = (*Broker[any])(nil)

	// ErrInvalidConnectionState indicate that the connection has not been opened properly.
	ErrInvalidConnectionState = errors.New("invalid connection state")
)

// New return new memory broker.
func New[T any](opts ...Option[T]) *Broker[T] {
	br := &Broker[T]{
		subs:   make(map[topic]map[queue]map[id]*subscriber[T]),
		mu:     &sync.RWMutex{},
		worker: 100,
		buf:    10_000,
		wg:     &sync.WaitGroup{},
	}
	for _, opt := range opts {
		opt(br)
	}
	br.ch = make(chan func() error, br.buf)
	return br
}

func (env *event[T]) Topic() string {
	return env.t
}

func (env *event[T]) Message() *T {
	return env.msg
}

func (env *event[T]) Ack() error {
	return nil
}

func (env *event[T]) Error() error {
	return env.err
}

func (env *event[T]) Reason() broker.Reason {
	return env.reason
}

// Topic implements broker.Subscriber interface.
func (sub *subscriber[T]) Topic() string {
	return sub.t
}

// Unsubscribe implements broker.Subscriber interface.
func (sub *subscriber[T]) Unsubscribe() error {
	if atomic.AddInt32(&sub.closed, 1) > 1 {
		return nil
	}
	return nil
}

func (sub *subscriber[T]) isClosed() bool {
	return atomic.LoadInt32(&sub.closed) > 0
}

// Open implements broker.Broker interface.
func (br *Broker[T]) Open(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(br.worker)
	br.wg.Add(br.worker)
	for i := 0; i < br.worker; i++ {
		go func() {
			wg.Done()
			defer br.wg.Done()
			for h := range br.ch {
				_ = h()
			}
		}()
	}
	wg.Wait()
	br.opened = true
	return nil
}

// Publish implements broker.Broker interface.
func (br *Broker[T]) Publish(ctx context.Context, topic string, m *T, opts ...broker.PublishOption) error {
	if !br.opened {
		return ErrInvalidConnectionState
	}
	br.mu.RLock()
	queueSubs := br.subs[topic]
	br.mu.RUnlock()
	env := &event[T]{
		t:   topic,
		msg: m,
	}
	for queue, queueSub := range queueSubs {
		switch queue {
		case "":
			// no queue, send to all subscribers in the list.
			for _, sub := range queueSub {
				if sub.isClosed() {
					continue
				}
				br.ch <- func() error { return sub.h(env) }
			}
		default:
			// queue, send to only 1 single random subscriber in the list.
			idx := rand.Intn(len(queueSub))
			i := 0
			for _, sub := range queueSub {
				if sub.isClosed() {
					continue
				}
				if i == idx {
					br.ch <- func() error { return sub.h(env) }
					break
				}
				i++
			}
		}
	}
	return nil
}

// Subscribe implements broker.Broker interface.
func (br *Broker[T]) Subscribe(ctx context.Context, topic string, h func(broker.Event[T]) error, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	if !br.opened {
		return nil, ErrInvalidConnectionState
	}
	subOpts := &broker.SubscribeOptions{}
	subOpts.Apply(opts...)
	newSub := &subscriber[T]{
		id:   uuid.New().String(),
		t:    topic,
		h:    h,
		opts: subOpts,
	}
	newSub.close = func() {
		br.mu.Lock()
		defer br.mu.Unlock()
		delete(br.subs[topic][newSub.opts.Queue], newSub.id)
	}
	br.mu.Lock()
	defer br.mu.Unlock()
	if br.subs[topic] == nil {
		br.subs[topic] = make(map[queue]map[id]*subscriber[T])
	}
	if br.subs[topic][newSub.opts.Queue] == nil {
		br.subs[topic][newSub.opts.Queue] = make(map[id]*subscriber[T])
	}
	br.subs[topic][newSub.opts.Queue][newSub.id] = newSub
	return newSub, nil
}

// CheckHealth implements health.Checker interface.
func (br *Broker[T]) CheckHealth(ctx context.Context) error {
	if !br.opened {
		return ErrInvalidConnectionState
	}
	return nil
}

// Close implements broker.Broker interface.
func (br *Broker[T]) Close(ctx context.Context) error {
	br.opened = false
	close(br.ch)
	br.wg.Wait()
	// unsubscribe all subscribers.
	for _, queue := range br.subs {
		for _, queue := range queue {
			for _, sub := range queue {
				sub.Unsubscribe()
			}
		}
	}
	return nil
}
