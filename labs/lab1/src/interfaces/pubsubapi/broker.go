// Package pubsubapi implements Lab 1's small in-memory message broker and its
// four HTTP handlers. Start with Subscribe, then Publish, then Events.
package pubsubapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Zakoroo/distributed-systems/httpserver"
)

// Broker stores subscriptions and messages. All state is private. Construct it
// with New; do not copy it after use. The four handlers may run concurrently.
// Each handler runs in the HTTP server's existing request goroutine.
type Broker struct {
	mu            sync.Mutex // Protects maps, counters, queue sends/closes, and polling flags.
	topics        map[string]*topic
	nextID        uint
	queueCapacity int
	pollTimeout   time.Duration
}

type topic struct {
	sequence    uint64
	subscribers map[uint]*subscriber
}

type subscriber struct {
	messages chan event // A bounded FIFO queue; receiving removes one message.
	polling  bool       // Only one handler may wait on this queue at a time.
}

type event struct {
	sequence uint64
	data     []byte // Owned by the broker and never changed after publication.
}

// New creates an empty broker. Capacity is per subscriber; timeout is per poll.
// Neither creating the broker nor subscribing starts a background goroutine.
func New(queueCapacity int, pollTimeout time.Duration) (*Broker, error) {
	if queueCapacity <= 0 || pollTimeout <= 0 {
		return nil, errors.New("queue capacity and polling timeout must be positive")
	}
	return &Broker{
		topics:        make(map[string]*topic),
		queueCapacity: queueCapacity,
		pollTimeout:   pollTimeout,
	}, nil
}

// HandleSubscribe creates a queue for future publications and returns its ID.
func (b *Broker) HandleSubscribe(req *http.Request) *http.Response {
	if req.Method != http.MethodPost {
		return response(http.StatusNotImplemented)
	}
	name := req.URL.Query().Get("topic")
	if name == "" {
		return response(http.StatusBadRequest)
	}

	b.mu.Lock()
	currentTopic := b.topics[name]
	if currentTopic == nil {
		currentTopic = &topic{subscribers: make(map[uint]*subscriber)}
		b.topics[name] = currentTopic
	}
	b.nextID++
	id := b.nextID
	currentTopic.subscribers[id] = &subscriber{messages: make(chan event, b.queueCapacity)}
	b.mu.Unlock()

	return jsonResponse(http.StatusCreated, map[string]uint{"id": id})
}

// HandleUnsubscribe removes membership and closes the queue to wake a waiting
// receiver. Publication uses this same mutex, so it cannot send to a closed queue.
func (b *Broker) HandleUnsubscribe(req *http.Request) *http.Response {
	if req.Method != http.MethodPost {
		return response(http.StatusNotImplemented)
	}
	name, id, valid := subscription(req)
	if !valid {
		return response(http.StatusBadRequest)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	current := b.findSubscriber(name, id)
	if current == nil {
		return response(http.StatusNotFound)
	}
	delete(b.topics[name].subscribers, id)
	close(current.messages)
	return response(http.StatusOK)
}

// HandlePublish offers the raw HTTP body to every current subscriber. A full
// queue skips this new message; it never makes publication wait for queue space.
func (b *Broker) HandlePublish(req *http.Request) *http.Response {
	if req.Method != http.MethodPost {
		return response(http.StatusNotImplemented)
	}
	name := req.URL.Query().Get("topic")
	if name == "" {
		return response(http.StatusBadRequest)
	}
	data := []byte{}
	if req.Body != nil {
		var err error
		data, err = io.ReadAll(req.Body)
		if err != nil {
			return response(http.StatusBadRequest)
		}
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	currentTopic := b.topics[name]
	if currentTopic == nil {
		return response(http.StatusOK) // No subscribers and no replay history.
	}
	currentTopic.sequence++
	// ReadAll gave us our own bytes. No code changes them afterward, so queues
	// can share this payload. The public API exposes only its JSON encoding.
	message := event{sequence: currentTopic.sequence, data: data}
	for _, current := range currentTopic.subscribers {
		select {
		case current.messages <- message:
		default:
			// This subscriber is full. Continue to the other subscribers.
		}
	}
	return response(http.StatusOK)
}

// HandleEvents waits for one message or a timeout. Clients must finish reading
// this response before starting another poll for the same ID. Overlapping waits
// return 400. Waiting still occupies one of the server's ten request handlers.
func (b *Broker) HandleEvents(req *http.Request) *http.Response {
	if req.Method != http.MethodGet {
		return response(http.StatusNotImplemented)
	}
	name, id, valid := subscription(req)
	if !valid {
		return response(http.StatusBadRequest)
	}

	// 1. Find the queue and claim its reader, then unlock before waiting.
	b.mu.Lock()
	current := b.findSubscriber(name, id)
	if current == nil {
		b.mu.Unlock()
		return response(http.StatusNotFound)
	}
	if current.polling {
		b.mu.Unlock()
		return response(http.StatusBadRequest)
	}
	current.polling = true
	b.mu.Unlock()

	// 2. The SAME request goroutine waits here. No mutex is held, so another
	// admitted request can publish or unsubscribe. No extra worker is started.
	var message event
	var received bool
	select {
	case message, received = <-current.messages:
		// received is false when the closed queue has no messages left.
	case <-time.After(b.pollTimeout):
		// No event selected: return 204 after the membership check below.
	}

	// 3. End the read and recheck membership: unsubscribe may have happened even
	// if select received an older buffered message from the now-closed queue.
	b.mu.Lock()
	current.polling = false
	stillSubscribed := b.findSubscriber(name, id) == current
	b.mu.Unlock()
	if !stillSubscribed {
		return response(http.StatusNotFound)
	}
	if !received {
		return response(http.StatusNoContent)
	}
	return jsonResponse(http.StatusOK, map[string]any{
		"sequence": message.sequence,
		"data":     message.data,
	})
}

// findSubscriber is called only while b.mu is locked.
func (b *Broker) findSubscriber(name string, id uint) *subscriber {
	if currentTopic := b.topics[name]; currentTopic != nil {
		return currentTopic.subscribers[id]
	}
	return nil
}

// subscription checks syntax. A valid number can still name an unknown subscriber.
func subscription(req *http.Request) (string, uint, bool) {
	name := req.URL.Query().Get("topic")
	id, err := strconv.ParseUint(req.URL.Query().Get("id"), 10, strconv.IntSize)
	if name == "" || err != nil || id == 0 {
		return "", 0, false
	}
	return name, uint(id), true
}

func response(status int) *http.Response {
	return httpserver.NewResponse(status, nil, "")
}

// any lets this small helper encode either an ID response or an event response.
// JSON encodes []byte as base64, so event data can contain arbitrary binary bytes.
func jsonResponse(status int, value any) *http.Response {
	data, err := json.Marshal(value)
	if err != nil {
		return response(http.StatusBadRequest)
	}
	return httpserver.NewResponse(status, data, "application/json")
}
