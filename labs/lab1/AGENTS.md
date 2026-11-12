# Lab 1 guide for agents

Source: [Lab 1: HTTP Server and Pub-Sub Broker](docs/lab1_instructions.pdf).
This guide summarizes the supplied assignment. Consult the PDF for exact wording
and details relevant to a change. The root [agent guide](../../AGENTS.md) also applies.
For beginner explanations and concept-to-code coverage, use the
[Lab 1 learning map](docs/LEARNING.md).
For collaborator setup, build/run commands, and curl examples, use the
[Lab 1 README](README.md).

## Layout and running

- `src/` contains the executable, tests, and application sub-packages.
- `docs/` contains the official instructions, design, and learning map.
- `go.mod` and this guide stay at the lab root. The workspace entry is still
  `./labs/lab1`; application import paths include `/labs/lab1/src/`.
- From `labs/lab1`, build with `go build -o http_server ./src` and run
  `./http_server 8080`. The runtime `public/` directory is relative to that working
  directory, not to `src/`.

## Required HTTP server — PDF pages 1–2

- Implement the server in Go. The executable must be named `http_server` and take
  its listening port as the first command-line argument. Do not hard-code a port
  or assume a particular server/client IP address.
- Use `net` for networking, such as `net.Listen("tcp", address)`. `net/http` is
  permitted for parsing and working with HTTP objects, not for networking.
  Do not use `http.ListenAndServe`, `http.Serve`, or web frameworks such as Gin
  or Martini to replace the assignment's server implementation.
- Handle each new client request in a goroutine. Limit active request handlers to
  **10**; wait for a handler to finish before starting one that would exceed this
  limit. The PDF sometimes calls these child processes, but specifies goroutines.
- Implement GET and POST; respond with **501** for other HTTP methods.
- Serve local files with these content types:

  | Extension | Content-Type |
  | --- | --- |
  | `.html` | `text/html` |
  | `.txt` | `text/plain` |
  | `.gif` | `image/gif` |
  | `.jpeg`, `.jpg` | `image/jpeg` |
  | `.css` | `text/css` |

- Respond with **400** for unsupported file extensions, malformed requests or
  headers, and other errors not assigned a different response by the instructions.
  Missing requested files require **404**. Errors must produce well-formed HTTP
  responses, rather than just closing the connection or logging an error.
- Store POSTed files so they are available through subsequent GET requests.
  Concurrent POSTs and GETs targeting one file must be synchronized: a GET must
  see a complete file version, never an incomplete write or a mix of versions.

## Optional pub-sub broker — PDF pages 2–3

This bonus is worth up to 7 extra points. Its inclusion has not been selected by
the group. If selected, it must share the file server's `http_server` executable
and port and follow the same networking restrictions.

| Request | Required behavior |
| --- | --- |
| `POST /subscribe?topic=X` | Register a subscriber; return 201 and its ID, for example `{"id":N}`. |
| `POST /unsubscribe?topic=X&id=N` | Remove the subscriber. |
| `POST /publish?topic=X` | Publish the request body to current subscribers of the topic. |
| `GET /events?topic=X&id=N` | Deliver messages promptly through long polling, streaming, or another explainable approach. |

- Long polling holds the request until a message or timeout; a timeout without
  messages returns **204**. The diagram's 15-second timeout is an example, not a
  fixed requirement. Replies use client-initiated connections.
- A missing topic or invalid ID requires **400**; an unknown subscriber requires
  **404**; other methods require **501**.
- Protect shared topics, subscriber state, and queues against concurrent access.
  Assign each message a sequence number and preserve topic message order for
  subscribers, including with concurrent publishers.
- Subscribers may join or leave at any time and receive only messages published
  during their subscription. Each subscriber must have a bounded queue.
  Publishing must not wait on a full subscriber queue; skipping that subscriber
  for that message is allowed.
- Long-poll requests waiting for messages may be excluded from the 10-handler
  limit. This is a specific exception, not permission to remove the general cap.

## Other scope and assessment — PDF pages 3–4

- Cloud deployment is a separate bonus worth up to 3 extra points. The PDF
  describes AWS Learner Lab access; deployment is not part of the required server.
- Containerization is strongly recommended, with Docker, LXC, and Vagrant listed
  as options. It is not stated as a mandatory requirement.
- No tests are supplied. The group must establish its own functional checks.
  Cover protocol responses, POST/GET behavior, the concurrency cap, and atomic
  file reads/updates when validating relevant changes.
- Assessment is a 30-minute group presentation with code questions and tests run
  against the implementation. Prepare a working demo and explainable code.

## Implementation and design decisions

- The group approved and implemented a shared `httpserver` module, ordered endpoint
  registration, separate fallback, path-scoped middleware, and generic
  `locks.LockMap[K]`. Read the [design notes](docs/DESIGN.md) for API and ownership
  contracts. The [accepted decision](../../knowledge/decisions/0002-lab1-server-and-locking.md)
  records the chosen validation boundary. [Decision 0003](../../knowledge/decisions/0003-shared-server-and-lab1-packages.md)
  updates the package placement.
- `src/interfaces/fileapi` owns file-request handling, `src/middleware` applies
  request validation, and `src/filepolicy` owns filename/depth rules and content-type mappings.
  These are ordinary Go packages, not `internal` packages; preserve their intended
  boundaries through dependency direction. Shared modules must not import Lab 1.
- The file manager and generic locker enforce no filename/depth policy. The
  constructor issue BL-009 is fixed and covered by configured-limit checks.
- Connection reuse is a separate future increment; slow-client concurrency
  improvements were approved in principle but explicitly delayed. Neither is
  part of the current implementation. Connections close after one response.
- The PDF requires POST storage and atomic visibility but does not explicitly
  choose append versus replacement semantics. The current implementation appends;
  do not present that choice as an assignment requirement.
- Pub-sub, its delivery strategy, cloud deployment, and container tooling have
  not been chosen. Optional requirements do not authorize implementing them.
- Scalability work must respect the assignment's handler limit and networking
  rules. Define the workload and success criteria before redesigning.
