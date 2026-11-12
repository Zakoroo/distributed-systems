# Distributed systems assignments

Coursework for a student group working through four distributed-systems labs.
The first three labs are mandatory. AI assistance is permitted provided the group
understands and can explain the implementation.

## Repository layout

| Path | Purpose |
| --- | --- |
| [labs/lab1](labs/lab1/README.md) | Basic HTTP file server; build, run, and curl examples. |
| `labs/lab2/`, `labs/lab3/`, `labs/lab4/` | Placeholders for later assignments. |
| [httpserver](httpserver/) | Reusable HTTP transport, routing, and middleware composition over TCP. |
| [concurrent](concurrent/) | Reusable file-access, keyed-lock, and semaphore utilities. |
| [knowledge](knowledge/README.md) | Beginner concept notes, diagrams, lab learning maps, and design decisions. |
| [BACKLOG.md](BACKLOG.md) | Ideas and work shared by the group and agents. |

## Setup

Install a Go toolchain compatible with the module files, currently **Go 1.27.3 or
newer**. The client examples also require **curl**. Check your installations with
`go version` and `curl --version`.

From the repository root, create or update the local Go workspace:

```sh
if [ ! -f go.work ]; then
    go work init ./concurrent ./httpserver ./labs/lab1
else
    go work use ./concurrent ./httpserver ./labs/lab1
fi
```

`go.work` is intentionally ignored by Git, so each checkout needs this setup.
It connects the local modules; run it before building or testing. An existing
workspace keeps its other module entries.

## Run Lab 1

After workspace setup, run these commands from the repository root:

```sh
cd labs/lab1
mkdir -p public
go build -o http_server ./src
./http_server 8080
```

Leave that terminal running. See the [Lab 1 README](labs/lab1/README.md) for curl
commands to create, read, and append files, plus examples of error responses.
Use **Ctrl+C** in the server terminal to stop it.

## Check the implementation

Run these commands from the **repository root**, with the workspace configured:

```sh
go test -race ./httpserver/... ./labs/lab1/... ./concurrent/...
go vet ./httpserver/... ./labs/lab1/... ./concurrent/...
```

The repository has multiple modules and no root `go.mod`, so the explicit paths
select the current modules. Format changed Go files with `gofmt`.

## Learn and contribute

Start with the [knowledge index](knowledge/README.md) or the
[Lab 1 learning map](labs/lab1/docs/LEARNING.md). Open the repository root as an
Obsidian vault if you prefer that interface; the notes also work as plain Markdown.

Capture ideas in the [backlog](BACKLOG.md). Agents should read the root
[AGENTS.md](AGENTS.md) and the relevant lab's guide before changing code. Keep
reusable components independent of lab-specific policy, and update the READMEs
when build steps or client-visible behavior change.
