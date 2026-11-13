# 0005: Add a private-state pub-sub broker with long polling

Parent: [Design decisions](decisions.md).

Status: Accepted

Date: 2026-11-13

## Context

The group approved the pub-sub service/API split, per-subscriber queues, sequence
numbers, and long polling, with an explicit requirement to keep internal fields
private. Waiting polls must leave publishers able to enter the server.

## Decision

- Keep topic/subscriber state private in `src/pubsub`. Expose service operations
  and immutable event accessors; copy payloads at publication and on access.
- Protect membership changes, ID/sequence assignment, and nonblocking fanout with
  one service mutex. Use one bounded queue per subscriber and drop a message for
  a full queue. Sequence numbers are per topic; missing sequence numbers are allowed.
- Configure queue capacity 64 and poll timeout 15 seconds in the application.
  These are choices, not fixed values prescribed by the assignment.
- Permit one active poll per subscription. Reject overlapping polls with 400;
  hold the HTTP polling lease until response completion to avoid overlapping
  delivery. Signal unsubscribe separately so it wakes a poll without closing a
  channel into which publishers might send.
- Place HTTP adapters under `interfaces/pubsubapi`, on the same server/port as
  file operations. JSON events contain sequence numbers and base64 payload data.
- Use the generic `concurrent/waiting` context policy so transport admission is
  released only during the broker's blocking channel wait. Reacquire it before
  queue processing and writing. Broker code does not import the HTTP transport.
- Keep connections close-after-response. General keep-alive, slow-client controls,
  cloud deployment, and containerization remain separate work.

## Alternatives considered

A topic-wide queue with subscriber lists requires tracking consumption, queue
retention, and individual subscriber limits. Per-subscriber queues make those
rules easier to implement and explain. Streaming needs a different response
writing contract; long polling fits the current handler interface.

## Consequences

The broker is in-memory with no replay guarantee or persistent subscriptions.
Publication does not wait for queue capacity. Empty topics retain sequence
metadata for the service lifetime. Polls expire or wake on unsubscribe; raw TCP
client-disconnect detection is not automatically propagated as context cancellation.

See the [pub-sub design](../../labs/lab1/docs/PUBSUB.md) and
[README examples](../../labs/lab1/README.md#pub-sub-client-example).
