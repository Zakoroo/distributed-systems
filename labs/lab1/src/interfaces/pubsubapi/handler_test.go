package pubsubapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Zakoroo/distributed-systems/httpserver"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/interfaces/pubsubapi"
	"github.com/Zakoroo/distributed-systems/labs/lab1/src/pubsub"
)

func TestHTTPContract(t *testing.T) {
	service, _ := pubsub.New(4)
	api, _ := pubsubapi.New(service, time.Millisecond)
	call := func(handler httpserver.Handler, method, path, body string, status int) *http.Response {
		t.Helper()
		response := handler(httptest.NewRequest(method, path, strings.NewReader(body)))
		t.Cleanup(func() { response.Body.Close() })
		if response.StatusCode != status {
			t.Fatalf("%s %s: %d, want %d", method, path, response.StatusCode, status)
		}
		return response
	}
	response := call(api.HandleSubscribe, "POST", "/subscribe?topic=news", "", 201)
	var subscription struct {
		ID uint `json:"id"`
	}
	if err := json.NewDecoder(response.Body).Decode(&subscription); err != nil || subscription.ID == 0 {
		t.Fatal("missing subscription id", err)
	}
	call(api.HandlePublish, "POST", "/publish?topic=news", "hello", 200)
	response = call(api.HandleEvents, "GET", "/events?topic=news&id=1", "", 200)
	call(api.HandleEvents, "GET", "/events?topic=news&id=1", "", 400)
	var messages []struct {
		Sequence uint64 `json:"sequence"`
		Data     []byte `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&messages); err != nil || len(messages) != 1 || string(messages[0].Data) != "hello" || messages[0].Sequence != 1 {
		t.Fatal("invalid events", err)
	}
	response.Body.Close()
	call(api.HandleEvents, "GET", "/events?topic=news&id=1", "", 204).Body.Close()
	call(api.HandleEvents, "GET", "/events?topic=news&id=-1", "", 400)
	call(api.HandleEvents, "GET", "/events?topic=news&id=0", "", 400)
	call(api.HandleEvents, "GET", "/events?topic=news&id=999", "", 404)
	call(api.HandleEvents, "GET", "/events?id=1", "", 400)
	call(api.HandleSubscribe, "GET", "/subscribe?topic=news", "", 501)
	call(api.HandlePublish, "GET", "/publish?topic=news", "", 501)
	call(api.HandleUnsubscribe, "GET", "/unsubscribe?topic=news&id=1", "", 501)
	call(api.HandleEvents, "POST", "/events?topic=news&id=1", "", 501)
	call(api.HandleUnsubscribe, "POST", "/unsubscribe?topic=news&id=1", "", 200)
	call(api.HandleEvents, "GET", "/events?topic=news&id=1", "", 404)
	call(api.HandleUnsubscribe, "POST", "/unsubscribe?topic=news&id=1", "", 404)
}
