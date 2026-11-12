# Repository guide for agents

## Purpose and learning goals

This repository contains university distributed-systems assignments completed by
a student group. There are four labs; the first three are mandatory. AI assistance
is permitted provided the students understand the resulting work.

- Prefer clear, explainable implementations and incremental changes.
- Explain significant design choices, algorithms, concurrency assumptions, and
  tradeoffs so the group can reason about and maintain the code.
- When the group identifies an unfamiliar concept, add or update a short note in
  `knowledge/concepts/` and link it from the relevant indexes and design notes.
  Explain the basics in plain language with small, standalone examples; avoid
  unnecessary theory or assuming prior familiarity. Keep the main explanation
  independent of the current implementation and link project usage separately.
- Add simple Mermaid diagrams when they clarify interaction order, waiting, or
  decisions. Keep them focused and pair them with a short explanation; concept
  diagrams should remain independent of implementation details.
- Preserve the learning objectives of each assignment. Check the lab requirements
  before substituting a library or framework for a mechanism students must build.

## Repository map

- [Root README](README.md): collaborator setup and repository-wide checks.
- [Lab 1 README](labs/lab1/README.md): build/run instructions and curl examples.
- `labs/lab1/`: a Go HTTP server implemented over TCP, with bounded concurrent
  connection handling and GET/POST file operations. Code lives in `src/` and lab
  documents in `docs/`. Its `src/interfaces/fileapi/` package handles file requests;
  `src/middleware/` applies rules defined in `src/filepolicy/`. Future `pubsubapi`
  work remains deferred. Build the executable from `labs/lab1` with
  `go build -o http_server ./src`; keep that working directory when running it.
- `httpserver/`: the reusable HTTP transport/router module. It owns middleware
  composition and configurable request admission, with no dependency on any lab.
- `labs/lab2/`, `labs/lab3/`, `labs/lab4/`: currently placeholders for later labs.
- `concurrent/`: a reusable Go module containing file access utilities,
  `locks.LockMap[K]` keyed read/write locking, and semaphores. Changes here may
  affect multiple labs. File policy belongs to callers, not the generic locker
  or file manager; Lab 1 enforces it in middleware.
- Other implementation directories outside `labs/` are intended for modular,
  reusable components that can serve more than one lab.
- [Lab 1 agent guide](labs/lab1/AGENTS.md): assignment constraints and a summary
  of the [official instructions](labs/lab1/docs/lab1_instructions.pdf).
- [Lab 1 learning map](labs/lab1/docs/LEARNING.md): a study path and concept-to-code
  coverage for the basic implementation and its tests.
- [Knowledge base](knowledge/README.md): an index of learning notes and design
  decisions, usable as plain Markdown or through Obsidian.
- [Backlog](BACKLOG.md): shared ideas and selected work, editable by humans and
  agents.

Lab 1 instructions are available; instructions for Labs 2–4 are still pending.
Do not invent assignment requirements, grading criteria, or restrictions that
have not been provided. Changes to shared code used by Lab 1 must also preserve
its constraints: raw `net` networking, at most 10 active request handlers, and
atomic file updates. See the Lab 1 guide for details and the long-polling exception.

## Finding context

1. Read this guide and any applicable directory-level `AGENTS.md` files.
2. Locate the instructions for the lab being changed, when available. Treat those
   instructions as the source of truth for assignment requirements; distinguish
   them from learning notes and proposed designs.
3. Use the [knowledge index](knowledge/README.md) to locate relevant explanations
   and decisions. Read only the relevant documentation, implementation, tests,
   and shared-code consumers. Use targeted searches rather than reading every
   document.
4. If missing instructions prevent a sound implementation choice, explain what is
   unknown and ask for the specific requirement. Continue independent work where
   possible.

## Refactoring and scalability

- Keep lab-specific behavior inside its lab. Shared components should not depend
  on lab implementations.
- Extract reusable components around a concrete use case and a small, clear API.
  Consider existing callers before changing a shared API.
- Preserve observable behavior during refactors. Identify intentional behavior
  changes separately from structural cleanup.
- Tie scalability changes to a concrete workload, resource limit, or bottleneck.
  State the expected benefit and verify it with appropriate evidence; label
  unmeasured improvements as hypotheses.
- For concurrent code, make ownership, synchronization, lock ordering, and
  resource lifetimes understandable. Consider blocking operations, bounded
  goroutines/memory, error paths, and cleanup when relevant to the change.
- Preserve unrelated work already present in the working tree.

## Go development and verification

The current modules are `concurrent/`, `httpserver/`, and `labs/lab1/`. Use the Go
version declared in their `go.mod` files. The root `go.work` is ignored by Git;
a local copy may already exist. If absent, create one from the repository root with:

```sh
go work init ./concurrent ./httpserver ./labs/lab1
```

Preserve an existing workspace configuration. If it predates the server
extraction, add the shared module with `go work use ./httpserver`. Add future lab
modules when they actually contain a `go.mod` file.

- Format changed Go files with `gofmt`.
- Run checks for affected modules and their consumers. For the current modules,
  run these commands from the repository root with the workspace configured:

  ```sh
  go test ./httpserver/... ./labs/lab1/... ./concurrent/...
  go vet ./httpserver/... ./labs/lab1/... ./concurrent/...
  ```

- Use race detection and focused concurrency checks when changing synchronization
  or concurrent behavior. Use benchmarks or load checks to support performance
  claims.
- Report what was checked, the outcome, and any validation limitations. A package
  with no test files has not had its behavior covered by `go test` alone.

## Knowledge and planning

The group uses Markdown files tracked in Git for knowledge and planning. See the
[adopted collaboration approach](knowledge/decisions/0001-collaboration-framework.md).

- Keep this guide short: orientation and working rules belong here; detailed
  explanations, experiments, and design decisions belong in linked documents.
- Distinguish assignment requirements, adopted decisions, and exploratory ideas.
- Use the knowledge index to find relevant notes and update affected notes when
  an implementation change makes them inaccurate. Keep navigation links current.
- Keep primary navigation tree-shaped: knowledge index, subject index, then
  concept note. Give each concept one primary parent. Use cross-links only for
  relevant prerequisites or relationships; keep one explanation per concept.
- Maintain each lab's learning map when implementation changes introduce or remove
  concepts. Distinguish intended architecture, current behavior, and deferred work.
- Keep root and lab READMEs current when setup, build/run commands, or documented
  client-visible behavior changes.
- Treat backlog ideas as proposals, not instructions to implement them. Work on
  items selected by the user. When working on an existing backlog item, preserve
  its ID and update its status and completion evidence. Direct user requests do
  not need a backlog entry before work can begin.
