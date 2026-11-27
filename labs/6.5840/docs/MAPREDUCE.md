# 6.5840 MapReduce design

The coordinator schedules one map task per input file and one reduce task per partition. Workers and the coordinator run in the same directory and exchange task assignments and completions over a Unix socket. Assignments carry filenames, and the records travel through the shared filesystem. [MIT lab write-up](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)

```mermaid
flowchart LR
    M[Maps write one JSON file per partition] --> B[All map submissions accepted]
    B --> R[Reducers read and sort their partitions]
    R --> P[Coordinator publishes each reduce output]
    P --> D[All outputs published, workers exit]
```

A mapper hashes each key to choose its partition. Each reducer gets one published file from every map for its partition. It loads those records, sorts by key, and calls the plugin once per group of equal keys. This keeps the sort local to each reducer and lets reducers run in parallel. Each reducer holds its partition's records in memory while it sorts them.

## Retries and publication

- The coordinator derives progress by scanning the task records under one mutex. It serves map tasks while any map is unfinished, then reduce tasks while any reduce is unfinished, then tells workers to exit. The scans keep scheduling simple for the lab's workload.
- Each assignment names a fresh staging directory. The worker closes its output and returns the assignment unchanged. After ten seconds, another poll can reassign the task with a new directory. The original submission can still succeed until reassignment. When no unfinished task is available, the coordinator returns `wait`.
- Under the same lock, the coordinator checks that the submission matches the current assignment, publishes the output with a rename, and marks the task done. A map publishes its whole directory to a stable path. A reduce publishes one file to `mr-out-N`. This keeps intermediate filenames independent of retry history and makes a failed rename leave the task unfinished.
- Each successful reduce publication removes that attempt's staging directory. The last reduce removes the whole job directory, including published map partitions and any failed attempts. The coordinator owns cleanup, so workers can handle an interrupted submission without deciding whether their files are still needed.

The coordinator owns scheduling and publication. The worker runs plugin functions and writes attempt files. The coordinator keeps state in memory, so restarting it starts a new job. [Coordinator source](../src/mr/coordinator.go), [Worker source](../src/mr/worker.go)

## Progress and logs

The coordinator's existing `Done()` poll refreshes a bar on stderr with published map and reduce counts. A retry leaves the completed count unchanged. Startup, deadline reassignments, and staging, publication, or cleanup errors have timestamped log lines. Both progress and event logging write after releasing the coordinator lock.

In a terminal the reporter redraws one line and keeps log entries above it. When stderr is redirected, progress uses plain lines and skips repeated snapshots. The reporter only needs an output writer and progress counts, so other exercises can reuse it. It lives inside the course module so the standalone lab archive includes it. [Reporter source](../src/observability/reporter.go)

## Verification

From the repository root, run the bundled suite for correctness, parallelism, job count, early exit, and worker crash recovery. [Bundled test script](../src/main/test-mr.sh)

```sh
cd labs/6.5840/src/main
GOWORK=off RACE=-race bash ./test-mr.sh
```

The focused tests check that a stale reduce attempt cannot publish and that a failed map publication keeps reducers waiting. [Coordinator tests](../src/mr/coordinator_test.go)
