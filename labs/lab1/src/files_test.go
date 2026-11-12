package main

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Zakoroo/distributed-systems/httpserver"
)

func TestFileServerResponsesAndMiddleware(t *testing.T) {
	root := t.TempDir()
	s, err := newServer(root)
	if err != nil {
		t.Fatal(err)
	}
	types := map[string]string{
		"txt": "text/plain", "html": "text/html", "css": "text/css",
		"jpg": "image/jpeg", "jpeg": "image/jpeg", "gif": "image/gif",
	}
	for extension := range types {
		if err := os.WriteFile(filepath.Join(root, "sample."+extension), []byte("hello"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(root, "directory.txt"), 0o755); err != nil {
		t.Fatal(err)
	}
	check := func(method, path, body string, status int, contentType, content string) {
		t.Helper()
		response := s.Handle(httptest.NewRequest(method, path, strings.NewReader(body)))
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != status || response.Header.Get("Content-Type") != contentType || string(data) != content {
			t.Fatalf("%s %s: got status=%d type=%q body=%q; want %d/%q/%q",
				method, path, response.StatusCode, response.Header.Get("Content-Type"), data, status, contentType, content)
		}
	}
	for extension, contentType := range types {
		check("GET", "/sample."+extension, "", 200, contentType, "hello")
	}
	check("GET", "/missing.txt", "", 404, "", "")
	check("GET", "/directory.txt", "", 400, "", "")
	check("GET", "/missing.bin", "", 400, "", "")
	check("GET", "/../outside.txt", "", 400, "", "")
	check("GET", "/sub/../../outside.txt", "", 400, "", "")
	check("GET", "/-invalid.txt", "", 400, "", "")
	check("PUT", "/invalid.bin", "", 501, "", "")
	check("POST", "/missing/child.txt", "body", 400, "", "")
	check("POST", "/sample.txt", " world", 200, "", "")
	check("GET", "/sample.txt", "", 200, "text/plain", "hello world")
	check("POST", "/created.txt", "new", 200, "", "")
	check("GET", "/created.txt", "", 200, "text/plain", "new")

	// File-only validation must not reject an unrelated endpoint without a file extension.
	if err := s.RegisterEndpoint("/health", func(*http.Request) *http.Response {
		return httpserver.NewResponse(200, []byte("ok"), "text/plain")
	}); err != nil {
		t.Fatal(err)
	}
	check("GET", "/health", "", 200, "text/plain", "ok")

	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(
		"POST /sample.txt HTTP/1.1\r\nHost: localhost\r\nContent-Length: 10\r\n\r\nshort")))
	if err != nil {
		t.Fatal(err)
	}
	defer req.Body.Close()
	response := s.Handle(req)
	response.Body.Close()
	if response.StatusCode != 400 {
		t.Fatalf("truncated POST = %d, want 400", response.StatusCode)
	}
	check("GET", "/sample.txt", "", 200, "text/plain", "hello world")
}

func TestConcurrentPostsExposeCompleteAppends(t *testing.T) {
	root := t.TempDir()
	s, err := newServer(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "shared.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	const chunk = "complete-message\n"
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 25 {
				response := s.Handle(httptest.NewRequest("POST", "/shared.txt", strings.NewReader(chunk)))
				response.Body.Close()
				if response.StatusCode != 200 {
					t.Errorf("POST returned %d", response.StatusCode)
				}
			}
		})
		wg.Go(func() {
			for range 25 {
				response := s.Handle(httptest.NewRequest("GET", "/shared.txt", nil))
				data, err := io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil {
					t.Error(err)
				}
				if response.StatusCode != 200 || strings.ReplaceAll(string(data), chunk, "") != "" {
					t.Errorf("GET observed incomplete messages: status=%d body=%q", response.StatusCode, data)
				}
			}
		})
	}
	wg.Wait()
	data, err := os.ReadFile(filepath.Join(root, "shared.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if want := strings.Repeat(chunk, 100); string(data) != want {
		t.Fatalf("final file contains %d bytes, want %d complete-message bytes", len(data), len(want))
	}
}
