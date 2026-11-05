# Lab 1 Doumentation

The purpose of this document is to take notes of all knowledge that was required to solve the lab.

## Table of Contents

- [Command-Line Arguments in Go](command-line-arguments-in-go)
- [Strigs-Integer Convertion](#strigs-integer-convertion)

## Command-Line Arguments in Go

In go the `main` function does not accept any arguments. Therefore, input parameters are handle via the `os` package, similar to how it is handled in Pyhton. We can access the command-line arguments via the slice `os.Args`. The first argument always carries the name of the executable. For instance running the executable `server` located in the current directory would yeild:

```go
fmt.Printf("Program name: \n", os.Args[0])
// Program name: ./server
```

It is always useful to check the number of arguments supplied to the program. In the case of running `./server a b` we get:

```go
fmt.Printf("Number of arguments: %d", len(os.Args))
// Number of arguments: 3
```

That is because it includes the name of the program (`Args[0]="./server"`), the first command-line argument (`Args[1]="a"`), and the second command-line argument (`Args[2]="b"`).

This is very useful for us because now we can specify the port for our server without hard-coding it in the source code.

## Strigs-Integer Convertion

There is a useful library called `strconv` for converting a string to an integer. Here is how it used:

```go
var i int
var err error
i, err = strconv.Atoi("123")
```

This is going to either return an integer or return an error depending on the input. If we wish to convert an integer to a string we could use `Sprintf` as demonstrated below:

```go
var s string = fmt.Sprintf("%v", 123)
```
