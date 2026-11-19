# Lab 1 design

[main.go](../src/main.go) configures the shared HTTP server. The server accepts
raw TCP connections, takes a semaphore permit, then starts a goroutine.
Each goroutine parses one HTTP request, sends the response, closes the
connection, and releases the permit. Lab 1 allows up to ten workers, including
workers waiting in an events poll.

## Routing and middleware

The router chooses the first endpoint whose pattern matches the whole URL path.
Query parameters are separate. When no endpoint matches, the router uses the
file fallback. Matching middleware runs in registration order. Each middleware
wraps a handler and can return a response early or call the handler it wraps.

The method check applies to every route. Path and extension checks wrap the
file fallback. [Application wiring](../src/main.go) · [Validation middleware](../src/middleware/)

## File access

The URL path is relative to `./public/`. POST reads the request body before
taking the file's exclusive lock, then creates or appends to the file.
GET holds a shared lock through the full read. Readers of the same file can
overlap, and a reader sees a complete version before or after an append.
Different normalized paths use separate locks.

Missing files return 404 on GET. Invalid paths, unsupported extensions, and
file operation errors return 400. Unsupported methods return 501.
[File handler](../src/interfaces/fileapi/handler.go)

[File tests](../src/files_test.go) cover validation, response types, and complete
reads during concurrent appends.

## Pub-sub broker

The [broker](../src/interfaces/pubsubapi/broker.go) stores topics and subscriber
queues in memory. A subscription creates an ID and a buffered channel with
space for 64 events. A restart clears broker state.

| Request | Success |
| --- | --- |
| POST `/subscribe?topic=X` | 201 with `{"id":N}` |
| POST `/publish?topic=X` | 200. The request body is the payload. |
| GET `/events?topic=X&id=N` | 200 with one event or 204 after 15 seconds |
| POST `/unsubscribe?topic=X&id=N` | 200 and the queue closes |

One mutex protects membership, topic sequence numbers, and queue sends.
Publishing offers the event to each current subscriber. A full queue skips the
new event for that subscriber, so their sequence numbers can have gaps.

An event request waits outside the mutex. It reads one event from the queue or
returns 204 on timeout. Unsubscribe removes membership and closes the queue
under the publication mutex. After waking, the poll checks membership again.
An ended or unknown subscription returns 404.

A client should finish reading each response before polling again with the same
ID. The broker rejects overlapping queue waits with 400. Each waiting poll
occupies one of the ten server workers, so ten waiting polls can delay publishers.

Event replies include a sequence number and base64-encoded payload bytes. The
[broker tests](../src/interfaces/pubsubapi/broker_test.go) check queue bounds,
concurrent ordering, wakeups, timeouts, and HTTP errors. The
[application test](../src/pubsub_test.go) checks the registered routes.

## Container setup

The [Dockerfile](../Dockerfile) packages the server and creates its storage
directory. The [README](../README.md) has commands to build, run, and test it.
