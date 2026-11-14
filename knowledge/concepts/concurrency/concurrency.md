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

## Optional background for later work

For the small Lab 1 bonus, read the pub-sub note below. The admission-release
and job-scheduler notes are optional alternatives, not implemented machinery.

5. [Waiting goroutines and admission limits](waiting-goroutines-and-admission.md):
   what remains allocated during a wait, how channel wakeups work, and why a
   released permit does not remove a goroutine.
6. [Publish-subscribe and long polling](publish-subscribe.md): broadcast, bounded
   queues, ordering, cancellation, and waiting without blocking publishers.
7. [Job queues, workers, and deadlines](job-scheduling.md): ready versus waiting
   work, deadline ordering, reserved worker groups, and scheduler tradeoffs. Exploratory;
   no job scheduler is implemented.

Read [functions and types](../go/functions-and-types.md) first if callbacks and
closures are unfamiliar. [Testing](../go/testing.md) explains how to gather
evidence about concurrent behavior without relying on arbitrary sleeps.
