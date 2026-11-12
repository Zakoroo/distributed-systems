# 0003: Share HTTP transport and separate Lab 1 policy from adapters

Parent: [Design decisions](decisions.md).

Status: Accepted

Date: 2026-11-12

Source and document locations are subsequently updated by
[decision 0004](0004-lab1-source-and-docs.md); the responsibility boundaries below
remain in effect.

## Context

The group approved moving reusable HTTP transport out of Lab 1 and completing the
separation of file-validation policy from generic concurrency utilities. It chose
`interfaces` for application adapters and a separate middleware directory.

## Decision

- Root `httpserver/` is a shared module for raw TCP transport, HTTP request parsing,
  ordered routing, middleware composition, and configurable handler admission.
  It depends on generic concurrency utilities, never a lab package.
- Lab 1's `interfaces/fileapi/` handles file GET/POST operations. Its `middleware/`
  package validates requests, and `filepolicy/` owns pure filename/depth rules and
  content-type mappings. Both adapters and middleware import policy; neither
  imports the other. All three share Lab 1's existing `go.mod`.
- Generic `concurrent/files` provides synchronized file access. Its lexical path
  normalization remains part of resource-key consistency, while naming/depth
  validation definitions and checks belong to the application.
- `main.go` assembles these components and supplies the assignment's limit of ten.
  Shared-server construction validates and stores its configurable positive limit.
- `interfaces` is an ordinary directory name for HTTP-facing adapters. The group
  deliberately chose it over Go's import-restricted `internal` directory. Maintain
  the intended boundary through dependency direction and review.
- A future `pubsubapi` would sit beside `fileapi`. No pub-sub implementation,
  connection reuse, or slow-client redesign is included in this increment.

## Alternatives considered

Keeping validation helpers in the shared file module leaves application policy
in the synchronization layer even if only middleware calls them. Keeping all
application code in `main` obscures its separate responsibilities. A shared
`filepolicy` package avoids making middleware depend on file handlers merely to
reuse the MIME mapping.

## Consequences

The repository now has three workspace modules: `concurrent`, `httpserver`, and
`labs/lab1`. Workspaces created before this change need `go work use ./httpserver`.
The shared components can be reused across labs without importing Lab 1's policy.
Direct file API or manager callers still own input-validation policy.

This updates the package placement in [decision 0002](0002-lab1-server-and-locking.md).
See the [Lab 1 design](../../labs/lab1/docs/DESIGN.md) and
[learning map](../../labs/lab1/docs/LEARNING.md) for current APIs and source links.
