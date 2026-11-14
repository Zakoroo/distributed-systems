# 0006: Pass request slots explicitly while reading events

Parent: [Design decisions](decisions.md).

Status: Withdrawn by [decision 0007](0007-return-to-basic-server.md)

This records a previous readability experiment. Its source/tests were removed
with the bonus reset; the handler signature is again request-only. The historical
decision below is not an instruction to reintroduce slots.

Date: 2026-11-14

## Context

The group found the broker and its admission handling difficult to explain to
novice Go readers. The transport stored a callback policy inside context, then
the broker retrieved it through `concurrent/waiting`. Understanding one channel
wait required tracing a context value and nested function calls across modules.
The user approved a simpler interface and beginner-oriented explanations.

## Decision

- HTTP handlers receive `(*http.Request, *httpserver.RequestSlot)`. Middleware
  forwards both arguments. The transport owns the slot and releases it after
  response/connection cleanup. Direct `Server.Handle(req)` supplies nil.
- `PubSubService.Events` takes an explicit `WorkSlot` argument. This small local
  interface requires only `Release()` and `Acquire()`; it imports no HTTP types.
  Direct broker callers pass nil. A nil concrete request slot is also harmless.
- Express the service read as claim, wait, collect, with named private helpers.
  The wait helper releases the slot, defers acquisition, and selects among event,
  cancellation/deadline, and unsubscribe. No broker mutex is held while waiting.
- Keep admission reacquisition before collection and all return paths. Keep the
  existing service/HTTP poll guards and response-body cleanup through final flush.
- Remove the obsolete `concurrent/waiting` package. Context now carries only
  cancellation/deadlines in this flow; no admission policy is stored in it.
- Preserve HTTP behavior, bounded subscriber queues, ordering, private state,
  one goroutine per request, and the ten-slot limit with its waiting-poll exception.

This supersedes only the context-policy mechanism in
[decision 0005](0005-pubsub-and-waiting-admission.md), not its other broker choices.

## Alternatives considered

A job scheduler would add pending states, wakeups, and deferred response ownership.
Reserving two event handlers would limit simultaneous polls and could delay a
subscriber with available messages behind quiet subscribers. Both remain proposals.
Keeping the old HTTP handler signature through a second slot-aware registration
API would preserve compatibility but introduce two handler/registration forms.

## Historical consequences and checks

Admission becomes visible in function signatures and the wait's source code.
Handlers that do not lend out their slot receive an unused argument; all existing
handlers/middleware and their tests were migrated together. This is an intentional
Go API change with no intended client-visible HTTP change. Slots belong to one
goroutine and must not be copied or shared with independent work.

The experiment's slot-handoff tests checked reacquisition before return on success
and error. Its admission and full HTTP-path tests exercised three waiting polls
and a publisher through a one-slot server. Those bonus-specific checks were removed
with the experiment. The basic file-server, router, transport, and locking tests
remain. No throughput improvement was claimed.

See the [current bonus proposal](../../labs/lab1/docs/PUBSUB.md).
