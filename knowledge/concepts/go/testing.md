# Testing and evidence

Parent: [Go fundamentals](go-fundamentals.md).

A test supplies inputs, observes results, and checks whether behavior matches an
expectation. In Go, tests normally live in `*_test.go` files:

```go
func TestDouble(t *testing.T) {
	if got := double(3); got != 6 {
		t.Fatalf("double(3) = %d, want 6", got)
	}
}
```

This snippet assumes `double` is available in that package. `t.Fatal`/`Fatalf`
stop the current test goroutine; `t.Error`/`Errorf` record a failure and continue.
Table-driven tests run several input/expectation pairs through the same check.
`t.Run` names a subtest, `t.Helper` marks a helper so failures point to its caller,
and `t.Cleanup` registers cleanup for when the test finishes.

## Make concurrency checks observable

Sleeping and hoping another goroutine has finished is unreliable. Use a channel
or a `sync.WaitGroup` to coordinate. `WaitGroup.Go` starts tracked work; `Wait`
waits for that work to finish. Add bounded waits when a deadlock is a possible
failure, so a broken implementation does not hang the test indefinitely.
`runtime.Gosched` yields execution to allow other goroutines to run; it does not
guarantee their order or replace synchronization.

`t.TempDir` supplies an isolated directory. A **test double** replaces a dependency
with controlled behavior: an injected listener or `net.Pipe` can exercise protocol
code without a real listening TCP socket. Such checks do not prove behavior
across real machines or under real network delays.

## Different commands give different evidence

| Command | What it tells you |
| --- | --- |
| `go test` | Whether the selected tests passed; it also compiles their packages. |
| `go test -race` | Whether executed code triggered detectable data races. |
| `go vet` | Whether its static checks found suspicious code. |
| `go build` | Whether the selected program compiles. |

A passing race run does not prove all possible interleavings are safe. A package
with no tests has not had its behavior checked merely because `go test` succeeded.
Throughput claims need a defined workload and measurement, not only unit tests.

See [Go's testing package](https://pkg.go.dev/testing) and
[race detector guide](https://go.dev/doc/articles/race_detector).
