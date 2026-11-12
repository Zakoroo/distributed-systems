// Package fileapi implements Lab 1 file operations as HTTP handlers.
package fileapi

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Zakoroo/distributed-systems/concurrent/files"
	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/filepolicy"
)

// Handler serves files below its configured root. Apply file-validation middleware
// before calling Handle with untrusted requests.
type Handler struct {
	root    string
	manager *files.Manager
}

func New(root string) *Handler {
	return &Handler{root: root, manager: files.NewManager()}
}

// Handle expects file policy to have been applied by its middleware chain.
func (h *Handler) Handle(req *http.Request) *http.Response {
	switch req.Method {
	case http.MethodGet:
		return h.get(req.URL.Path)
	case http.MethodPost:
		return h.post(req)
	default:
		return httpserver.NewResponse(http.StatusNotImplemented, nil, "")
	}
}

func (h *Handler) get(path string) *http.Response {
	filename := filepath.Join(h.root, path)
	var data []byte
	err := h.manager.WithFile(filename, "r", 0o644, func(f *os.File) error {
		var err error
		data, err = io.ReadAll(f)
		return err
	})
	if errors.Is(err, os.ErrNotExist) {
		return httpserver.NewResponse(http.StatusNotFound, nil, "")
	}
	if err != nil {
		log.Print(err)
		return httpserver.NewResponse(http.StatusBadRequest, nil, "")
	}
	return httpserver.NewResponse(http.StatusOK, data, filepolicy.ContentTypeForPath(path))
}

func (h *Handler) post(req *http.Request) *http.Response {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		if err != nil {
			return httpserver.NewResponse(http.StatusBadRequest, nil, "")
		}
	}
	filename := filepath.Join(h.root, req.URL.Path)
	err := h.manager.WithFile(filename, "a", 0o644, func(f *os.File) error {
		_, err := f.Write(body)
		return err
	})
	if err != nil {
		log.Print(err)
		return httpserver.NewResponse(http.StatusBadRequest, nil, "")
	}
	return httpserver.NewResponse(http.StatusOK, nil, "")
}
