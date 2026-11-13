# Lab 1 pub-sub implementation

Accepted choices are recorded in [decision 0005](../../../knowledge/decisions/0005-pubsub-and-waiting-admission.md).

The [official assignment](lab1_instructions.pdf), pages 2–3, defines the broker
requirements. The general [pub-sub concept note](../../../knowledge/concepts/concurrency/publish-subscribe.md)
explains the vocabulary independently of this implementation.

## Packages and private state

- `src/pubsub` owns the broker without HTTP dependencies. `PubSubService` has only
  private fields. Topic/subscriber types, their maps, queue capacity, ID/sequence
  counters, channels, and polling flags are private.
- `Event` has private payload/sequence fields with `Data()` and `Sequence()`
  accessors. Publication copies input bytes; `Data()` returns another copy.
- `src/interfaces/pubsubapi` owns HTTP parsing, JSON responses, and poll deadlines.
  Its handler fields are private. No exported mutable broker collections exist.
- `main` constructs one broker and registers its four exact paths before the file
  fallback. File-only validation never applies to those routes.

The service methods are `Subscribe`, `Unsubscribe`, `Publish`, and `Events`.
They expose operations and errors instead of topic/channel internals. A single
mutex serializes ID assignment, topic membership, sequence assignment, and
nonblocking queue sends. Polling releases the mutex before waiting.

## Delivery contract

The application configures 64 pending events per subscriber and a 15-second poll
timeout. Those values are implementation choices. Each topic has its own increasing
sequence counter. Unknown-topic publication succeeds without retaining history.
New subscribers receive only later publications; there is no replay or persistence.
Empty topics retain their counter for the lifetime of the broker.

The HTTP adapter keeps a polling lease until the response body closes after the
transport finishes writing it. Overlapping requests cannot overtake an earlier
response still being written. Direct callers must close response bodies too.

A full subscriber queue drops that message for that subscriber, so sequence gaps
are allowed. Other subscribers still receive it. Unsubscribe removes membership
and closes a separate private signal to wake a waiting poll. It does not close
the event queue. One poll per subscriber is allowed; overlapping polls return 400.
A completed poll returns the first event and any further queued events available
when it drains the queue under the broker mutex.

## HTTP contract

| Request | Response |
| --- | --- |
| `POST /subscribe?topic=X` | 201 with JSON `{"id":N}`. |
| `POST /unsubscribe?topic=X&id=N` | 200 with no body. |
| `POST /publish?topic=X` | 200 with no body; raw request bytes are published. |
| `GET /events?topic=X&id=N` | 200 with a JSON array, or 204 when the poll times out. |

Events use `[{"sequence":1,"data":"aGVsbG8="}]`. JSON encodes byte slices as
base64, so the example data decodes to `hello`. This preserves arbitrary payload
bytes. The event shape is a documented application choice, not an assignment mandate.

Missing topics or malformed/zero/overflowing IDs return 400. Unknown subscribers
return 404, including IDs used with the wrong topic. Wrong methods on these routes
return 501. Concurrent polls for one subscriber return 400. A canceled context
ends a service wait; normal HTTP waits use the configured timeout.

## Admission while polling

The shared transport attaches a wait policy to each request context using
`concurrent/waiting`. The service calls `waiting.Wait` only around its blocking
channel selection. The policy releases the request's permit before that wait and
reacquires it before the service inspects or drains the queue. Validation, broker
state changes, JSON serialization, and response writing remain admitted work.

This applies the assignment's waiting-poll exception: publishers and ordinary
file requests can run even when more than ten event requests are waiting.
Direct service/Handle calls have no policy and wait normally. Wait policies must
run callbacks synchronously; do not nest waits or share another handler's context
with an independently running goroutine. The service depends on a generic
concurrency capability, not on HTTP; the transport knows nothing about broker
routes or topics.

Each response still closes its connection. Client-disconnect cancellation is not
automatically wired by this raw TCP server; the configured poll timeout bounds
broker waiting. General slow-client limits and connection reuse remain separate work.

## Checks

Service tests cover concurrent ordering/fanout, bounded queues, payload isolation,
subscription windows, timeout, overlapping polls, and unsubscribe waking a waiter.
HTTP tests cover response payloads, status codes, invalid input, and wrong methods.
A transport test uses a one-permit server and three waiting requests to verify
that a publisher can still enter and wake them. Existing file-server checks also
run under race detection.
