package httpserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
)

func statusHandler(status int) Handler {
	return func(*http.Request) *http.Response { return NewResponse(status, nil, "") }
}

func dispatchStatus(s *Server, path string) int {
	response := s.Handle(httptest.NewRequest(http.MethodGet, path, nil))
	defer response.Body.Close()
	return response.StatusCode
}

func TestEndpointOrderRemovalAndFallback(t *testing.T) {
	s, _ := New(10)
	if got := dispatchStatus(s, "/unknown"); got != 501 {
		t.Fatalf("default fallback = %d, want 501", got)
	}
	for _, route := range []struct {
		pattern string
		status  int
	}{{"/items/.*", 201}, {"/items/special", 202}} {
		if err := s.RegisterEndpoint(route.pattern, statusHandler(route.status)); err != nil {
			t.Fatal(err)
		}
	}
	if got := dispatchStatus(s, "/items/special?ignored=yes"); got != 201 {
		t.Fatalf("first registered match = %d, want 201", got)
	}
	if err := s.UnregisterEndpoint("/items/.*"); err != nil {
		t.Fatal(err)
	}
	if got := dispatchStatus(s, "/items/special"); got != 202 {
		t.Fatalf("after removing first match = %d, want 202", got)
	}
	if err := s.SetFallback(statusHandler(404)); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterEndpoint("/later", statusHandler(200)); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]int{
		"/later": 200, "/unknown": 404, "/items/special/extra": 404,
		"/prefix/items/special": 404, "/unknown?path=/items/special": 404,
	} {
		if got := dispatchStatus(s, path); got != want {
			t.Errorf("%s: got %d, want %d", path, got, want)
		}
	}
}

func TestRegistrationErrors(t *testing.T) {
	s, _ := New(10)
	for _, pattern := range []string{"", "["} {
		if err := s.RegisterEndpoint(pattern, statusHandler(200)); err == nil {
			t.Errorf("accepted invalid endpoint pattern %q", pattern)
		}
		if err := s.RegisterMiddleware(pattern, func(next Handler) Handler { return next }); err == nil {
			t.Errorf("accepted invalid middleware pattern %q", pattern)
		}
	}
	if err := s.RegisterEndpoint("/nil", nil); err == nil {
		t.Error("accepted nil handler")
	}
	if err := s.RegisterMiddleware(".*", nil); err == nil {
		t.Error("accepted nil middleware")
	}
	if err := s.SetFallback(nil); err == nil {
		t.Error("accepted nil fallback")
	}
	if err := s.RegisterEndpoint("/duplicate", statusHandler(200)); err != nil {
		t.Fatal(err)
	}
	if err := s.RegisterEndpoint("/duplicate", statusHandler(201)); err == nil {
		t.Error("accepted duplicate endpoint")
	}
	if err := s.UnregisterEndpoint("/absent"); err == nil {
		t.Error("removed unregistered endpoint")
	}
}

func TestScopedMiddlewareOrderAndShortCircuit(t *testing.T) {
	s, _ := New(10)
	var events []string
	trace := func(name string) Middleware {
		return func(next Handler) Handler {
			return func(req *http.Request) *http.Response {
				events = append(events, name+" before")
				response := next(req)
				events = append(events, name+" after")
				return response
			}
		}
	}
	for _, registration := range []struct {
		pattern string
		name    string
	}{{".*", "global"}, {"/files/.*", "file A"}, {"/files/.*", "file B"}, {"/admin/.*", "admin"}} {
		if err := s.RegisterMiddleware(registration.pattern, trace(registration.name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetFallback(func(*http.Request) *http.Response {
		events = append(events, "handler")
		return NewResponse(200, nil, "")
	}); err != nil {
		t.Fatal(err)
	}
	dispatchStatus(s, "/files/a.txt")
	want := []string{"global before", "file A before", "file B before", "handler", "file B after", "file A after", "global after"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("middleware order = %v, want %v", events, want)
	}
	events = nil
	dispatchStatus(s, "/health")
	if want := []string{"global before", "handler", "global after"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("nonmatching middleware ran: %v", events)
	}
	if err := s.RegisterMiddleware("/blocked", func(Handler) Handler { return statusHandler(400) }); err != nil {
		t.Fatal(err)
	}
	events = nil
	if got := dispatchStatus(s, "/blocked"); got != 400 {
		t.Fatalf("short-circuit response = %d, want 400", got)
	}
	if want := []string{"global before", "global after"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("short-circuit called handler: %v", events)
	}
}

func TestHandlerCanUnregisterItself(t *testing.T) {
	s, _ := New(10)
	if err := s.RegisterEndpoint("/once", func(*http.Request) *http.Response {
		if err := s.UnregisterEndpoint("/once"); err != nil {
			t.Error(err)
		}
		return NewResponse(200, nil, "")
	}); err != nil {
		t.Fatal(err)
	}
	if got := dispatchStatus(s, "/once"); got != 200 {
		t.Fatalf("selected handler = %d, want 200", got)
	}
	if got := dispatchStatus(s, "/once"); got != 501 {
		t.Fatalf("removed handler = %d, want fallback 501", got)
	}
}

func TestConcurrentRegistryUpdatesAndDispatch(t *testing.T) {
	s, _ := New(10)
	if err := s.RegisterEndpoint("/stable", statusHandler(200)); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for worker := range 4 {
		wg.Go(func() {
			pattern := fmt.Sprintf("/temporary/%d", worker)
			if err := s.RegisterMiddleware(".*", func(next Handler) Handler { return next }); err != nil {
				t.Error(err)
			}
			for range 100 {
				if err := s.RegisterEndpoint(pattern, statusHandler(201)); err != nil {
					t.Error(err)
				}
				if err := s.SetFallback(statusHandler(404)); err != nil {
					t.Error(err)
				}
				if got := dispatchStatus(s, "/stable"); got != 200 {
					t.Errorf("stable endpoint returned %d", got)
				}
				if err := s.UnregisterEndpoint(pattern); err != nil {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
}
