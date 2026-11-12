# Functions, interfaces, and generics

Parent: [Go fundamentals](go-fundamentals.md).

Read [values and types](values-and-control-flow.md) first if declarations and
pointers are unfamiliar.

## Functions can be values

A function takes inputs and returns results. Its signature tells you their types:

```go
func double(n int) int {
	return n * 2
}

func apply(n int, operation func(int) int) int {
	return operation(n)
}
```

`apply(3, double)` produces `6`. Passing a function for another component to call
is a **callback**. A named function type makes a repeated signature easier to read:
`type Operation func(int) int`.

An anonymous function has no declared name. It may refer to variables around it;
this is a **closure**:

```go
prefix := "Hello, "
greet := func(name string) string { return prefix + name }
```

Captured variables are shared variables, not automatically copied snapshots.
If multiple goroutines use them, consider [synchronization](../concurrency/mutexes-and-races.md).
A variadic parameter such as `operations ...Operation` accepts several arguments.
`operations...` expands a slice when passing it to another variadic function.

## Interfaces describe capabilities

An **interface** specifies methods a value must provide. A type satisfies it by
having those methods; no explicit `implements` declaration is needed.

```go
type Greeter interface {
	Greet() string
}
```

Any type with `Greet() string` can be used as a Greeter. This lets code work with
a capability instead of a particular implementation. Standard examples are
`io.Reader`, `io.Writer`, and `net.Conn`.

A struct can also **embed** a type as an unnamed field. For example,
`struct{ io.Reader }` exposes the embedded reader's `Read` method. A wrapper can
reuse that reading behavior while adding its own `Close` method.

## Generics reuse an algorithm with different types

```go
func contains[K comparable](items map[K]bool, key K) bool {
	return items[key]
}
```

`K` is a type parameter. `comparable` allows types usable for equality comparison
and map keys, such as strings and integers. A generic struct works similarly:
`LockMap[string]` selects strings as its key type. Generic means reusing an
algorithm with different types; it does not automatically make access concurrent-safe.

See [Go's generics tutorial](https://go.dev/doc/tutorial/generics) and the
[Tour of Go](https://go.dev/tour/list) for interfaces and function values.
