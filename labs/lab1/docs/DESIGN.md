# Lab 1 server design

Status: Approved and implemented for routing, middleware, and generic keyed
locking. Connection reuse is a separate increment; slow-client improvements are
explicitly delayed. See [decision 0002](../../../knowledge/decisions/0002-lab1-server-and-locking.md)
and the package-boundary update in [decision 0003](../../../knowledge/decisions/0003-shared-server-and-lab1-packages.md).
Source/document placement follows [decision 0004](../../../knowledge/decisions/0004-lab1-source-and-docs.md).

Scope: the basic HTTP file server. See the [assignment guide](../AGENTS.md) and
[official instructions](lab1_instructions.pdf) for requirements.

## Package responsibilities

- `src/main.go`: CLI configuration and connecting the components.
- Root `httpserver/`: raw TCP networking, request parsing, ordered routing,
  middleware composition, response writing, configurable admission, and connection
  cleanup. It has its own `go.mod` and depends only on generic concurrency code.
- `src/interfaces/fileapi/`: application GET/POST operations using the reusable file
  manager and `filepolicy` content types.
- `src/middleware/`: the basic lab's method, filename/depth, and extension checks.
- `src/filepolicy/`: pure filename/depth rules and extension-to-content-type mappings,
  shared by the application handlers and middleware.
- `concurrent/locks/`: generic `LockMap[K comparable]` read/write synchronization.
- `concurrent/files/`: synchronized file operations, without application-specific
  naming or path-validation rules.

`main` imports the application packages and shared server. Middleware and file
handlers both import `filepolicy`; neither imports the other. `httpserver` and
`concurrent` never import Lab 1. Sub-packages inside Lab 1 share its existing
module. The workspace connects all three modules.

The module file and agent guide stay at the Lab 1 root. Source and tests live in
`src/`; instructions and learning/design documents live in `docs/`. Build from
`labs/lab1` with `go build -o http_server ./src` and run the executable from there,
so the relative runtime path `./public/` keeps its meaning. Moving source does not
move the runtime data or require a new module/workspace entry.

The directory name `interfaces` identifies HTTP-facing application adapters; it
does not require every declaration to be a Go interface type. The group chose
it instead of Go's special `internal` import boundary. A future `pubsubapi` belongs
under it, but is not part of this increment.

This is an extensibility change. Workload and measurable scalability targets are
still open under BL-001; no throughput improvement is claimed.

## Shared `httpserver` endpoint and middleware API

```go
type Handler func(*http.Request) *http.Response
type Middleware func(Handler) Handler

func New(maxConcurrentRequests uint) (*Server, error)
func (s *Server) RegisterEndpoint(pattern string, handler Handler) error
func (s *Server) UnregisterEndpoint(pattern string) error
func (s *Server) SetFallback(handler Handler) error
func (s *Server) RegisterMiddleware(pattern string, middleware Middleware) error
func Chain(handler Handler, middleware ...Middleware) Handler
```

- Patterns are regular expressions matched against the entire `req.URL.Path`,
  excluding queries. They are compiled once at registration. `/files/.*` matches
  descendants of `/files/`; `/health` matches that exact path. Literal regex
  metacharacters must be escaped when intended as ordinary characters. See
  [regex basics in Go](../../../knowledge/concepts/regex-basics-in-go.md) for a short
  explanation with examples.
- `compilePattern` rejects empty patterns and calls `regexp.Compile` on
  `\A(?:` + pattern + `)\z`. The boundaries require a whole-path match; grouping
  keeps them around every alternative in the supplied pattern. Invalid syntax
  is returned as a registration error. Dispatch reuses the prepared matcher with
  `MatchString(req.URL.Path)`; it does not compile the pattern on every request.
- Endpoint pattern strings are unique registration keys. Empty/invalid patterns,
  duplicate endpoints, and nil handlers/middleware are rejected. Removing an
  unknown pattern returns an error. Function identity is not used for removal.
- Endpoints run first-match in registration order. The fallback is stored
  separately and is considered only when no endpoint matches, including after
  runtime registrations. Its generic default response is 501.
- Lab 1 sets a file handler as its fallback, preserving the assignment's GET/POST,
  400, 404, and 501 behavior. JPEG extensions both map to `image/jpeg`.
- Matching registered middleware runs in registration order for both endpoints
  and fallback. Reusing a middleware pattern is allowed. A middleware may return
  a response without invoking the next handler.
- `Chain(h, a, b)` runs `a(b(h))`: requests enter A, then B, then the handler;
  response processing unwinds through B, then A. This is also how registered
  middleware is composed.
