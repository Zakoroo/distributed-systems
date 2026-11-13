package pubsub_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Zakoroo/distributed-systems/labs/lab1/src/pubsub"
)

func broker(t *testing.T, capacity int) *pubsub.PubSubService {
	t.Helper()
	s, err := pubsub.New(capacity)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func events(t *testing.T, s *pubsub.PubSubService, name string, id uint) []pubsub.Event {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := s.Events(ctx, name, id)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func TestFanoutOrderingAndPayloadOwnership(t *testing.T) {
	s := broker(t, 128)
	a, _ := s.Subscribe("news")
	b, _ := s.Subscribe("news")
	input := []byte("original")
	s.Publish("news", input)
	input[0] = 'X'
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 20 {
				if err := s.Publish("news", []byte("message")); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	first, second := events(t, s, "news", a), events(t, s, "news", b)
	if len(first) != 81 || len(second) != 81 {
		t.Fatalf("fanout lengths %d/%d", len(first), len(second))
	}
	for i := range first {
		if first[i].Sequence() != uint64(i+1) || second[i].Sequence() != first[i].Sequence() {
			t.Fatal("out-of-order delivery")
		}
	}
	if string(first[0].Data()) != "original" {
		t.Fatal("retained caller's mutable bytes")
	}
	data := first[0].Data()
	data[0] = 'Y'
	if string(first[0].Data()) != "original" || string(second[0].Data()) != "original" {
		t.Fatal("accessor exposed mutable storage")
	}
}
func TestBoundedQueuesAndSubscriptionWindow(t *testing.T) {
	s := broker(t, 1)
	s.Publish("news", []byte("before subscription"))
	a, _ := s.Subscribe("news")
	s.Publish("news", []byte("first"))
	b, _ := s.Subscribe("news")
	s.Publish("news", []byte("second")) // A is full; B must still receive it.
	if got := events(t, s, "news", a); len(got) != 1 || string(got[0].Data()) != "first" {
		t.Fatal("queue bound violated")
	}
	if got := events(t, s, "news", b); len(got) != 1 || string(got[0].Data()) != "second" {
		t.Fatal("slow subscriber blocked another")
	}
	s.Unsubscribe("news", a)
	if _, err := s.Events(context.Background(), "news", a); !errors.Is(err, pubsub.ErrUnknownSubscriber) {
		t.Fatal(err)
	}
}
func TestWaitingPollUnsubscribeAndConcurrentPoll(t *testing.T) {
	s := broker(t, 2)
	id, _ := s.Subscribe("news")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result := make(chan error, 2)
	for range 2 {
		go func() { _, err := s.Events(ctx, "news", id); result <- err }()
	}
	select {
	case err := <-result:
		if !errors.Is(err, pubsub.ErrAlreadyPolling) {
			t.Fatalf("concurrent poll: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("polls did not coordinate")
	}
	if err := s.Unsubscribe("news", id); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-result:
		if !errors.Is(err, pubsub.ErrUnknownSubscriber) {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("unsubscribe did not wake poll")
	}
}
func TestTimeoutAndValidation(t *testing.T) {
	if _, err := pubsub.New(0); err == nil {
		t.Fatal("accepted zero capacity")
	}
	s := broker(t, 1)
	if _, err := s.Subscribe(""); !errors.Is(err, pubsub.ErrInvalidTopic) {
		t.Fatal(err)
	}
	id, _ := s.Subscribe("news")
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if _, err := s.Events(ctx, "news", id); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	s.Publish("news", []byte("after timeout"))
	events(t, s, "news", id)
}

func TestConcurrentMembershipPublicationAndPolling(t *testing.T) {
	s := broker(t, 16)
	id, _ := s.Subscribe("shared")
	s.Publish("shared", []byte("initial"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		var last uint64
		for {
			batch, err := s.Events(ctx, "shared", id)
			if err != nil {
				if ctx.Err() == nil {
					t.Error(err)
				}
				return
			}
			for _, event := range batch {
				if event.Sequence() <= last {
					t.Error("poll observed non-increasing sequence")
				}
				if last == 0 {
					close(started)
				}
				last = event.Sequence()
			}
		}
	}()
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("initial poll did not receive message")
	}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 30 {
				temporary, err := s.Subscribe("shared")
				if err != nil {
					t.Error(err)
					return
				}
				if err := s.Publish("shared", []byte("concurrent")); err != nil {
					t.Error(err)
				}
				if err := s.Unsubscribe("shared", temporary); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	cancel()
	<-done
}
