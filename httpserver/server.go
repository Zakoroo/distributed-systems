// Package httpserver implements HTTP transport and routing over raw TCP connections.
package httpserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/Zakoroo/distributed-systems/concurrent/semaphores"
	"github.com/Zakoroo/distributed-systems/concurrent/waiting"
)

// Server dispatches requests through ordered endpoints and middleware.
// Construct it with New. A Server must not be copied after first use.
type Server struct {
	mu                    sync.RWMutex
	endpoints             []endpoint
	middleware            []scopedMiddleware
	fallback              Handler
	logger                *log.Logger
	maxConcurrentRequests uint
}

func New(maxConcurrentRequests uint) (*Server, error) {
	if maxConcurrentRequests == 0 {
		return nil, fmt.Errorf("maxConcurrentRequests must be greater than 0")
	}
	return &Server{
		fallback: func(*http.Request) *http.Response {
			return NewResponse(http.StatusNotImplemented, nil, "")
		},
		logger:                log.Default(),
		maxConcurrentRequests: maxConcurrentRequests,
	}, nil
}

// Start listens on address and serves requests until the listener fails.
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer listener.Close()
	return s.Serve(listener)
}

// Serve bounds active request work by the configured admission limit. Blocking
// waiting.Wait callbacks temporarily yield their permit. The caller owns
// listener; closing it stops acceptance while admitted requests finish normally.
func (s *Server) Serve(listener net.Listener) error {
	sem, err := semaphores.NewSemaphore(s.maxConcurrentRequests)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		if err != nil {
			return err
		}
		sem.Down()
		s.logger.Printf("Accepted new client connection: %v", conn.RemoteAddr())
		go func() {
			defer sem.Up()
			ctx := waiting.WithPolicy(context.Background(), func(wait func() error) error {
				sem.Up()
				defer sem.Down()
				return wait()
			})
			if err := s.serveConnection(ctx, conn); err != nil {
				s.logger.Print(err)
			}
		}()
	}
}

func (s *Server) handleConnection(conn net.Conn) error {
	return s.serveConnection(context.Background(), conn)
}

func (s *Server) serveConnection(ctx context.Context, conn net.Conn) error {
	var req *http.Request
	defer func() {
		// Close the socket first: Body.Close may try to drain unread input.
		conn.Close()
		if req != nil && req.Body != nil {
			req.Body.Close()
		}
	}()

	// Read/write deadlines and connection reuse are separate increments.
	var err error
	req, err = http.ReadRequest(bufio.NewReader(conn))
	if errors.Is(err, io.EOF) {
		return nil
	}
	var response *http.Response
	if err != nil {
		s.logger.Print(err)
		response = NewResponse(http.StatusBadRequest, nil, "")
	} else {
		req = req.WithContext(ctx)
		s.logger.Printf("method: %s, path: %s", req.Method, req.URL.Path)
		response = s.Handle(req)
	}
	if response.Body != nil {
		body := response.Body
		defer body.Close()
		// Response.Write may close Body before its final buffered flush. Keep
		// ownership until the entire write completes, including that flush.
		response.Body = io.NopCloser(body)
	}
	response.Request = req
	response.Close = true
	return response.Write(conn)
}
