# CLAUDE.md

Guidance for Claude Code (claude.ai/code) working in this repository.

## Project: mwanachama-backend-workspace

Workbooks: the unit of work. A person opens one, works in it across whatever
capabilities the work needs, and it accumulates what it touched and what was
done in it. Shortcuts are earned, never authored.

Built 2026-09-27 on [`mwanachama-backend-catalog`](../mwanachama-backend-catalog)'s
shape, at the owner's instruction — read that repo before changing this one.
Blueprint plus domain spec, `specstore` rather than row structs, operations
declared in JSON and dispatched into both routes and MCP tools, a
`domain_agnostic_test.go` that fails the build on a domain word, two example
domains, `cmd/ddl`, a Makefile.

**Reusing catalog itself was tried and measured first.** It needs all nine
of catalog's roles (four of them meaningless here, and their names broke the
63-byte identifier budget for registry-minted agencies), `ListEntries`
strips `doc` so the workbook list could show nothing from inside it (CAT8),
and ownership would have been a doc-path filter every caller must remember.
The full measurement is in
[documentation/2. design/workspace.md](documentation/2.%20design/workspace.md).

## It names no domain, and that is the point

Nothing here knows what an action means or what a reference points at.
`agency.goal.create` and `x_ray.order.approve` are both opaque strings.

> A word that means something in one domain and nothing in another does not
> belong in this module.

`domain_agnostic_test.go` enforces it, and `spec/examples/` carries a
studio's workbooks and a clinic's case files so the claim is tested rather
than asserted.

## Conventions

- **The objects are declared, not written.** `workspace.blueprint.json`
  declares every object, field and index; a domain spec names them. No row
  structs, no `AutoMigrate`. Adding a field means editing the blueprint and
  the `models/` type together — they are checked against each other when the
  manager is built.
- **A table is `<instance>_workspace_<table>`.** Three segments, longest
  identifier 56 for a registry-minted agency — measured, and the reason this
  module needs no fourth segment.
- **Ownership is an argument, never a filter field.** `ListWorkbooks(ctx,
  owner, filter)`. A filter field can be forgotten; an argument cannot.
  `WorkbookFilter` must not gain an `Owner` — `dispatch` refuses the build
  when two arguments arrive as `owner`, which is how this was caught.
- **A shortcut is derived.** `Record` writes the act and moves the shortcut
  in one transaction. Nothing authors a shortcut.
- **Only `workspace.workbook.open` is anonymous**, because somebody holding
  a key has no account here. Everything else arrives gated.
- Documentation follows the workspace-wide four-phase layout
  (`1. requirements`/`2. design`/`3. implementation`/`4. qa`).

## Code comments

Write code with no comments. Not one-liners above a function, not section
banners, not doc comments on exported symbols, not "why" notes next to a
tricky line. A name, a type, or a smaller function carries it instead.

Anything that genuinely needs explaining goes in this repo's `documentation/`
folder, under the phase it belongs to (`1. requirements`, `2. design`,
`3. implementation`, `4. qa`) — never inline.

**Why:** inline prose drifts out of sync with the code, duplicates what
`documentation/` already owns, and buries the explanation where nobody
looking for it will search.

**How to apply:**

- New code ships without comments. If a line seems to need one, rename or
  split until it doesn't.
- Touching code that already has comments: strip the ones in the code you are
  changing. Do not sweep untouched files unless asked.
- If the reasoning matters, add or update the matching `documentation/` page
  in the same change and leave nothing behind in the source.
- Machine-read directives are not comments and stay: build tags, `//go:embed`,
  `//go:generate`, linter pragmas (`//nolint`, `// eslint-disable-next-line`,
  `// ignore:`), license headers, codegen "do not edit" banners, and generated
  files as a whole.
- Commit messages, PR descriptions, and test names carry the narration that
  used to go in comments.

This rule is repeated verbatim in every mwanachama repo's `CLAUDE.md` so that
it reaches sessions that do not load this machine's user-level config —
scheduled cloud routines, other machines, and other agent harnesses.
