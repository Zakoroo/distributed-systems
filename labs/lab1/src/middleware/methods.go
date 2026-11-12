// Package middleware applies Lab 1 request-validation policy to HTTP handlers.
package middleware

import (
	"net/http"

	"github.com/Zakoroo/distributed-systems/httpserver"
)

// ValidateMethod permits only the methods implemented by the basic assignment.
func ValidateMethod(next httpserver.Handler) httpserver.Handler {
	return func(req *http.Request) *http.Response {
		if req.Method != http.MethodGet && req.Method != http.MethodPost {
			return httpserver.NewResponse(http.StatusNotImplemented, nil, "")
		}
		return next(req)
	}
}
