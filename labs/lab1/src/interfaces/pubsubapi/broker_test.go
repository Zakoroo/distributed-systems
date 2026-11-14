package pubsubapi_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/interfaces/pubsubapi"
)

func newBroker(t *testing.T, capacity int, timeout time.Duration) *pubsubapi.Broker {
	t.Helper()
	broker, err := pubsubapi.New(capacity, timeout)
	if err != nil {
		t.Fatal(err)
	}
	return broker
}

func call(t *testing.T, handler httpserver.Handler, method, path, body string, status int) []byte {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	defer req.Body.Close()
	response := handler(req)
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != status {
		t.Fatalf("%s %s: status=%d err=%v; want %d", method, path, response.StatusCode, err, status)
	}
	if len(data) > 0 && response.Header.Get("Content-Type") != "application/json" {
		t.Fatal("JSON response is missing its content type")
	}
	return data
}

func subscribe(t *testing.T, broker *pubsubapi.Broker, topic string) uint {
	t.Helper()
	data := call(t, broker.HandleSubscribe, "POST", "/subscribe?topic="+url.QueryEscape(topic), "", 201)
	var result struct{ ID uint }
	if err := json.Unmarshal(data, &result); err != nil || result.ID == 0 {
		t.Fatal("invalid subscriber ID", err)
	}
	return result.ID
}

func subscriberQuery(topic string, id uint) string {
	return fmt.Sprintf("?topic=%s&id=%d", url.QueryEscape(topic), id)
}

type message struct {
	Sequence uint64 `json:"sequence"`
	Data     []byte `json:"data"`
}

func nextEvent(t *testing.T, broker *pubsubapi.Broker, topic string, id uint) message {
	t.Helper()
	data := call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery(topic, id), "", 200)
	var result message
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestQueueBoundsFanoutAndSubscriptionWindow(t *testing.T) {
	broker := newBroker(t, 1, time.Millisecond)
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "before subscription", 200)
	alice := subscribe(t, broker, "news")
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "first", 200)
	bob := subscribe(t, broker, "news")
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "second", 200)
	// Alice's full queue keeps first. Bob joined later and receives second only.
	if got := nextEvent(t, broker, "news", alice); got.Sequence != 1 || string(got.Data) != "first" {
		t.Fatalf("Alice: %+v", got)
	}
	if got := nextEvent(t, broker, "news", bob); got.Sequence != 2 || string(got.Data) != "second" {
		t.Fatalf("Bob: %+v", got)
	}
	call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("news", alice), "", 204)
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "after timeout", 200)
	if got := nextEvent(t, broker, "news", alice); got.Sequence != 3 || string(got.Data) != "after timeout" {
		t.Fatalf("poll did not recover after timeout: %+v", got)
	}
	call(t, broker.HandleUnsubscribe, "POST", "/unsubscribe"+subscriberQuery("news", alice), "", 200)
	call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("news", alice), "", 404)
	call(t, broker.HandleUnsubscribe, "POST", "/unsubscribe"+subscriberQuery("news", alice), "", 404)
	other := subscribe(t, broker, "other")
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "third", 200)
	call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("other", other), "", 204)
}

func TestValidationAndMethods(t *testing.T) {
	for _, config := range []struct {
		capacity int
		timeout  time.Duration
	}{{0, time.Second}, {-1, time.Second}, {1, 0}, {1, -time.Second}} {
		if _, err := pubsubapi.New(config.capacity, config.timeout); err == nil {
			t.Fatal("accepted invalid broker configuration")
		}
	}
	broker := newBroker(t, 2, time.Millisecond)
	for _, endpoint := range []struct {
		path    string
		handler httpserver.Handler
	}{{"/subscribe", broker.HandleSubscribe}, {"/unsubscribe", broker.HandleUnsubscribe}, {"/publish", broker.HandlePublish}, {"/events", broker.HandleEvents}} {
		call(t, endpoint.handler, "PUT", endpoint.path, "", 501)
		method := "POST"
		if endpoint.path == "/events" {
			method = "GET"
		}
		call(t, endpoint.handler, method, endpoint.path, "", 400)
	}
	call(t, broker.HandleSubscribe, "GET", "/subscribe?topic=news", "", 501)
	call(t, broker.HandlePublish, "GET", "/publish?topic=news", "", 501)
	call(t, broker.HandleUnsubscribe, "GET", "/unsubscribe?topic=news&id=1", "", 501)
	call(t, broker.HandleEvents, "POST", "/events?topic=news&id=1", "", 501)
	for _, id := range []string{"", "0", "-1", "+1", "1_0", "abc", "18446744073709551616"} {
		query := "?topic=news&id=" + url.QueryEscape(id)
		call(t, broker.HandleEvents, "GET", "/events"+query, "", 400)
		call(t, broker.HandleUnsubscribe, "POST", "/unsubscribe"+query, "", 400)
	}
	id := subscribe(t, broker, "news")
	call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("other", id), "", 404)
	call(t, broker.HandleEvents, "GET", "/events?topic=news&id=999", "", 404)

	// A failed body read must not publish a partial payload.
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(
		"POST /publish?topic=news HTTP/1.1\r\nHost: localhost\r\nContent-Length: 10\r\n\r\nshort")))
	if err != nil {
		t.Fatal(err)
	}
	defer req.Body.Close()
	response := broker.HandlePublish(req)
	response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatal("accepted a truncated body")
	}
	call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("news", id), "", 204)
}

