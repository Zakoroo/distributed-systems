package httpserver

import (
	"bufio"
	"github.com/Zakoroo/distributed-systems/concurrent/waiting"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestWaitingRequestsLeavePublisherCapacity(t *testing.T) {
	s, _ := New(1)
	s.logger = log.New(io.Discard, "", 0)
	listener := &pipeListener{connections: make(chan net.Conn, 4), accepted: make(chan struct{}, 4), closed: make(chan struct{})}
	defer listener.Close()
	gate := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(gate) }) }
	defer unblock()
	entered := make(chan struct{}, 3)
	s.SetFallback(func(req *http.Request) *http.Response {
		if req.URL.Path == "/wait" {
			waiting.Wait(req.Context(), func() error { entered <- struct{}{}; <-gate; return nil })
		} else {
			unblock()
		}
		return NewResponse(200, nil, "")
	})
	served := make(chan error, 1)
	go func() { served <- s.Serve(listener) }()
	results := make(chan error, 4)
	request := func(path string) {
		client, conn := net.Pipe()
		t.Cleanup(func() { client.Close() })
		listener.connections <- conn
		go func() {
			client.SetDeadline(time.Now().Add(3 * time.Second))
			_, err := io.WriteString(client, "GET "+path+" HTTP/1.1\r\nHost: localhost\r\n\r\n")
			if err == nil {
				var response *http.Response
				response, err = http.ReadResponse(bufio.NewReader(client), nil)
				if err == nil {
					_, err = io.Copy(io.Discard, response.Body)
					response.Body.Close()
				}
			}
			results <- err
		}()
	}
	for range 3 {
		request("/wait")
		select {
		case <-entered:
		case <-time.After(3 * time.Second):
			t.Fatal("waiting request retained its admission slot")
		}
	}
	request("/publish")
	for range 4 {
		select {
		case err := <-results:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("publisher or waiting response stalled")
		}
	}
	listener.Close()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}
