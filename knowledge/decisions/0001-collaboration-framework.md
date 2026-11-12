# 0001: Keep collaboration guidance, knowledge, and planning in Markdown

Parent: [Design decisions](decisions.md).

Status: Accepted

Date: 2026-11-12

## Context

The student group needs agents to find repository context efficiently while
keeping the work understandable to students. Learning notes should be easy to
navigate, and ideas should remain accessible when there is no time to implement
them. The group accepted this as an initial approach to refine through use.

## Decision

- Keep a short root `AGENTS.md` for repository orientation and working rules.
- Use `knowledge/README.md` as the entry point to concept notes, design decisions,
  and links to lab-specific material. Keep official instructions beside each lab.
- Maintain a root `BACKLOG.md` with stable item IDs and Ideas, Ready, In progress,
  and Done sections. Capturing an idea does not authorize its implementation.
- Use plain Markdown and relative links, tracked in Git. Obsidian is optional.

## Alternatives considered

An external task tracker could provide richer assignment and filtering features,
but is not needed for the initial backlog. An Obsidian-specific workflow could add
navigation features, but the group can start with files readable in any editor.

## Consequences

Humans and agents can use the same documents and review their changes alongside
code. The group must maintain navigation links and backlog status manually.
Revisit the layout or tooling if the volume of notes or parallel work makes it
difficult to use.
