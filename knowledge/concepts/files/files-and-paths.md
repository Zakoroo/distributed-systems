# Files, paths, and validation

Parent: [Files and I/O](files-and-io.md).

A **filename** identifies a directory entry. A **path** describes how to reach it.
`notes.txt` is relative to the current working directory; `/tmp/notes.txt` is an
absolute path on Unix-like systems. A URL path and a filesystem path serve
different purposes even when they look similar.

## Opening does not always mean replacing

`os.OpenFile` combines access flags and permissions:

| Flag | Purpose |
| --- | --- |
| `os.O_RDONLY` | Open for reading. |
| `os.O_WRONLY` | Open for writing. |
| `os.O_CREATE` | Create if absent. |
| `os.O_TRUNC` | Truncate existing contents when opening. |
| `os.O_APPEND` | Append writes to the existing content. |

Flags are combined with `|`, which is bitwise OR in Go code, not regex alternation.
`0o644` is an octal permission value: owner read/write, group and others read.
Creation permissions are also affected by OS policy such as the process umask;
the mode argument does not reset permissions of an existing file.
See [os.OpenFile](https://pkg.go.dev/os#OpenFile).

## Separate three questions

1. **Normalization:** what consistent representation should this path use?
2. **Validation:** does it meet the application's naming or structural rules?
3. **Access policy:** is this caller allowed to access this resource?

`filepath.Join` combines components and cleans lexical `.`/`..` segments;
`filepath.Abs` resolves a path relative to the working directory. Neither proves
that an untrusted path is allowed inside a chosen directory. If validation needs
to detect attempted parent traversal, cleaning it away first hides that evidence.
See [Go's filepath package](https://pkg.go.dev/path/filepath).

**Symlinks** are filesystem references to other paths; different path strings
can refer to the same file. Lexical cleanup does not resolve all such aliases.
Filename allowlists, reserved-name rules, and depth counting are application
policies, not universal requirements for file I/O.

Concurrent file access also needs a defined resource identity and a lock covering
the complete operation; see
[keyed locks and atomicity](../concurrency/keyed-locks-and-atomic-operations.md).
