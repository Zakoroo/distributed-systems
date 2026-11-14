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

func TestPubSubRoutesShareTheFileServer(t *testing.T) {
	server, err := newServer(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	call := func(method, path, body string, status int) []byte {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		defer req.Body.Close()
		response := server.Handle(req)
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil || response.StatusCode != status {
			t.Fatalf("%s %s: status=%d err=%v", method, path, response.StatusCode, err)
		}
		return data
	}
	data := call("POST", "/subscribe?topic=weather/in-town", "", 201)
	var subscription struct{ ID uint }
	if err := json.Unmarshal(data, &subscription); err != nil || subscription.ID == 0 {
		t.Fatal("invalid subscriber ID", err)
	}
	query := "?topic=weather/in-town&id=" + strconv.FormatUint(uint64(subscription.ID), 10)
	call("POST", "/publish?topic=weather/in-town", "rain", 200)
	data = call("GET", "/events"+query, "", 200)
	var event struct {
		Sequence uint64
		Data     []byte
	}
	if err := json.Unmarshal(data, &event); err != nil || event.Sequence != 1 || string(event.Data) != "rain" {
		t.Fatalf("invalid event %s: %v", data, err)
	}
	call("POST", "/unsubscribe"+query, "", 200)
	call("GET", "/events"+query, "", 404)
	call("GET", "/publish?topic=weather/in-town", "", 501)
	call(http.MethodPut, "/events"+query, "", 501)
	// The normal file fallback still works alongside the four bonus endpoints.
	call("POST", "/notes.txt", "hello", 200)
	if got := string(call("GET", "/notes.txt", "", 200)); got != "hello" {
		t.Fatalf("file fallback returned %q", got)
	}
}
