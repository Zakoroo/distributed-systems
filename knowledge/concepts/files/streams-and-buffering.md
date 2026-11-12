# Streams and buffering

Parent: [Files and I/O](files-and-io.md).

A **stream** supplies a sequence of bytes over time. A string already in memory,
a file, and a connection can all supply bytes, even though their storage and
waiting behavior differ.

In Go, `io.Reader` describes reading into a byte slice; `io.Writer` describes
writing one. These are [interfaces](../go/functions-and-types.md), so code can
work with a capability rather than a specific data source.

## Reads can be partial

`Read` returns a byte count and an error. It is allowed to supply fewer bytes than
requested. Process returned bytes before acting on an accompanying error.
`io.EOF` indicates no more input; `io.ErrUnexpectedEOF` indicates data ended
before an expected structure was complete.

`io.ReadAll(reader)` collects the remaining stream until EOF or an error. It can
wait and can allocate a large amount of memory; the caller must decide what
input sizes and waits are acceptable. `io.Copy` moves data from reader to writer
without requiring the caller to collect the complete content first.
See [Go's io package](https://pkg.go.dev/io).

## A buffer keeps bytes already obtained

`bufio.NewReader(source)` adds a temporary in-memory buffer. A parser may consume
a line while the buffer retains bytes read ahead for the next line or message.
Discarding that buffer and reading directly from the source can lose access to
those already-buffered bytes. Keep the same reader when continuing a stream.
See [Go's bufio package](https://pkg.go.dev/bufio).

`bytes.NewReader(data)` reads from existing bytes in memory. `io.NopCloser(reader)`
adds a `Close` method that does nothing, useful where an API requires an
`io.ReadCloser`. It does not close resources owned by the wrapped reader.
`io.Discard` accepts writes and throws the bytes away, useful for silencing a
test's logger without changing the logging code.

Closing a file or connection releases its resource. Some higher-level body
objects may drain unread bytes when closed, so cleanup order matters; define
[ownership](../go/errors-and-cleanup.md) rather than assuming every Close is identical.
