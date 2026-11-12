# 0004: Separate Lab 1 source and documentation

Parent: [Design decisions](decisions.md).

Status: Accepted

Date: 2026-11-12

## Context

The group approved separating source and documents to make Lab 1 easier to
navigate while preserving the package boundaries from
[decision 0003](0003-shared-server-and-lab1-packages.md).

## Decision

- Move the executable, tests, `interfaces`, `middleware`, and `filepolicy` under
  `labs/lab1/src/`.
- Move assignment instructions, design notes, and the learning map under
  `labs/lab1/docs/`.
- Keep `go.mod` and `AGENTS.md` at the lab root. The module path and workspace
  entry stay the same; sub-package import paths gain `/src/`.
- Keep runtime `public/` at the lab root. Build and run from `labs/lab1` with
  `go build -o http_server ./src` and `./http_server 8080`.
- Update relative documentation links and the local debugger's source target.

## Alternatives considered

Keeping the previous flat layout requires fewer path changes, but mixes source,
assignment material, and learning notes. Moving the module file into `src` would
also change workspace configuration; the group selected a stable lab-root module
boundary instead.

## Consequences

The change is organizational. The server's protocol behavior, application policy,
ten-handler limit, and working-directory-relative runtime data keep their meaning.
Commands targeting the executable now select `src`; recursive module test
commands still cover the lab. Agent instructions remain discoverable above both
source and documentation.

See the [Lab 1 guide](../../labs/lab1/AGENTS.md),
[design](../../labs/lab1/docs/DESIGN.md), and
[learning map](../../labs/lab1/docs/LEARNING.md) for current navigation and commands.
