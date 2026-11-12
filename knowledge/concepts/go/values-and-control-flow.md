# Values, types, and control flow in Go

Parent: [Go fundamentals](go-fundamentals.md).

A **value** is data, such as `10` or `"hello"`. Its **type** describes what kind of
data it is and which operations are allowed.

## Read the declarations

```go
var count int      // Starts at zero.
name := "Ada"      // Declare a variable inside a function; infer its type.
const limit = 10   // A constant cannot be reassigned.
```

`string` holds text, `bool` holds `true`/`false`, and `int` holds signed integers.
`uint` is unsigned: it cannot represent a negative number. A `byte` is an unsigned
8-bit integer. `[]byte` is a slice of bytes, often used for file or network data.

Go provides zero values: numeric zero, `false`, `""`, or `nil` for types such as
pointers, slices, and maps. A nil map must be initialized before assigning entries.

## Collections and records

```go
names := []string{"Ada", "Lin"}             // Slice: an ordered sequence.
names = append(names, "Sam")                // May allocate new storage.
scores := map[string]int{"Ada": 7}         // Map: key -> value lookup.
score, exists := scores["Ada"]             // exists distinguishes missing keys.
```

`range` iterates over a collection. Slice order is meaningful; map iteration order
must not be used to decide which operation has priority. Slices can share backing
storage, so copying a slice value does not necessarily copy its elements.
Helpers such as `slices.Contains` and `slices.Delete` perform common collection
operations. Strings have helpers too, such as `strings.HasPrefix` and
`strings.ToLower`. A `rune` represents a Unicode code point; ranging over a string
visits runes, while `len(text)` counts bytes. ASCII-only validation is an explicit
policy, not a general limitation of Go strings.

A **struct** groups named fields; a **method** attaches behavior to a type:

```go
type Counter struct{ Value int }

func (c *Counter) Increment() {
	c.Value++
}
```

`*Counter` is a pointer to a Counter. `&value` obtains a pointer; `*pointer`
accesses what it points to. The pointer receiver above lets the method update
the original Counter. A nil pointer points to no value and cannot be dereferenced.
Copying a struct copies its fields, which may still refer to shared data.

## Follow execution

`if` selects a branch, `switch` selects among cases, and `for` repeats work.
`return` ends a function; `break` exits a loop; `continue` starts its next iteration.
`for range 3` repeats three times. `&&`, `||`, and `!` mean “and”, “or”, and “not”.

Capitalized names such as `Counter` are exported to other packages; lowercase
names are package-private. See the [Tour of Go](https://go.dev/tour/list) for
small exercises, and the [language specification](https://go.dev/ref/spec) for
precise rules.
