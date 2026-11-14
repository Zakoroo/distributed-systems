# Lab 1: the small pub-sub broker

Status: implemented. All broker code lives in
[pubsubapi/broker.go](../src/interfaces/pubsubapi/broker.go). The shared HTTP
server and its request-only handler interface are unchanged. The
[official instructions](lab1_instructions.pdf), pages 2–3, define the requirements;
[decision 0007](../../../knowledge/decisions/0007-return-to-basic-server.md)
records why we chose this deliberately limited version.

## Start with Alice and Bob

1. Alice subscribes to `weather`. The broker creates a message queue and returns
   her subscriber ID. That queue stays in memory between requests.
2. Bob publishes `rain` to `weather`. The broker puts the publication into each
   current subscriber's queue, provided that queue has space.
3. Alice requests events using her ID. Her request removes one message from her
   queue and returns it. If the queue is empty, the request waits for a message
   or a timeout. Alice finishes reading that response before asking again.
4. Alice unsubscribes. The broker removes her membership and closes her queue.
   An events request already waiting on that queue wakes and returns 404.

A subscription is stored data, not a background goroutine. The HTTP server starts
one goroutine per admitted request. An events request waits in that same goroutine.
For the independent terminology, read
[publish-subscribe concepts](../../../knowledge/concepts/concurrency/publish-subscribe.md).

## What is stored?

All fields below are private:

| Type | Purpose |
| --- | --- |
| `Broker` | Topic map, subscriber-ID counter, queue capacity, timeout, and one mutex. |
| `topic` | Its sequence counter and map of subscriber IDs. |
| `subscriber` | Its bounded message channel and a flag identifying a waiting reader. |
| `event` | Publication sequence number and payload bytes. |

`New(64, 15*time.Second)` configures 64 queued messages per subscriber and a
15-second event wait. These are our choices, not assignment-prescribed values.
Constructing the broker or a subscription starts no additional goroutine.

The broker is in memory only: subscriptions and messages disappear when the
process stops. A subscriber receives only later publications; there is no replay.
Empty topics retain their sequence counter until the broker is discarded.

## Read the four handlers in order

**Subscribe:** validate the method/topic, lock the broker, create a topic if
needed, allocate an ID and channel, unlock, then return the ID. A channel is a
FIFO queue: receiving removes the oldest queued value.

**Publish:** read the HTTP body before taking the mutex. Under the mutex, assign
the topic's next sequence number and try a send to each subscriber. `default`
means skip a full queue instead of waiting. This serializes concurrent publishers
so all subscribers' queued messages have the same topic order. A slow subscriber
can miss messages and observe gaps in sequence numbers. `ReadAll` owns the payload
bytes; broker code never changes them afterward.

**Events:** this is the only handler that waits:

1. Under the mutex, find the subscriber and set its `polling` flag. Another
   concurrent wait for that ID returns 400.
2. Release the mutex and wait in this select:

   ```go
   select {
   case message, received = <-current.messages:
   case <-time.After(b.pollTimeout):
   }
   ```

3. Lock briefly again to clear the flag and check membership. Return 404 if the
   subscription ended, 204 if no event was received, or 200 with one event.

There is no polling loop or new worker here. A channel receive can park the
current goroutine until it can proceed. The mutex is released so another
admitted request can publish or unsubscribe. The timeout channel becomes ready
when the duration elapses. See [channels and select](../../../knowledge/concepts/concurrency/goroutines-and-channels.md)
and [time.After](https://pkg.go.dev/time#After).

**Unsubscribe:** under the same mutex used by publication, remove membership and
close the message channel. Publication cannot send to a removed queue while close
is happening because both operations take that mutex. Closed queues may still
contain buffered messages, so Events rechecks membership even after receiving a
value. A closed, empty channel returns `received == false` immediately; it does
not insert a fake event. See [channel close](https://go.dev/ref/spec#Close).

## The simple concurrency limit

Every request, including an event wait, counts toward the server's existing limit
of ten. The assignment permits excluding waiting polls, but does not require it.
We leave the server's semaphore and handler interface unchanged.

**If ten polls are waiting, a publisher must wait until a handler finishes.**
This is the accepted simplicity tradeoff. A timeout ends channel waiting; socket
reads/writes can take longer. Disconnect detection and slow-client deadlines are
not added by this broker.

Each subscriber client must keep **one outstanding events request per ID** and
finish reading its response before starting the next. The `polling` flag rejects
overlapping queue waits and is cleared before the HTTP response is written.
It is not a guard covering network output. This version has no special response
lifetime wrapper; it relies on the sequential client poll cycle for delivery
order across responses. Different subscribers may poll concurrently.

## HTTP responses

| Request | Success |
| --- | --- |
| `POST /subscribe?topic=X` | 201 with `{"id":N}`. |
| `POST /unsubscribe?topic=X&id=N` | 200 with no body. |
| `POST /publish?topic=X` | 200 with no body; the request body is the payload. |
| `GET /events?topic=X&id=N` | 200 with one event object, or 204 after an empty wait. |

An event looks like `{"sequence":1,"data":"cmFpbg=="}`. JSON encodes byte slices
as base64; here the bytes decode to `rain`. This preserves binary data. One poll
returns one event, not an array or a batch. See
[JSON and byte payloads](../../../knowledge/concepts/go/json-and-payloads.md).

A missing topic or missing/malformed/zero/overflowing ID returns 400. A syntactically
valid but unknown subscriber, including an ID used with the wrong topic, returns
404. Wrong methods return 501. Publication without subscribers succeeds and
retains no history. Timeout ends a poll, not the subscription.

The four endpoint paths are registered before the file fallback, on the same
executable/port. File-only path validation applies to the fallback; topic names
are not treated as filenames. The public handlers return responses using the
existing `httpserver.NewResponse` helper.

## Checks

[Broker tests](../src/interfaces/pubsubapi/broker_test.go) cover method/input errors,
timeout/recovery, independent bounded queues, subscription windows, binary/empty
payloads, concurrent publisher order, wakeups, overlapping waits, and concurrent
subscription changes. [Application tests](../src/pubsub_test.go) check endpoint
registration, topic/file-policy separation, JSON responses, and the file fallback.
The unchanged [transport tests](../../../httpserver/server_test.go) cover raw HTTP
handling, connection cleanup, and the configured handler cap. Tests use request
objects and in-memory connections, rather than external listeners.
