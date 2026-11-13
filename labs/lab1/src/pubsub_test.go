package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestPubSubRoutesUseTheApplicationServer(t *testing.T) {
	s, err := newServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body string, status int) *http.Response {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		t.Cleanup(func() { req.Body.Close() })
		response := s.Handle(req)
		t.Cleanup(func() { response.Body.Close() })
		if response.StatusCode != status {
			t.Fatalf("%s %s: status %d, want %d", method, path, response.StatusCode, status)
		}
		return response
	}
	response := call("POST", "/subscribe?topic=integration", "", 201)
	var subscription struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&subscription); err != nil {
		t.Fatal(err)
	}
	path := "?topic=integration&id=" + strconv.FormatUint(uint64(subscription.ID), 10)
	call("POST", "/publish?topic=integration", "hello", 200)
	response = call("GET", "/events"+path, "", 200)
	body, err := io.ReadAll(response.Body)
	if err != nil || !strings.Contains(string(body), `"data":"aGVsbG8="`) {
		t.Fatalf("events %s: %v", body, err)
	}
	call("GET", "/publish?topic=integration", "", 501)
	call("POST", "/unsubscribe"+path, "", 200)
	call("GET", "/events"+path, "", 404)
}
