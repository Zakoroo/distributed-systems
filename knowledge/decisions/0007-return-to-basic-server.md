# 0007: Return to the basic server before restarting pub-sub

Parent: [Design decisions](decisions.md).

Status: Accepted; rollback completed, then the approved small broker implemented

Date: 2026-11-14

## Context

The group still found the bonus implementation too difficult after documentation
and an explicit-slot refactor. The user requested returning to the implementation
before the bonus and finding a simpler approach, accepting lower performance.
Understanding the code takes priority over extensibility for this increment.

## Decision

- Restore the basic Go implementation from commit `870d6bd`, before the bonus.
  Keep the shared router, middleware, file policy, generic locking, atomic file
  access, raw TCP transport, and ordinary ten-handler semaphore.
- Remove the broker/API code, bonus-only tests, context wait policy, and explicit
  request-slot machinery. Restore `func(*http.Request) *http.Response` handlers.
- Leave Git history, unrelated settings, and reusable concept explanations intact.
  Mark decisions 0005 and 0006 withdrawn; reopen backlog item BL-012.
- The user subsequently approved the smaller broker. Implement its state and four
  HTTP handlers together in `src/interfaces/pubsubapi/broker.go`, leaving the
  shared server and its request-only handler interface unchanged.
- Use one mutex, one bounded message channel per subscriber, and an ordinary
  event/timeout select. Unsubscribe removes membership and closes the queue under
  the publication mutex; event reading rechecks membership after wakeup.
- Return one JSON event per poll. Clients keep one outstanding poll per ID and
  finish the response before polling again. The reader flag rejects overlapping
  waits and does not extend through transport writing. No response-lifetime or
  request-slot framework is added.

## Simpler direction

The implementation retains the normal handler limit even while a poll
waits. The official instructions allow excluding waiting polls but do not require
it. One broker mutex, bounded subscriber channels, and an ordinary event/timeout
select can then live behind the existing HTTP handler API.

If ten polls occupy all handlers, a publisher waits for capacity. This is an
explicit accepted limitation. The broker must still meet the bonus's
ordering, queue, status-code, and nonblocking-publication requirements. Ordinary
short polling that returns immediately is not being claimed as a complete bonus.

## Verification

At the reset, all 17 Go source/test files matched `870d6bd` exactly. The new bonus
adds one broker file, its tests, route wiring, and an application integration test.
Shared `httpserver/` and `concurrent/` code still match the basic baseline.

[Broker checks](../../labs/lab1/src/interfaces/pubsubapi/broker_test.go) cover HTTP
errors, queue bounds, subscription windows, binary payloads, concurrent order and
membership changes, publication/unsubscribe wakeups, and timeout recovery.
[Application checks](../../labs/lab1/src/pubsub_test.go) cover the four routes and
their coexistence with file operations. Run the module race tests, vet, build,
and documentation link checks when validating the implementation.
