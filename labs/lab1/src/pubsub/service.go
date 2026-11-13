// Package pubsub implements an in-memory, ordered publish-subscribe broker.
package pubsub

import (
	"context"
	"errors"
	"github.com/Zakoroo/distributed-systems/concurrent/waiting"
	"sync"
)

var (
	ErrInvalidTopic      = errors.New("topic must not be empty")
	ErrUnknownSubscriber = errors.New("unknown subscriber")
	ErrAlreadyPolling    = errors.New("subscriber already has an events request")
)

// Event is an immutable message. Accessors do not expose the broker's storage.
type Event struct {
	sequence uint64
	data     []byte
}

func (e Event) Sequence() uint64 { return e.sequence }
func (e Event) Data() []byte     { return append([]byte{}, e.data...) }

type subscriber struct {
	events  chan Event
	done    chan struct{}
	polling bool
}
type topic struct {
	subscribers  map[uint]*subscriber
	nextSequence uint64
}

// PubSubService owns all topic and subscription state. Construct it with New;
// do not copy it after use.
type PubSubService struct {
	mu       sync.Mutex
	topics   map[string]*topic
	nextID   uint
	capacity int
}

func New(queueCapacity int) (*PubSubService, error) {
	if queueCapacity <= 0 {
		return nil, errors.New("queue capacity must be positive")
	}
	return &PubSubService{topics: make(map[string]*topic), capacity: queueCapacity}, nil
}

func (s *PubSubService) Subscribe(name string) (uint, error) {
	if name == "" {
		return 0, ErrInvalidTopic
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.topics[name]
	if t == nil {
		t = &topic{subscribers: make(map[uint]*subscriber)}
		s.topics[name] = t
	}
	s.nextID++
	id := s.nextID
	t.subscribers[id] = &subscriber{events: make(chan Event, s.capacity), done: make(chan struct{})}
	return id, nil
}

func (s *PubSubService) Unsubscribe(name string, id uint) error {
	if name == "" {
		return ErrInvalidTopic
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	sub := s.lookup(name, id)
	if sub == nil {
		return ErrUnknownSubscriber
	}
	delete(s.topics[name].subscribers, id)
	close(sub.done) // Only the service closes done; message channels stay open.
	return nil
}

// Publish never waits for subscriber queue space. Slow subscribers may miss events.
func (s *PubSubService) Publish(name string, data []byte) error {
	if name == "" {
		return ErrInvalidTopic
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := s.topics[name]
	if t == nil {
		return nil
	} // No subscribers and no retained history.
	t.nextSequence++
	event := Event{sequence: t.nextSequence, data: append([]byte(nil), data...)}
	for _, sub := range t.subscribers {
		select {
		case sub.events <- event:
		default:
		}
	}
	return nil
}

// Events waits for a message or context cancellation, then drains available messages.
// Only one events request per subscription is allowed, preserving delivery order.
func (s *PubSubService) Events(ctx context.Context, name string, id uint) ([]Event, error) {
	if name == "" {
		return nil, ErrInvalidTopic
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	sub := s.lookup(name, id)
	if sub == nil {
		s.mu.Unlock()
		return nil, ErrUnknownSubscriber
	}
	if sub.polling {
		s.mu.Unlock()
		return nil, ErrAlreadyPolling
	}
	sub.polling = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); sub.polling = false; s.mu.Unlock() }()

	var first Event
	err := waiting.Wait(ctx, func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-sub.done:
			return ErrUnknownSubscriber
		case first = <-sub.events:
			return nil
		}
	})
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lookup(name, id) != sub {
		return nil, ErrUnknownSubscriber
	}
	events := []Event{first}
	for {
		select {
		case event := <-sub.events:
			events = append(events, event)
		default:
			return events, nil
		}
	}
}

// lookup requires mu to be held.
func (s *PubSubService) lookup(name string, id uint) *subscriber {
	if t := s.topics[name]; t != nil {
		return t.subscribers[id]
	}
	return nil
}
