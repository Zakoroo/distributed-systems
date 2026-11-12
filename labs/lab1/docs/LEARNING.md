# Lab 1 learning map

Parent: [Knowledge base](../../../knowledge/README.md).

This map covers the concepts used by the **basic** Lab 1 implementation and its
tests. It is an introduction, not an exhaustive Go manual. General explanations
live in the concept notes; the mapping and walkthrough here may change with the
code. [Official instructions](lab1_instructions.pdf) define requirements, while
[design notes](DESIGN.md) define the chosen architecture.

This map lives under `docs/`; application code lives under `src/`. Build and run
from the lab root using `go build -o http_server ./src` and `./http_server 8080`
so the runtime `public/` directory remains relative to the correct location.

## Suggested study order

1. Start with [Go fundamentals](../../../knowledge/concepts/go/go-fundamentals.md):
   values and pointers, functions and types, then errors and cleanup.
2. Read [TCP](../../../knowledge/concepts/networking/tcp-and-connections.md) and
   [HTTP](../../../knowledge/concepts/networking/http-basics.md) to understand what
   crosses the network.
3. Read [streams and buffering](../../../knowledge/concepts/files/streams-and-buffering.md)
   and [files and paths](../../../knowledge/concepts/files/files-and-paths.md).
4. Read [routing and middleware](../../../knowledge/concepts/networking/routing-and-middleware.md)
   and the [regex guide](../../../knowledge/concepts/regex-basics-in-go.md).
5. Follow the [concurrency branch](../../../knowledge/concepts/concurrency/concurrency.md)
   in its numbered order.
6. Use [testing and evidence](../../../knowledge/concepts/go/testing.md) to understand
   how the design is checked. Then trace a request using the walkthrough below.

## Coverage of implementation concepts

| Code                                                                              | Concepts and explanations                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [main.go](../src/main.go)                                                         | [Packages, modules, CLI arguments, conversion, and formatting](../../../knowledge/concepts/go/modules-and-cli.md); [errors and construction](../../../knowledge/concepts/go/errors-and-cleanup.md).                                                                                                                                                                                                                                                                                                      |
| [httpserver/server.go](../../../httpserver/server.go)                             | [TCP listener, Accept, connections, and blocking](../../../knowledge/concepts/networking/tcp-and-connections.md); [goroutines](../../../knowledge/concepts/concurrency/goroutines-and-channels.md); [semaphore admission](../../../knowledge/concepts/concurrency/semaphores-and-conditions.md); [HTTP parsing/writing](../../../knowledge/concepts/networking/http-basics.md); [body ownership and defer](../../../knowledge/concepts/go/errors-and-cleanup.md).                                        |
| [httpserver/router.go](../../../httpserver/router.go)                             | [Function types and closures](../../../knowledge/concepts/go/functions-and-types.md); [regex compilation and matching](../../../knowledge/concepts/regex-basics-in-go.md); [routing, fallback, middleware, short-circuiting, and responsibility boundaries](../../../knowledge/concepts/networking/routing-and-middleware.md); [mutexes and snapshots](../../../knowledge/concepts/concurrency/mutexes-and-races.md); [slices and iteration](../../../knowledge/concepts/go/values-and-control-flow.md). |
| [httpserver/response.go](../../../httpserver/response.go)                         | [Status, media type, length, and framing](../../../knowledge/concepts/networking/http-basics.md); [byte readers, interfaces, and NopCloser](../../../knowledge/concepts/files/streams-and-buffering.md).                                                                                                                                                                                                                                                                                                 |
| [middleware/](../src/middleware)                                                  | [Middleware composition and early rejection](../../../knowledge/concepts/networking/routing-and-middleware.md); [path and extension policy](../../../knowledge/concepts/files/files-and-paths.md).                                                                                                                                                                                                                                                                                                       |
| [interfaces/fileapi/handler.go](../src/interfaces/fileapi/handler.go)             | [POST body streams and full reads](../../../knowledge/concepts/files/streams-and-buffering.md); [append versus replacement and file modes](../../../knowledge/concepts/files/files-and-paths.md); [error classification](../../../knowledge/concepts/go/errors-and-cleanup.md).                                                                                                                                                                                                                          |
| [concurrent/files/manager.go](../../../concurrent/files/manager.go)               | [Callbacks and interfaces](../../../knowledge/concepts/go/functions-and-types.md); [flags, permissions, and absolute paths](../../../knowledge/concepts/files/files-and-paths.md); [complete-operation locking](../../../knowledge/concepts/concurrency/keyed-locks-and-atomic-operations.md).                                                                                                                                                                                                           |
| [filepolicy/paths.go](../src/filepolicy/paths.go)                                 | [Strings, runes, slices, and control flow](../../../knowledge/concepts/go/values-and-control-flow.md); [allowlists, reserved names, normalization versus validation, and depth checks](../../../knowledge/concepts/files/files-and-paths.md).                                                                                                                                                                                                                                                            |
| [concurrent/locks/lockmap.go](../../../concurrent/locks/lockmap.go)               | [Generic comparable keys and zero values](../../../knowledge/concepts/go/functions-and-types.md); [mutex/RWMutex](../../../knowledge/concepts/concurrency/mutexes-and-races.md); [reference counts, holders, waiters, and lock lifetime](../../../knowledge/concepts/concurrency/keyed-locks-and-atomic-operations.md).                                                                                                                                                                                  |
| [concurrent/semaphores/semaphore.go](../../../concurrent/semaphores/semaphore.go) | [Permits, condition variables, Wait/Signal, and the recheck loop](../../../knowledge/concepts/concurrency/semaphores-and-conditions.md).                                                                                                                                                                                                                                                                                                                                                                 |

