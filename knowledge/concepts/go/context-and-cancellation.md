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

Derived contexts inherit cancellation and values from their parents. Deriving a
timeout context does not start your function in a new goroutine or make all its
operations cancellable: the function must still observe Done where it can wait.
Context values can carry request metadata, but resource ownership is often easier
to follow as an explicit function argument. Cancellation and admission are
separate responsibilities; see [waiting and admission](../concurrency/waiting-goroutines-and-admission.md).

A context attached to a raw parsed HTTP request does not automatically observe a
client disconnect. That integration must be implemented by the transport or
bounded with an application timeout. See [Go's context package](https://pkg.go.dev/context).
