package httpserver

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
)

// Handler returns a response whose body is owned and closed by the server.
type Handler func(*http.Request) *http.Response

// Middleware wraps a handler, optionally returning a response without calling it.
type Middleware func(Handler) Handler

type endpoint struct {
	pattern string
	matcher *regexp.Regexp
	handler Handler
}

type scopedMiddleware struct {
	matcher    *regexp.Regexp
	middleware Middleware
}

// RegisterEndpoint appends a unique regular-expression path pattern. Patterns
// match the whole URL path, excluding its query. The first matching endpoint wins.
func (s *Server) RegisterEndpoint(pattern string, handler Handler) error {
	if handler == nil {
		return fmt.Errorf("endpoint handler must not be nil")
	}
	matcher, err := compilePattern(pattern)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, endpoint := range s.endpoints {
		if endpoint.pattern == pattern {
			return fmt.Errorf("endpoint pattern %q is already registered", pattern)
		}
	}
	s.endpoints = append(s.endpoints, endpoint{pattern, matcher, handler})
	return nil
}

// UnregisterEndpoint removes a pattern. Requests that already selected its
// handler may finish; dispatches after removal use the updated registry.
func (s *Server) UnregisterEndpoint(pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, endpoint := range s.endpoints {
		if endpoint.pattern == pattern {
			s.endpoints = slices.Delete(s.endpoints, i, i+1)
			return nil
		}
	}
	return fmt.Errorf("endpoint pattern %q is not registered", pattern)
}

// SetFallback replaces the handler used when no ordinary endpoint matches.
func (s *Server) SetFallback(handler Handler) error {
	if handler == nil {
		return fmt.Errorf("fallback handler must not be nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fallback = handler
	return nil
}

// RegisterMiddleware appends middleware for a whole-path regular expression.
// Matching middleware run in registration order, including for fallback requests.
// Multiple middleware may use the same pattern.
func (s *Server) RegisterMiddleware(pattern string, middleware Middleware) error {
	if middleware == nil {
		return fmt.Errorf("middleware must not be nil")
	}
	matcher, err := compilePattern(pattern)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.middleware = append(s.middleware, scopedMiddleware{matcher, middleware})
	return nil
}

// Handle dispatches a parsed request. It snapshots registry state and runs user
// code without holding registry locks. The caller owns the request/response bodies.
func (s *Server) Handle(req *http.Request) *http.Response {
	s.mu.RLock()
	handler := s.fallback
	for _, endpoint := range s.endpoints {
		if endpoint.matcher.MatchString(req.URL.Path) {
			handler = endpoint.handler
			break
		}
	}
	var middleware []Middleware
	for _, entry := range s.middleware {
		if entry.matcher.MatchString(req.URL.Path) {
			middleware = append(middleware, entry.middleware)
		}
	}
	s.mu.RUnlock()
	handler = Chain(handler, middleware...)
	return handler(req)
}

// Chain applies middleware in the supplied order: Chain(h, a, b) runs a(b(h)).
// Use it to attach policy only to a specific handler, such as a file fallback.
func Chain(handler Handler, middleware ...Middleware) Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, fmt.Errorf("path pattern must not be empty")
	}
	matcher, err := regexp.Compile("\\A(?:" + pattern + ")\\z")
	if err != nil {
		return nil, fmt.Errorf("invalid path pattern %q: %w", pattern, err)
	}
	return matcher, nil
}