The [content-type mapping](../src/filepolicy/contenttypes.go) is also part of the policy
package: handlers use it for response metadata, and middleware uses it to reject
unsupported extensions. See [HTTP media types](../../../knowledge/concepts/networking/http-basics.md).

## Coverage of testing concepts

| Examples | Concepts |
| --- | --- |
| [Router tests](../../../httpserver/router_test.go) | Table-driven checks, concurrent configuration updates, and handler removal during dispatch. |
| [Transport tests](../../../httpserver/server_test.go) | Injected listeners, `net.Pipe`, channels, deadlines in tests, atomic counters/CAS, and checking the handler cap. |
| [File-server tests](../src/files_test.go) | Temporary directories, request objects without networking, complete append visibility, and goroutine coordination. |
| [File-manager tests](../../../concurrent/files/manager_test.go), [lock tests](../../../concurrent/locks/lockmap_test.go) | Path aliases, same-key exclusion, concurrent readers, independent keys, callback failures, and panic cleanup. |

Read [testing](../../../knowledge/concepts/go/testing.md),
[goroutines/channels](../../../knowledge/concepts/concurrency/goroutines-and-channels.md),
and [mutexes/atomics](../../../knowledge/concepts/concurrency/mutexes-and-races.md)
for the vocabulary behind these checks.

## Trace a basic request

Lab 1 configures the shared server with an admission limit of 10.

1. `main` parses the supplied port, constructs the server, registers method
   middleware, and installs the file handler with its validation as the fallback.
2. `Start` creates a TCP listener. `Serve` accepts connections and acquires a
   semaphore permit before starting each connection handler goroutine.
3. A buffered reader supplies bytes to `http.ReadRequest`. A parse failure
   produces 400; parsed requests proceed to dispatch.
4. Dispatch selects the first matching endpoint, or the separate fallback, and
   gathers matching registered middleware under the registry read lock. It
   releases that lock before invoking user code.
5. Global method validation permits GET/POST. For the file fallback, filename/
   depth and extension checks run before file access. Other endpoints do not
   automatically inherit those file-only rules.
6. GET reads the whole file under its shared key lock. POST first reads the body,
   then appends it under the file's exclusive key lock. Missing GET files produce
   404; other file failures use the assignment's 400 response.
7. The transport writes the response, closes its body and the connection, and
   cleans up the request body. The goroutine releases the semaphore permit.

For a concrete example, POSTing `hello` to `/notes.txt`, then GETting that path,
should make the submitted bytes visible. Posting ` world` again appends it in
the current design. Appending is an implementation choice, not a universal
definition of POST.

## Package boundaries and current checks

`httpserver` is a shared module. Lab 1 imports it alongside `src/interfaces/fileapi`,
`src/middleware`, and `src/filepolicy`. The handlers and middleware both depend on
`filepolicy` without depending on each other. The shared server and generic
concurrency code have no import of Lab 1.

The earlier missing-constructor-limit issue, **BL-009**, is fixed. The constructor
stores the configured positive limit; transport tests check limits of two and
ten, slot recovery, and rejection of zero. Lab 1 still supplies ten. The shared
server's configuration must be distinguished from the assignment's chosen cap.

## Scope and later material

Connection reuse and slow-client limits remain deferred work. Their vocabulary
is introduced in the networking and concurrency notes; those features are not
claimed as implemented. Pub-sub, cloud deployment, and container tooling are
outside this basic-part learning map.

