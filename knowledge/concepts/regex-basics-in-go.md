# Regular expressions in Go: a beginner's guide

Parent: [Go fundamentals](go/go-fundamentals.md).

A **regular expression**, usually called a **regex**, is a rule for recognizing
text. You give it a pattern and some text, then ask: “Does this text match?”
It can also help extract parts of text, such as a number or a name.

This guide covers the basics with standalone examples. You do not need to know
how any server or other application works to follow it.

## Start with ordinary text

The pattern `cat` looks for the letters `cat` together. In Go, a normal regex
search can find a match **anywhere** in the input: both `cat` and `a cat sleeps`
match. So does `caterpillar`! Requiring an exact match takes an extra step,
explained below.

Most ordinary characters match themselves. Some symbols have special meanings.

## Read special symbols one at a time

| Symbol | Meaning | Example |
| --- | --- | --- |
| `.` | Any one character except a newline, by default | `c.t` matches `cat`, `cut`, or `c9t`. |
| `*` | Zero or more repetitions of the preceding item | `ab*c` matches `ac`, `abc`, or `abbbc`. |
| `\.` | A literal dot | `notes\.txt` matches the text `notes.txt`. |
| `|` | Either alternative | `cat|dog` matches `cat` or `dog`. |

In `.*`, the dot means “any character” and the star means “repeat that zero or
more times.” Together they accept any sequence of characters without newlines,
including an empty sequence. A slash is just another character, so `.*` can also
match `folder/notes.txt`.

The dot in `.*` does **not** mean a filename extension. Regex syntax is different
from filename wildcards such as `*.txt`. To describe a dot followed by `txt`, use
`\.txt`; add `.*` before it if other characters may come first.

## Require the entire input to match

Two markers specify where a match must begin and end:

- `\A`: the beginning of the input.
- `\z`: the end of the input.

The pattern `\Acat\z` accepts exactly `cat`. It rejects both `caterpillar` and
`a cat sleeps`. These markers describe positions; they do not consume characters.

For a complete filename ending in `.txt`, use `\A.*\.txt\z`:

| Input | Matches? |
| --- | --- |
| `notes.txt` | Yes |
| `folder/notes.txt` | Yes |
| `notes.jpg` | No |
| `notes.txt.backup` | No |

This only checks text. It does not check whether a file exists.

## Capturing groups and grouping alternatives

Parentheses keep parts of a pattern together, much like parentheses in arithmetic.
There are two useful forms to recognize:

- `(cat|dog)` groups the alternatives and also makes that part available for
  extraction. This is a **capturing group**.
- `(?:cat|dog)` groups the alternatives without adding an extractable part.
  This is a **non-capturing group**.

For a yes/no check, both forms give the same matching result. Captures become
useful when asking for parts of a match through methods such as
`FindStringSubmatch`. You can leave extraction for later.

Grouping matters when combining alternatives with other rules. Compare:

```text
\Acat|dog\z       means: starts with cat OR ends with dog
\A(?:cat|dog)\z   means: the entire input is either cat OR dog
```

The first pattern accepts `caterpillar` and `hotdog`. The second accepts only
`cat` or `dog`. The group makes both boundaries apply to both alternatives;
`?:` says we only need grouping, not extraction.

## Use a pattern in Go

There are two steps: **compile** the pattern to prepare a reusable matcher, then
**match** some text against it. This is a complete program:

```go
package main

import (
	"fmt"
	"regexp"
)

func main() {
	matcher, err := regexp.Compile(`\A(?:cat|dog)\z`)
	if err != nil {
		fmt.Println("Invalid pattern:", err)
		return
	}

	fmt.Println(matcher.MatchString("cat"))         // true
	fmt.Println(matcher.MatchString("caterpillar")) // false
}
```

`Compile` reports invalid syntax as an error, for example an opening parenthesis
without a closing one. `MatchString` returns a boolean (`true` or `false`); it
does not change the input or perform an application action. The same matcher can
be reused for more strings without compiling again.

See the [Go regexp documentation](https://pkg.go.dev/regexp) for these operations.

## Keep Go strings separate from regex syntax

These Go strings contain the same pattern:

```go
`\A.*\.txt\z`      // Backticks preserve backslashes.
"\\A.*\\.txt\\z"  // Double quotes need escaped backslashes.
```

The backticks and quotes belong to Go; they are not part of the regex. Backticks
often make patterns easier to read. See [Go string literals](https://go.dev/ref/spec#String_literals).

You do not need to memorize every regex feature. Start with the few symbols used
here and consult [Go's syntax reference](https://pkg.go.dev/regexp/syntax) when a
new one appears.

## Where this appears in the repository

Lab 1 currently uses regex for endpoint and middleware paths. Its
[server design](../../labs/lab1/docs/DESIGN.md) explains the application-specific rules
and the `compilePattern` helper; [router.go](../../httpserver/router.go)
contains the implementation. Those details may change independently of the
concepts in this guide.
