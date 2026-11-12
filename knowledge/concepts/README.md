# Concept notes

Parent: [Knowledge base](../README.md).

These explanations are reusable across labs. Pick a subject, then follow its
numbered reading order. The [Lab 1 learning map](../../labs/lab1/docs/LEARNING.md)
connects the topics to the actual implementation.

## Subjects

| Subject index | Start here to understand |
| --- | --- |
| [Go fundamentals](go/go-fundamentals.md) | Types, functions, interfaces, generics, errors, modules, CLI arguments, tests, and regex. |
| [Networking and HTTP](networking/networking-and-http.md) | TCP, HTTP messages, request routing, and middleware. |
| [Files and I/O](files/files-and-io.md) | Streams, buffering, file modes, paths, and validation. |
| [Concurrency](concurrency/concurrency.md) | Goroutines, channels, mutexes, semaphores, keyed locks, and atomic operations. |

## Reading examples

Examples are small fragments unless explicitly called complete programs. A
fragment may omit imports or its enclosing function to focus on one idea. Read
the surrounding explanation before copying it into a program.

## Reading diagrams

Some notes include Mermaid sequence diagrams. Each column is a participant;
read from top to bottom to follow the interaction order. In these notes, solid
arrows show requests/calls and dashed arrows show responses/returns. Notes mark
waiting or skipped work. A participant is not necessarily a separate goroutine.

## Writing notes

Give each note one primary parent topic and list it in that topic's index. Keep
cross-links for useful prerequisites or related ideas; do not link every note to
every other note just to make it discoverable.

Introduce terms before relying on them. Use small, standalone examples and
explain a common misunderstanding. Keep implementation walkthroughs in the
relevant lab's learning map or design notes, with a short link from the concept
note when useful. Include authoritative sources for further reading.

Add a small Mermaid diagram when interaction order, waiting, or a decision is
easier to understand visually. Keep each diagram focused on one idea, introduce
its participants, and retain a short explanation of the takeaway. Prefer plain
sequence diagrams for interactions and flowcharts for decisions; simple facts
do not need a diagram. Keep concept diagrams independent of implementation names.
Use standard fenced `mermaid` blocks and basic syntax; see the
[Mermaid sequence reference](https://mermaid.js.org/syntax/sequenceDiagram.html).
