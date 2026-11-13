# Concurrency

Parent: [Concepts](../README.md).

1. [Goroutines and channels](goroutines-and-channels.md): overlapping tasks,
   blocking, messages, `select`, and waiting for completion.
2. [Mutexes, races, and atomics](mutexes-and-races.md): shared state, read/write
   locks, deadlocks, snapshots, and atomic counters.
3. [Semaphores and condition variables](semaphores-and-conditions.md): limiting
   active work and waiting until a shared condition changes.
4. [Keyed locks and atomic operations](keyed-locks-and-atomic-operations.md): one
   lock per resource, safe cleanup, complete versions, and backpressure.

5. [Publish-subscribe and long polling](publish-subscribe.md): broadcast, bounded
   queues, ordering, cancellation, and waiting without blocking publishers.

Read [functions and types](../go/functions-and-types.md) first if callbacks and
closures are unfamiliar. [Testing](../go/testing.md) explains how to gather
evidence about concurrent behavior without relying on arbitrary sleeps.
