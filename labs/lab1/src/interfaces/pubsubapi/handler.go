// Package pubsubapi translates HTTP requests to broker operations.
package pubsubapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/pubsub"
)

type Handler struct {
	service *pubsub.PubSubService
	timeout time.Duration
	mu      sync.Mutex
	polls   map[pollKey]*pollBody
}

func New(service *pubsub.PubSubService, timeout time.Duration) (*Handler, error) {
	if service == nil || timeout <= 0 {
		return nil, errors.New("service and positive polling timeout are required")
	}
	return &Handler{service: service, timeout: timeout, polls: make(map[pollKey]*pollBody)}, nil
}

func (h *Handler) HandleSubscribe(req *http.Request) *http.Response {
	if req.Method != http.MethodPost {
		return response(501)
	}
	name := req.URL.Query().Get("topic")
	if name == "" {
		return response(400)
	}
	id, err := h.service.Subscribe(name)
	if err != nil {
		return errorResponse(err)
	}
	return jsonResponse(201, struct {
		ID uint `json:"id"`
	}{id})
}

func (h *Handler) HandleUnsubscribe(req *http.Request) *http.Response {
	name, id, err := subscription(req, http.MethodPost)
	if err != nil {
		return errorResponse(err)
	}
	if err := h.service.Unsubscribe(name, id); err != nil {
		return errorResponse(err)
	}
	h.mu.Lock()
	delete(h.polls, pollKey{name, id})
	h.mu.Unlock()
	return response(200)
}

func (h *Handler) HandlePublish(req *http.Request) *http.Response {
	if req.Method != http.MethodPost {
		return response(501)
	}
	name := req.URL.Query().Get("topic")
	if name == "" {
		return response(400)
	}
	var data []byte
	if req.Body != nil {
		var err error
		data, err = io.ReadAll(req.Body)
		if err != nil {
			return response(400)
		}
	}
	if err := h.service.Publish(name, data); err != nil {
		return errorResponse(err)
	}
	return response(200)
}

func (h *Handler) HandleEvents(req *http.Request) *http.Response {
	name, id, err := subscription(req, http.MethodGet)
	if err != nil {
		return errorResponse(err)
	}
	key := pollKey{name, id}
	lease := &pollBody{}
	h.mu.Lock()
	if h.polls[key] != nil {
		h.mu.Unlock()
		return response(400)
	}
	h.polls[key] = lease
	h.mu.Unlock()
	result := h.events(req, name, id)
	lease.body = result.Body
	lease.release = func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.polls[key] == lease {
			delete(h.polls, key)
		}
	}
	result.Body = lease
	return result
}

func (h *Handler) events(req *http.Request, name string, id uint) *http.Response {
	ctx, cancel := context.WithTimeout(req.Context(), h.timeout)
	defer cancel()
	events, err := h.service.Events(ctx, name, id)
	if errors.Is(err, context.DeadlineExceeded) {
		return response(204)
	}
	if err != nil {
		return errorResponse(err)
	}
	// Raw payload bytes are base64-encoded by encoding/json, preserving binary data.
	type message struct {
		Sequence uint64 `json:"sequence"`
		Data     []byte `json:"data"`
	}
	messages := make([]message, 0, len(events))
	for _, event := range events {
		messages = append(messages, message{event.Sequence(), event.Data()})
	}
	return jsonResponse(200, messages)
}

var errMethod = errors.New("unsupported method")
var errSubscription = errors.New("missing topic or invalid subscriber id")

func subscription(req *http.Request, method string) (string, uint, error) {
	if req.Method != method {
		return "", 0, errMethod
	}
	name := req.URL.Query().Get("topic")
	raw := req.URL.Query().Get("id")
	if name == "" || raw == "" {
		return "", 0, errSubscription
	}
	for _, c := range raw {
		if c < '0' || c > '9' {
			return "", 0, errSubscription
		}
	}
	id, err := strconv.ParseUint(raw, 10, strconv.IntSize)
	if err != nil || id == 0 {
		return "", 0, errSubscription
	}
	return name, uint(id), nil
}

func errorResponse(err error) *http.Response {
	if errors.Is(err, errMethod) {
		return response(501)
	}
	if errors.Is(err, pubsub.ErrUnknownSubscriber) {
		return response(404)
	}
	return response(400)
}
func response(status int) *http.Response { return httpserver.NewResponse(status, nil, "") }
func jsonResponse(status int, value any) *http.Response {
	data, err := json.Marshal(value)
	if err != nil {
		return response(400)
	}
	return httpserver.NewResponse(status, data, "application/json")
}
