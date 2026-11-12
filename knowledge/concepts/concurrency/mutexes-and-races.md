# Mutexes, races, and atomic counters

Parent: [Concurrency](concurrency.md).

A **data race** occurs when goroutines access the same memory without the needed
synchronization and at least one access writes it. `count++` is not a safe shared
counter by itself: it involves reading and updating a value.

A **mutex** allows one goroutine at a time into a protected operation, called a
critical section:

```go
var mu sync.Mutex
var count int

increment := func() {
	mu.Lock()
	defer mu.Unlock()
	count++
}
```

Every conflicting access must follow the same locking rule, including reads.
The mutex does not automatically protect a variable because it is stored nearby.

## Read/write locks and deadlocks

`sync.RWMutex` permits shared readers with `RLock`/`RUnlock`, or an exclusive
writer with `Lock`/`Unlock`. Readers must actually refrain from modifying protected
state. A read/write lock is not automatically faster for every workload.
`sync.Once.Do` is another coordination tool: it runs one function at most once,
even when multiple goroutines call it, for example to close a shared signal once.

A **deadlock** is a wait that cannot resolve, such as two goroutines each holding
a lock the other needs. Keep a consistent lock order and avoid holding a registry
lock while calling arbitrary callbacks or waiting on a resource-specific lock.
Mutex-containing values must not be copied after first use. See
[Go's synchronization package](https://pkg.go.dev/sync).

## Snapshots and atomic operations

A registry can select a handler and copy relevant configuration under a lock,
then release it before executing user code. That selection is a **snapshot**;
future updates need not change work already selected. Copy references/values
needed for dispatch rather than copying the mutex itself.

`atomic.Int32` provides synchronized operations such as `Add`, `Load`, and
`CompareAndSwap`. Compare-and-swap updates a value only if it still equals an
expected old value. It is useful for counters and maximum-tracking loops, but
separate atomic updates do not make a whole multi-step algorithm atomic.
See [sync/atomic](https://pkg.go.dev/sync/atomic) and the
[race detector](https://go.dev/doc/articles/race_detector) for verification.
