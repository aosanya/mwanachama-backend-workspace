# Workspace — why it is shaped this way

The cross-repo design record is
[architecture-workbooks.md](../../../developer/documentation/2.%20design/architecture-workbooks.md).
This page carries what would otherwise be comments in the source.

## Why not a catalog domain

`mwanachama-backend-catalog` was tried first and measured, on 2026-09-27.
Three things decided against it.

A workbook domain must name **all nine** of catalog's roles, because
`NewCatalogManager` requires all nine carriers — so `suggestion`,
`suggestion_comment`, `suggestion_match` and `run` would exist as empty
tables in every agency, forever, and their long names broke the 63-byte
identifier budget for registry-minted agencies.

`ListEntries` strips `doc` unconditionally (catalog's own CAT8), so the
workbook list — the product's front door — could show a title and a summary
and nothing else. No shortcut counts, no last-touched.

And ownership would have been `Docs: [{Path: "owner"}]`, a filter every
caller must remember, in a table holding every person in that agency. That
is the shape of the WK39 leak the security sweep found in insights the same
day.

What it gave up: catalog's `share_link` was free. It is rebuilt here as
`share`, deliberately in the same shape — issue, retire, rotate, retired
keys kept for the trail — because that design was already proved.

## Decisions

**Ownership is an argument, not a filter field.** `ListWorkbooks(ctx, owner,
filter)` rather than a filter carrying an owner. A filter field can be
forgotten; an argument cannot, and an empty one is refused rather than read
as a wildcard. `WorkbookFilter` deliberately has no `Owner`, which
`dispatch` enforces — declaring both made two arguments arrive as `owner`
and the tool build refused it.

**A reference is one row per thing, not per visit.** `Touch` moves
`last_seen` on the row that already exists. A workbook that opened one
agency forty times has touched one thing; the forty visits are uses.

**A reference keeps its own label.** Captured when the thing was touched, so
a row still renders after the thing it names is deleted. A workbook is a
record of work, and a record that breaks when its subject goes is not one.

**A shortcut is derived, never authored.** `Record` writes the act and moves
the shortcut in one transaction. `labelFor` turns an action into a label
mechanically — this module does not know what any action means, so it can do
no better than the words the action itself carries, and a caller wanting a
nicer label supplies one.

**Opening with a key answers the same for wrong, retired and unknown.**
Telling them apart lets anyone probe which workbooks exist.

**Only `workspace.workbook.open` is anonymous.** Somebody sent a key has no
account here, which is the whole reason a key exists. Every other operation
arrives gated, and the tools exclude the anonymous one because it has no
place in an operator's toolbox.

## Open

- Whether `use` records **every** act — an audit log nobody prunes — or only
  what earns a shortcut. Currently every act, which is the simpler rule and
  the more expensive table.
- Whether shortcuts rank per person or per workbook. Currently per workbook:
  two people sharing one see the same sidebar.
- Sharing by key exists; sharing by grant (`workbook:<id>` scope in
  `mwanachama-backend-permissions`) does not, and the two answer different
  questions — a grant needs the other person to have an account.
