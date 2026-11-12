# Knowledge base

Start here to find explanations, assignment context, and design decisions without
reading every document. The repository root is the Obsidian vault, with Markdown
notes in `knowledge/` and lab-specific material beside each lab.
For installation and running the code, start with the
[repository README](../README.md) or [Lab 1 README](../labs/lab1/README.md).

## Navigation

Use a tree for the primary reading path:

```text
Knowledge base
├── Concepts
│   ├── Go fundamentals
│   ├── Networking and HTTP
│   ├── Files and I/O
│   └── Concurrency
├── Lab 1 learning map
└── Design decisions
```

- [Concept notes](concepts/README.md): subject indexes and independent explanations.
- [Lab 1 learning map](../labs/lab1/docs/LEARNING.md): a study path, concept coverage,
  and a request walkthrough tied to the current code.
- [Design decisions](decisions/decisions.md): adopted choices and their reasoning.

Each concept has one primary parent index. Useful cross-links are allowed:
shared concepts such as ownership should have one explanation even when several
topics need it. Obsidian's graph displays those links, so the graph is not expected
to be a strict tree. See [Graph view](https://help.obsidian.md/plugins/graph).

For project coordination, use [agent guidance](../AGENTS.md) and the
[backlog](../BACKLOG.md).

## Labs

There are four labs; the first three are mandatory. Lab 1 instructions are
available; instructions for Labs 2–4 are pending. Keep supplied instructions
beside the corresponding lab and link them here.

| Lab | Implementation | Notes | Official instructions |
| --- | --- | --- | --- |
| 1 | [Lab 1](../labs/lab1) | [Learning map](../labs/lab1/docs/LEARNING.md), [agent guide](../labs/lab1/AGENTS.md), [server design](../labs/lab1/docs/DESIGN.md) | [Instructions (PDF)](../labs/lab1/docs/lab1_instructions.pdf) |
| 2 | [Lab 2](../labs/lab2) (placeholder) | None yet | Pending |
| 3 | [Lab 3](../labs/lab3) (placeholder) | None yet | Pending |
| 4 | [Lab 4](../labs/lab4) (placeholder) | None yet | Pending |

## Writing and maintaining notes

- Put cross-lab explanations under their subject in `concepts/`. Keep lab-specific
  walkthroughs beside the relevant lab and link them rather than copying concepts.
- Record significant design choices in `decisions/` using the next available
  number, for example `0002-descriptive-title.md`. Include status, context, the
  choice, alternatives considered, and consequences. Mark proposals explicitly;
  update or supersede accepted decisions when the group changes direction.
- Use ordinary relative Markdown links so notes work in editors, Git viewers,
  and Obsidian. No plugins or generated indexes are required.
- Distinguish official requirements from explanations, adopted decisions, and
  ideas. Link evidence for experiments and label unverified assumptions.
- Add links to new notes in the appropriate index and update affected notes when
  code or decisions change. Keep each topic in one place where practical.
- Treat `.obsidian/` as editor configuration rather than the source of knowledge.
  The navigation and explanations remain readable without plugins or Obsidian.
