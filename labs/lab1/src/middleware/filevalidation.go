package middleware

import (
	"net/http"
	"strings"

	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/filepolicy"
)

// ValidateFilePath checks the request path before it is joined to a file root.
func ValidateFilePath(next httpserver.Handler) httpserver.Handler {
	return func(req *http.Request) *http.Response {
		// Check relative depth before path normalization can hide parent segments.
		path := strings.TrimLeft(req.URL.Path, "/")
		if !filepolicy.IsValidFilePath(path) {
			return httpserver.NewResponse(http.StatusBadRequest, nil, "")
		}
		return next(req)
	}
}

// ValidateFileExtension rejects extensions outside the assignment's allowlist.
func ValidateFileExtension(next httpserver.Handler) httpserver.Handler {
	return func(req *http.Request) *http.Response {
		if filepolicy.ContentTypeForPath(req.URL.Path) == "" {
			return httpserver.NewResponse(http.StatusBadRequest, nil, "")
		}
		return next(req)
	}
}
