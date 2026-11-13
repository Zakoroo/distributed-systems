# Backlog

Shared ideas and work for the student group and agents. Items are proposals until
selected by a group member; placement in Ready does not authorize an agent to
start work automatically.

## How to use

- Capture a short title, why it matters, and what completion would look like.
- Assign the next unused `BL-NNN` ID. Keep IDs stable when editing or moving items.
- Move items between the sections below as their status changes. Add an owner
  when useful and link to relevant notes, decisions, or implementation evidence.
- Ideas may need clarification. Ready items have a clear, actionable scope.
  In progress means someone has picked up the work. Done items include a brief
  outcome or link to the result.

## Ideas

- [ ] **BL-001: Define the scope of the scalability redesign**
  - Why: the group plans to refactor for scalability; a concrete target will help
    guide the implementation and avoid speculative changes.
  - Done when: the target lab/component, workload, constraints, and success
    criteria are documented, with a plan to measure the current behavior.
  - Constraint: Lab 1 redesigns must preserve the
    [assignment's networking rules and handler limit](labs/lab1/AGENTS.md).
  - Progress: the initial modularization, routing, and middleware design is
    [implemented](labs/lab1/docs/DESIGN.md). Workload and measurable performance targets
    remain open; this increment makes no throughput-improvement claim.

- [ ] **BL-002: Incorporate the official lab instructions when supplied**
  - Why: implementation choices and learning objectives should follow the actual
    assignment requirements.
  - Done when: supplied instructions are stored beside their labs, linked from
    the knowledge index, and reflected in relevant agent guidance.
  - Progress: Lab 1's [instructions](labs/lab1/docs/lab1_instructions.pdf) are linked
    from the knowledge index and summarized in its [agent guide](labs/lab1/AGENTS.md).
  - Remaining dependency: instructions for Labs 2–4 have not been supplied.

- [ ] **BL-008: Add opt-in HTTP connection reuse**
  - Why: the group proposes endpoint-specific reuse with closure as the default.
  - Done when: the agreed option supports sequential requests on one connection,
    correct body/framing/close semantics, bounded idle waits, and the handler cap.
  - Dependency: the server extraction and a defined connection-lifecycle contract;
    see the [proposal](labs/lab1/docs/DESIGN.md).
  - Status: explicitly reserved for a later increment; the current server closes
    each connection after one response.

## Ready

- [ ] **BL-007: Bound and measure slow-client impact**
  - Why: idle or partially sending clients can occupy all request-handler slots.
  - Done when: configurable read/write deadlines and header/body limits are
    checked with idle/partial requests, slot recovery is demonstrated, and the
    10-handler limit is verified. Document the workload and timeout choices.
  - Design: [concurrency discussion](labs/lab1/docs/DESIGN.md). A custom readiness
    scheduler remains an exploratory alternative requiring evidence.
  - Status: direction approved, implementation explicitly delayed by the group.

## In progress

No items yet.

## Done

- [x] **BL-003: Establish the initial collaboration framework**
  - Outcome: added [agent guidance](AGENTS.md), a [knowledge index](knowledge/README.md),
    this backlog, and a [decision record](knowledge/decisions/0001-collaboration-framework.md).

- [x] **BL-004: Align Lab 1 HTTP responses with the supplied instructions**
  - Outcome: malformed HTTP receives 400, `.jpg` and `.jpeg` use `image/jpeg`,
    missing GET files receive 404, and other file failures receive 400. Validation
    rejects unsupported extensions before opening files.
  - Evidence: [transport tests](httpserver/server_test.go) and
    [file-server tests](labs/lab1/src/files_test.go), including truncated POST input.

- [x] **BL-005: Extract the Lab 1 server package and ordered endpoint registry**
  - Outcome: [server package](httpserver) supports whole-path regexes,
    registration/removal by unique pattern, runtime registry updates, and a
    separate replaceable fallback. TCP networking and the 10-handler cap remain.
  - Evidence: [routing tests](httpserver/router_test.go) and
    [transport/concurrency tests](httpserver/server_test.go).

- [x] **BL-006: Separate keyed locking, file policy, and HTTP middleware**
  - Outcome: [LockMap](concurrent/locks/lockmap.go) replaces the file-specific
    locker. The file manager has no filename/depth policy. Middleware applies
    those rules, runs in registration order, and supports path-pattern scoping.
  - Decision: [accepted validation boundary](knowledge/decisions/0002-lab1-server-and-locking.md).
  - Evidence: race-enabled tests for routing, middleware, keyed locks, file-manager
    atomicity, and concurrent POST/GET operations.

- [x] **BL-009: Store the configured server concurrency limit**
  - Outcome: shared-server construction stores a positive limit, and Lab 1 handles
    construction errors and configures ten handlers. The zero-count startup
    regression is fixed.
  - Evidence: [transport tests](httpserver/server_test.go) exercise active-handler
    limits of two and ten, slot recovery, and zero-limit rejection under race detection.

- [x] **BL-010: Separate shared HTTP transport from Lab 1 application policy**
  - Outcome: moved the reusable server into `httpserver/` and application code into
    `interfaces/fileapi/`, `middleware/`, and `filepolicy/`. Generic concurrency
    code contains no application validation rules. Updated modules, workspace,
    tests, source links, and the learning map.
  - Decision: [accepted package boundaries](knowledge/decisions/0003-shared-server-and-lab1-packages.md).
  - Evidence: race-enabled transport and Lab 1 integration checks cover routing,
    middleware, file policy, POST/GET, and atomicity; dependency inspection checks
    that shared modules do not import Lab 1.

- [x] **BL-011: Separate Lab 1 source and documentation**
  - Outcome: placed code/tests under `src/` and assignment/design/learning
    documents under `docs/`. Updated import paths, links, build instructions,
    learning-map references, and the local debugger target. The module and
    agent guide remain at the lab root alongside runtime data.
  - Decision: [accepted source/docs layout](knowledge/decisions/0004-lab1-source-and-docs.md).
  - Evidence: race-enabled module checks, `go vet`, executable build, and local
    link/navigation validation.

- [x] **BL-012: Implement pub-sub service and HTTP adapters**
  - Outcome: added private broker/topic/subscriber/event state, bounded subscriber
    queues, ordered nonblocking fanout, long polling, unsubscribe wakeups, and the
    four HTTP routes on the file server's port. Waiting requests temporarily
    release admission only around their blocking channel wait.
  - Decision: [pub-sub and admission](knowledge/decisions/0005-pubsub-and-waiting-admission.md).
  - Evidence: race-enabled service, adapter, application-routing, admission-capacity,
    and existing file-server tests. Updated client examples and learning material.
