# Packages, modules, workspaces, and the command line

Parent: [Go fundamentals](go-fundamentals.md).

A **package** groups Go files into one compilation unit, usually in one directory.
Its files share package-level declarations. Imports refer to other packages, not
individual source files. An executable starts at `func main()` in `package main`.

A **module** groups related packages. Its `go.mod` declares its import-path prefix,
Go language version, and module dependencies. A sub-package normally does not
need its own `go.mod`.

A folder named `src` is an organizational choice, not a special Go requirement.
With `go.mod` above it, a package at `src/widgets` has import path
`<module-path>/src/widgets`. The build target follows the source location; the
module boundary stays where `go.mod` lives.

A **workspace** uses `go.work` to connect local modules. It lets one local module
use another module's current source without publishing a dependency version.
See the [go.mod reference](https://go.dev/doc/modules/gomod-ref) for module fields.

## Arguments arrive as strings

Running `./program 8080` supplies:

```text
os.Args[0] = "./program"
os.Args[1] = "8080"
```

Check the argument count before indexing. `strconv.Atoi("8080")` returns an
integer and an error. Successful conversion does not establish a valid port
range: validation is a separate step. A port is an integer from 0 through 65535;
port 0 asks the OS to select a port when listening.

`fmt.Printf` formats output: `%s` prints text, `%d` a decimal integer, `%v` a value
in its usual form, and `%q` a quoted representation. `fmt.Sprintf` returns the
formatted text instead of printing it.

## Build and run from a known directory

`go build -o program .` compiles the current package into an executable named
`program`; `./program 8080` runs it. Relative filesystem paths are resolved from
the process's working directory, which is not necessarily its source directory.
`go test` runs package tests; [testing](testing.md) explains the checks.

The root [agent guide](../../../AGENTS.md) lists this repository's module/workspace
commands. The [Lab 1 learning map](../../../labs/lab1/docs/LEARNING.md) connects these
concepts to the current code. See [strconv.Atoi](https://pkg.go.dev/strconv#Atoi)
and [os.Args](https://pkg.go.dev/os#Args) for the APIs.
