# Errors, cleanup, and ownership

Parent: [Go fundamentals](go-fundamentals.md).

An operation that can fail commonly returns both a value and an `error`:

```go
file, err := os.Open("notes.txt")
if err != nil {
	return err
}
defer file.Close()
```

`nil` means no error here. Check the error before using the returned value.
Returning an error lets a caller decide how to respond; logging only records it.
`log.Fatal` records a message and exits the program, so deferred cleanup does not
run on that exit path. It usually belongs at the outer application boundary.

## Add context without losing the original error

`fmt.Errorf("open notes: %w", err)` wraps an error. `errors.Is(err, os.ErrNotExist)`
can recognize a missing-file error even through wrapping. Comparing error strings
is fragile: wording is not a reliable error category.

An error is a Go value; an HTTP status is a protocol response. The application
must deliberately translate between them. A generic file error does not by itself
mean HTTP 404. See [HTTP basics](../networking/http-basics.md).

## `defer` schedules cleanup

A deferred call runs when the surrounding function returns, including early
returns. Multiple deferred calls run in reverse order. Arguments to the call are
evaluated when it is deferred, not later when it runs.

```go
defer fmt.Println("first scheduled")
defer fmt.Println("second scheduled")
// On return: second scheduled, then first scheduled.
```

**Ownership** means deciding who must close a file, connection, or body and when.
It is a program contract, not a Go keyword. Closing too early prevents another
component from using the resource; never closing can leak resources.

A **panic** unwinds the goroutine's call stack and runs deferred calls. Recovery
is possible in deferred code, but ordinary expected failures should use error
returns. Cleanup that runs during a panic is valuable even without recovery.
See [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover).
