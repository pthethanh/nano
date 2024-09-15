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
	// Broker is a memory message broker.
	Broker[T any] struct {
		subs   map[string][]*subscriber[T]
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
		opts   *broker.SubscribeOptions[T]
		close  func()
		closed int32
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
		subs:   make(map[string][]*subscriber[T]),
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
	sub.close()
	return nil
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
func (br *Broker[T]) Publish(ctx context.Context, topic string, m *T, opts ...broker.PublishOption[T]) error {
	if !br.opened {
		return ErrInvalidConnectionState
	}
	br.mu.RLock()
	subs := br.subs[topic]
	br.mu.RUnlock()
	// queue, list of sub
	queueSubs := make(map[string][]*subscriber[T])
	env := &event[T]{
		t:   topic,
		msg: m,
	}
	for _, sub := range subs {
		sub := sub
		if sub.opts.Queue != "" {
			queueSubs[sub.opts.Queue] = append(queueSubs[sub.opts.Queue], sub)
			continue
		}
		// broad cast
		br.ch <- func() error { return sub.h(env) }
	}
	// queue subscribers, send to only 1 single random subscriber in the list.
	for _, queueSub := range queueSubs {
		queueSub := queueSub
		idx := rand.Intn(len(queueSub))
		br.ch <- func() error { return queueSub[idx].h(env) }
	}
	return nil
}

// Subscribe implements broker.Broker interface.
func (br *Broker[T]) Subscribe(ctx context.Context, topic string, h func(broker.Event[T]) error, opts ...broker.SubscribeOption[T]) (broker.Subscriber[T], error) {
	if !br.opened {
		return nil, ErrInvalidConnectionState
	}
	subOpts := &broker.SubscribeOptions[T]{}
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
		subs := br.subs[topic]
		// remove the sub
		newSubs := make([]*subscriber[T], 0)
		for _, sub := range subs {
			if newSub.id == sub.id {
				continue
			}
			newSubs = append(newSubs, sub)
		}
		br.subs[topic] = newSubs
	}
	br.mu.Lock()
	defer br.mu.Unlock()
	br.subs[topic] = append(br.subs[topic], newSub)
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
	for _, subs := range br.subs {
		for _, sub := range subs {
			sub.Unsubscribe()
		}
	}
	return nil
}
