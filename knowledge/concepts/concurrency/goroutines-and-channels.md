# Goroutines and channels

Parent: [Concurrency](concurrency.md).

**Concurrency** means multiple tasks can be in progress during overlapping
periods. **Parallelism** means tasks actually execute at the same time. You can
have concurrency even on one CPU core.

A **goroutine** is work scheduled by Go's runtime. `go work()` starts the call in
a goroutine and lets the caller continue; it does not wait for completion.
Multiple goroutines can share memory, which requires coordination. Starting more
goroutines does not automatically increase useful throughput.

## Channels communicate values

```go
done := make(chan string)
go func() {
	done <- "finished"
}()
message := <-done
```

This fragment sends a message from one goroutine and receives it in another.
An unbuffered send waits for a corresponding receiver. `make(chan string, 2)`
creates a buffer that can hold two values; a send blocks when that buffer is full.

`close(ch)` tells receivers no more values will be sent. It is not a requirement
after every send, and sending on a closed channel panics. A channel is not closed
merely because a goroutine stops using it.

A receive can return a second value telling us whether it received a real value:

```go
value, open := <-messages
if !open {
    // The channel is closed and its buffer is empty; value is the zero value.
    return
}
use(value)
```

This fragment assumes `messages` and `use` exist. A closed buffered channel still
delivers values already queued before returning `open == false`. Closing safely
requires ownership: coordinate every sender with the operation that closes the
channel, because a later send would panic. See
[Go's close rules](https://go.dev/ref/spec#Close).

## Wait for events deliberately

`select` waits on channel communication cases. A `default` case makes it return
immediately if none can proceed. This can implement a nonblocking send, but it is
not a general “which socket is readable?” operation: `net.Conn` is not a channel.
`time.After(duration)` supplies a channel that becomes ready after a delay, useful
as a timeout case. Selecting the timeout does not cancel other goroutines by itself.

When a channel wait blocks, the goroutine still exists with its call stack; it
does not continuously check the channel. The runtime can park it and run other
goroutines on available OS threads. This does not automatically release any mutex
or semaphore permit the goroutine holds. See
[waiting goroutines and admission limits](waiting-goroutines-and-admission.md)
for a step-by-step example and resource accounting.

A `sync.WaitGroup` tracks work whose completion must be awaited. Tests often use
channels for specific milestones and a WaitGroup for overall completion. Avoid
assuming a goroutine has finished just because some time has passed.

See the concurrency lessons in the [Tour of Go](https://go.dev/tour/list) and
[Go's select rules](https://go.dev/ref/spec#Select_statements). Continue with
[shared-state synchronization](mutexes-and-races.md).
