package httpserver

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// NewResponse constructs an HTTP/1.1 response with an optional content type.
func NewResponse(statusCode int, body []byte, contentType string) *http.Response {
	headers := make(http.Header)
	if contentType != "" {
		headers.Set("Content-Type", contentType)
	}
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        headers,
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
	}
}
