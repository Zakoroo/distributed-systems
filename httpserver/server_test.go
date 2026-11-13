package httpserver

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestConnectionResponseAndClosure(t *testing.T) {
	for _, test := range []struct {
		name, request string
		status        int
	}{
		{"valid", "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n", 200},
		{"malformed", "GET / HTTP/1.1\r\nnot a header\r\n\r\n", 400},
		// The handler does not read this body. Cleanup must close the connection
		// before trying to drain it, rather than waiting for the missing bytes.
		{"unread body", "GET / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 100\r\n\r\n", 200},
	} {
		t.Run(test.name, func(t *testing.T) {
			s, _ := New(10)
			s.logger = log.New(io.Discard, "", 0)
			body := &trackedBody{Reader: strings.NewReader("response")}
			if err := s.SetFallback(func(*http.Request) *http.Response {
				response := NewResponse(200, []byte("response"), "text/plain")
				response.Body = body
				return response
			}); err != nil {
				t.Fatal(err)
			}
			client, conn := net.Pipe()
			defer client.Close()
			if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- s.handleConnection(conn) }()
			if _, err := io.WriteString(client, test.request); err != nil {
				t.Fatal(err)
			}
			reader := bufio.NewReader(client)
			response, err := http.ReadResponse(reader, nil)
			if err != nil {
				t.Fatal(err)
			}
			data, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != test.status || !response.Close {
				t.Fatalf("status=%d close=%v, want %d/true", response.StatusCode, response.Close, test.status)
			}
			if test.status == 200 && string(data) != "response" {
				t.Fatalf("body = %q", data)
			}
			if _, err := reader.ReadByte(); err != io.EOF {
				t.Fatalf("connection not closed: %v", err)
			}
			if err := <-done; err != nil {
				t.Fatal(err)
			}
			if test.status == 200 && !body.closed {
				t.Error("response body was not closed")
			}
		})
	}
}

type trackedBody struct {
	*strings.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

func TestServeLimitsHandlersAndReleasesSlots(t *testing.T) {
	for _, limit := range []int{2, 10} {
		t.Run(fmt.Sprintf("limit=%d", limit), func(t *testing.T) {
			testServeLimit(t, limit)
		})
	}
}

func TestNewRejectsZeroLimit(t *testing.T) {
	s, err := New(0)
	if err == nil || s != nil {
		t.Fatalf("New(0) = %v, %v; want nil server and an error", s, err)
	}
}

func testServeLimit(t *testing.T, limit int) {
	t.Helper()
	s, err := New(uint(limit))
	if err != nil {
		t.Fatal(err)
	}
	s.logger = log.New(io.Discard, "", 0)
	listener := &pipeListener{
		connections: make(chan net.Conn, limit+1),
		accepted:    make(chan struct{}, limit+1),
		closed:      make(chan struct{}),
	}
	defer listener.Close()
	release := make(chan struct{})
	defer close(release)
	entered := make(chan struct{}, limit+1)
	var active, peak atomic.Int32
	if err := s.SetFallback(func(*http.Request) *http.Response {
		count := active.Add(1)
		for previous := peak.Load(); count > previous; previous = peak.Load() {
			if peak.CompareAndSwap(previous, count) {
				break
			}
		}
		entered <- struct{}{}
		<-release
		active.Add(-1)
		return NewResponse(200, nil, "")
	}); err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- s.Serve(listener) }()
	results := make(chan error, limit+1)
	for range limit + 1 {
		client, conn := net.Pipe()
		defer client.Close()
		listener.connections <- conn
		go func() {
			client.SetDeadline(time.Now().Add(5 * time.Second))
			if _, err := io.WriteString(client, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
				results <- err
				return
			}
			response, err := http.ReadResponse(bufio.NewReader(client), nil)
			if err == nil {
				_, err = io.Copy(io.Discard, response.Body)
				response.Body.Close()
			}
			results <- err
		}()
	}
	wait := func(events <-chan struct{}) {
		t.Helper()
		select {
		case <-events:
		case <-time.After(5 * time.Second):
			t.Fatal("server did not make progress")
		}
	}
	for range limit {
		wait(entered)
	}
	for range limit + 1 {
		wait(listener.accepted)
	}
	if got := active.Load(); got != int32(limit) {
		t.Fatalf("active handlers = %d, want %d", got, limit)
	}
	select {
	case <-entered:
		t.Fatal("extra handler started before a slot was released")
	default:
	}
	release <- struct{}{}
	wait(entered) // The queued connection can now get the released slot.
	for range limit {
		release <- struct{}{}
	}
	for range limit + 1 {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	listener.Close()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	if got := peak.Load(); got != int32(limit) {
		t.Fatalf("peak handlers = %d, want %d", got, limit)
	}
}

type pipeListener struct {
	connections chan net.Conn
	accepted    chan struct{}
	closed      chan struct{}
	once        sync.Once
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connections:
		l.accepted <- struct{}{}
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return &net.TCPAddr{} }

func TestResponseBodyClosesAfterTheFinalWrite(t *testing.T) {
	s, _ := New(1)
	s.logger = log.New(io.Discard, "", 0)
	closed := make(chan struct{})
	s.SetFallback(func(*http.Request) *http.Response {
		response := NewResponse(200, []byte("message"), "text/plain")
		response.Body = &closingBody{reader: strings.NewReader("message"), closed: closed}
		return response
	})
	client, conn := net.Pipe()
	defer client.Close()
	client.SetDeadline(time.Now().Add(3 * time.Second))
	writing := make(chan struct{})
	done := make(chan error, 1)
	go func() { done <- s.handleConnection(&writeSignalConn{Conn: conn, writing: writing}) }()
	if _, err := io.WriteString(client, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-writing:
	case <-time.After(3 * time.Second):
		t.Fatal("response write did not start")
	}
	select {
	case <-closed:
		t.Fatal("body closed before the blocked response flush completed")
	default:
	}
	response, err := http.ReadResponse(bufio.NewReader(client), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(io.Discard, response.Body); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	select {
	case <-closed:
	default:
		t.Fatal("body not closed after response completion")
	}
}

type closingBody struct {
	reader *strings.Reader
	closed chan struct{}
}

func (b *closingBody) Read(p []byte) (int, error) { return b.reader.Read(p) }
func (b *closingBody) Close() error               { close(b.closed); return nil }

type writeSignalConn struct {
	net.Conn
	writing chan struct{}
	once    sync.Once
}

func (c *writeSignalConn) Write(p []byte) (int, error) {
	c.once.Do(func() { close(c.writing) })
	return c.Conn.Write(p)
}