- The basic application registers method validation globally. It uses `Chain`
  to attach filename/depth and extension checks only to the file fallback, so a
  separately registered `/health` endpoint does not need a file extension.
- Dispatch selects the handler and matching middleware from one consistent
  registry snapshot before invoking user code. Registry locks are not held while
  factories or handlers run. Unregistering does not cancel already-selected
  handlers. Later dispatches see the updated registry.

## Transport and ownership

`Start(address)` creates a raw TCP listener and calls `Serve(listener)`. `Serve`
supports an externally supplied listener, which its caller owns. Closing that
listener stops acceptance; already-admitted requests finish normally.

The constructor validates and stores a positive concurrency limit. Lab 1 supplies
10 in `src/main.go`; the shared server does not hard-code an assignment-specific cap.
BL-009 fixed the missing assignment of that value. Tests verify admission and
slot recovery with limits of both 2 and 10, and reject a zero limit.

- Each admitted connection gets a goroutine under a semaphore, with at most the
  configured number of active handlers. One request/response is served per connection.
- Malformed HTTP receives 400. An empty connection ending at EOF is closed.
- The transport advertises connection closure, writes the response, and closes
  its body. Handlers must return a non-nil, per-request response.
- The transport owns the parsed request body. Cleanup closes the connection
  before the request body, so draining unread body bytes cannot keep a finished
  handler waiting on the client.
- Direct callers of `Handle(req)` own both request and response bodies; this
  method dispatches an already-parsed request without doing networking.

## File policy and locking

The group explicitly selected middleware as the filename/depth validation
boundary. The pure rules live in Lab 1's `filepolicy` package. The file manager
and `LockMap` contain no such policy definitions or checks. Direct file-manager
callers are responsible for their own access policy.

Filename/depth validation examines the URL path relative to the public directory
before joining and cleaning it. Extension validation rejects unsupported types
before attempting file access. POST continues to append; the assignment does not
explicitly choose append versus replacement semantics.

The file manager converts lexical path aliases to absolute paths so those calls
share a lock key. This is normalization, not name/depth validation or symlink
resolution. Each full read holds a shared lock; each full write holds an exclusive
lock. A generic lock entry counts holders **and waiters** before waiting on its
read/write mutex. It is removed only when that count reaches zero, preventing
waiters and newcomers from accidentally using different locks for the same key.
The registry mutex is never held while waiting for a resource lock or running a
callback. Callback errors and panics both release the resource lock.

## Deferred connection reuse — BL-008

Reuse needs a request loop using one buffered reader, correct response framing,
consumed or safely discarded request bodies, and bounded idle waits. An endpoint
option may permit reuse but cannot override client close requests or parsing and
writing failures. The option API is deferred with the feature; it is not exposed
as a nonfunctional flag in this increment.

## Delayed slow-client improvements — BL-007

Ten slow clients can still occupy all ten handler slots. Go already parks network
I/O waiters using its [network poller](https://go.dev/src/runtime/netpoll.go).
The planned first step is configurable read/write deadlines and header/body size
limits, using a concrete workload to select and verify the limits. The group has
explicitly delayed this work.

A registry alone cannot select readable `net.Conn` values using channel `select`.
Readability can mean one byte or EOF, not a complete HTTP request. A custom
readiness scheduler would need bounded connection/request state and polling
integration; it remains an exploratory alternative. Extra parsing goroutines must
not silently bypass the assignment's handler cap.

## Verification

- Root `httpserver/router_test.go`: first-match order, whole-path matching, fallback
  replacement, removal, middleware order/scoping/short-circuiting, and concurrent
  registry updates. A handler can unregister itself without a registry deadlock.
- Root `httpserver/server_test.go`: malformed HTTP, response ownership/closure,
  unread-body cleanup, configured handler caps, zero-limit rejection, and slot
  recovery with one more connection than the configured limit.
- `src/files_test.go`: required MIME/status behavior, middleware validation, custom
  endpoint isolation, POST create/append and truncated bodies, concurrent POST/GET.
- `concurrent/locks/lockmap_test.go`: generic keys, exclusive updates, concurrent
  readers, independent keys, and cleanup after callback errors/panics.
- `concurrent/files/manager_test.go`: caller-owned filename policy and atomic
  complete file versions when callers use lexical aliases.

Run affected modules and consumers with race detection as described in the root
agent guide. Network behavior can be exercised with in-memory connections and
an injected listener, without opening TCP ports.