func TestBinaryPayloadOwnershipAndEmptyMessage(t *testing.T) {
	broker := newBroker(t, 2, time.Millisecond)
	alice := subscribe(t, broker, "news")
	bob := subscribe(t, broker, "news")
	input := []byte{0, 255, 'x'}
	req := httptest.NewRequest("POST", "/publish?topic=news", bytes.NewReader(input))
	defer req.Body.Close()
	response := broker.HandlePublish(req)
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatal("binary publication failed")
	}
	input[0] = 42
	for _, id := range []uint{alice, bob} {
		if got := nextEvent(t, broker, "news", id); !bytes.Equal(got.Data, []byte{0, 255, 'x'}) {
			t.Fatalf("payload changed: %v", got.Data)
		}
	}
	call(t, broker.HandlePublish, "POST", "/publish?topic=news", "", 200)
	data := call(t, broker.HandleEvents, "GET", "/events"+subscriberQuery("news", alice), "", 200)
	if !strings.Contains(string(data), `"data":""`) {
		t.Fatalf("empty payload must be a base64 string: %s", data)
	}
}

func TestConcurrentPublishersPreserveOrderForBothSubscribers(t *testing.T) {
	broker := newBroker(t, 128, time.Second)
	alice := subscribe(t, broker, "news")
	bob := subscribe(t, broker, "news")
	var workers sync.WaitGroup
	for publisher := range 4 {
		workers.Go(func() {
			for number := range 20 {
				call(t, broker.HandlePublish, "POST", "/publish?topic=news", fmt.Sprintf("%d/%d", publisher, number), 200)
			}
		})
	}
	workers.Wait()
	if t.Failed() {
		return
	}
	for number := uint64(1); number <= 80; number++ {
		first, second := nextEvent(t, broker, "news", alice), nextEvent(t, broker, "news", bob)
		if first.Sequence != number || second.Sequence != number || !bytes.Equal(first.Data, second.Data) {
			t.Fatalf("publication %d: Alice=%+v Bob=%+v", number, first, second)
		}
	}
}

func TestWaitingPollWakeupsAndOverlappingWait(t *testing.T) {
	for _, wakeup := range []string{"publish", "unsubscribe"} {
		t.Run(wakeup, func(t *testing.T) {
			broker := newBroker(t, 1, 5*time.Second)
			id := subscribe(t, broker, "news")
			defer func() {
				result := broker.HandleUnsubscribe(httptest.NewRequest("POST", "/unsubscribe"+subscriberQuery("news", id), nil))
				result.Body.Close()
			}()
			type result struct {
				status int
				data   []byte
			}
			results := make(chan result, 2)
			for range 2 {
				go func() {
					response := broker.HandleEvents(httptest.NewRequest("GET", "/events"+subscriberQuery("news", id), nil))
					data, err := io.ReadAll(response.Body)
					response.Body.Close()
					if err != nil {
						t.Error(err)
					}
					results <- result{response.StatusCode, data}
				}()
			}
			// The first result proves that the other handler has claimed the queue.
			select {
			case got := <-results:
				if got.status != 400 {
					t.Fatalf("overlapping wait returned %d", got.status)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("overlapping poll was not rejected")
			}
			want := 200
			if wakeup == "publish" {
				call(t, broker.HandlePublish, "POST", "/publish?topic=news", "hello", 200)
			} else {
				call(t, broker.HandleUnsubscribe, "POST", "/unsubscribe"+subscriberQuery("news", id), "", 200)
				want = 404
			}
			select {
			case got := <-results:
				if got.status != want {
					t.Fatalf("%s woke poll with %d, want %d", wakeup, got.status, want)
				}
				if wakeup == "publish" {
					var event message
					if err := json.Unmarshal(got.data, &event); err != nil || event.Sequence != 1 || string(event.Data) != "hello" {
						t.Fatalf("invalid wakeup event: %s (%v)", got.data, err)
					}
				}
			case <-time.After(3 * time.Second):
				t.Fatalf("%s did not wake the poll", wakeup)
			}
		})
	}
}

func TestConcurrentSubscribePublishAndUnsubscribe(t *testing.T) {
	broker := newBroker(t, 2, time.Millisecond)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			for range 25 {
				id := subscribe(t, broker, "news")
				call(t, broker.HandlePublish, "POST", "/publish?topic=news", "message", 200)
				call(t, broker.HandleUnsubscribe, "POST", "/unsubscribe"+subscriberQuery("news", id), "", 200)
			}
		})
	}
	workers.Wait() // In particular, publication must never send to a closed queue.
}
