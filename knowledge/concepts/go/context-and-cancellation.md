# Context, deadlines, and cooperative cancellation

Parent: [Go fundamentals](go-fundamentals.md).

A Go `context.Context` carries a cancellation signal, an optional deadline, and
request-scoped values through function calls. It is commonly passed as the first
argument rather than stored globally.

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

select {
case value := <-messages:
    // Use the received value.
    _ = value
case <-ctx.Done():
    // Stop waiting: ctx.Err() explains cancellation or timeout.
}
```

This fragment assumes a messages channel exists. A timeout or `cancel()` closes
Done; it does not forcibly terminate goroutines. Code must observe the signal.
`context.Canceled` and `context.DeadlineExceeded` distinguish the two outcomes.
Call the returned cancel function to release timeout resources promptly.

Derived contexts inherit cancellation and values from their parents. A private
key type can carry a request-scoped capability without exposing mutable state.
For example, a caller may supply a callback that accounts for a blocking wait;
the waiting algorithm need not know about HTTP or semaphore implementations.
Those callbacks still need explicit ownership and synchronous-use contracts.

A context attached to a raw parsed HTTP request does not automatically observe a
client disconnect. That integration must be implemented by the transport or
bounded with an application timeout. See [Go's context package](https://pkg.go.dev/context).
