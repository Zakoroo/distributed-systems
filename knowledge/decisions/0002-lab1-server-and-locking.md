# 0002: Separate Lab 1 transport, routing, middleware, and keyed locking

Parent: [Design decisions](decisions.md).

Status: Accepted

Date: 2026-11-12

The package-placement part of this record is superseded by
[decision 0003](0003-shared-server-and-lab1-packages.md). Its routing, locking,
and caller-owned validation contracts remain in effect.

## Context

The group approved extracting the basic file server from `main.go`, adding
ordered endpoint registration and a separate fallback, and moving validation out
of file-specific locking. The group explicitly chose middleware as the filename
and path-depth validation boundary, including removing these checks from the
file manager. This replaces the earlier proposal to retain policy in the manager.

## Decision

- Use the existing Lab 1 module's `server` package for raw TCP networking, HTTP
  parsing/writing, ordered routing, and middleware registration.
- Use unique whole-path regex patterns for endpoint registration/removal. Keep
  fallback separate and replaceable so later registrations still precede it.
- Permit registry updates while serving. Dispatch takes a consistent snapshot;
  handlers run without holding registry locks.
- Run middleware in registration order, selecting matching path patterns. Allow
  direct handler composition so file-only policy can wrap the file fallback
  without affecting unrelated endpoints.
- Replace `files.Locker` with `locks.LockMap[K comparable]`. Keep lock entries
  alive for both current holders and waiting callers, then reclaim idle entries.
- The file manager normalizes lexical path aliases for locking but does not
  validate filename syntax, path depth, or filesystem-access policy. Middleware
  owns the basic server's validation. Other callers supply their own policy.
- Retain the ten-handler limit and close each connection after its response.
  Keep connection reuse for another increment. Slow-client improvements are
  approved in principle but explicitly delayed.

## Alternatives considered

A catch-all endpoint registered at startup can shadow routes added later; a
separate fallback avoids that ordering dependency. Function identity is not a
suitable removal key in Go, so registration uses unique pattern strings.
Retaining validation inside the file manager was considered; the group chose
caller-owned policy to keep the reusable components independent of file rules.

## Consequences

The current implementation separates transport and application policy, and
supports path-specific middleware without changing the shared locker. Callers
using the file manager directly must validate inputs according to their needs.
Lexical normalization is not a filesystem sandbox or symlink canonicalization.

See the [design notes](../../labs/lab1/docs/DESIGN.md) for API contracts, verification,
and deferred work. No performance improvement is claimed without workload data.
