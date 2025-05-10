package subpub

import (
	"context"
	"errors"
	"log"
	"sync"
)

// MessageHandler is a callback function that process massages delivered to subscribers.
type MessageHandler func(msg interface{})

type Subscription interface {
	// Unsubscribe will remove interest in the current subject subscription is for.
	Unsubscribe()
}

type SubPub interface {
	// Subscribe creates an asynchronous queue subscribers on the given subject.
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	// Publish publishes the msg argument to the give subject.
	Publish(subject string, msg interface{}) error

	// Close will shutdown the sub-pub system.
	// May be blocked by data deliver until the context is canceled.
	Close(ctx context.Context) error
}

type subscription struct {
	ch     chan interface{}
	cb     MessageHandler
	subpub *subPub
	closed bool
	mu     sync.Mutex
}

type subPub struct {
	mu      sync.Mutex
	subs    map[string][]*subscription
	closed  bool
	subject string
	wg      sync.WaitGroup
}

func NewSubPub() SubPub {
	return &subPub{
		subs: make(map[string][]*subscription),
	}
}

func (s *subscription) Unsubscribe() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	s.subpub.mu.Lock()
	defer s.subpub.mu.Unlock()
	for i, sub := range s.subpub.subs[s.subpub.subject] {
		if sub == s {
			s.subpub.subs[s.subpub.subject] = append(s.subpub.subs[s.subpub.subject][:i], s.subpub.subs[s.subpub.subject][i+1:]...)
			break
		}
	}
	close(s.ch) // Закрываем канал только здесь
}

func (sp *subPub) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	if sp.closed {
		return nil, errors.New("subpub is closed")
	}
	sub := &subscription{
		ch:     make(chan interface{}, 100),
		cb:     cb,
		subpub: sp,
	}
	sp.subs[subject] = append(sp.subs[subject], sub)
	sp.wg.Add(1)
	go func() {
		defer sp.wg.Done()
		for msg := range sub.ch {
			cb(msg)
		}
	}()
	return sub, nil
}

func (sp *subPub) Publish(subject string, msg interface{}) error {
	sp.mu.Lock()
	subs, ok := sp.subs[subject]
	sp.mu.Unlock()
	if !ok {
		return nil
	}
	for _, sub := range subs {
		select {
		case sub.ch <- msg:
		default:
			log.Println("subpub buffer is full")
		}
	}
	return nil
}

func (sp *subPub) Close(ctx context.Context) error {
	sp.mu.Lock()
	if sp.closed {
		sp.mu.Unlock()
		return nil
	}
	sp.closed = true

	// Закрываем только незакрытые каналы подписок
	for subject, subs := range sp.subs {
		for _, sub := range subs {
			sub.mu.Lock()
			if !sub.closed {
				sub.closed = true
				close(sub.ch)
			}
			sub.mu.Unlock()
		}
		delete(sp.subs, subject)
	}
	sp.mu.Unlock()

	// Ждем завершения всех горутин
	done := make(chan struct{})
	go func() {
		sp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
