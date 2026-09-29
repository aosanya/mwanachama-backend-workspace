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

## Layouts

Added 2026-09-29 for the Studio's desks, designed in
[desks.md](../../../mwanachama-wakala-studio/documentation/2.%20design/desks.md).
A workbook can carry its own arrangement: regions, blocks bound to declared
actions, pinned actions on keys. Three roles hold it.

| role | holds |
| --- | --- |
| `layout` | one arrangement, whole or as a merge patch over another |
| `layout_change` | one applied change and its inverse, so it can be undone |
| `layout_choice` | which layout one person sees in one workbook |

**The body is a document, and this module reads exactly one key of it.**
`body` and `set_aside` are `json`. Nothing here knows what a block, a region
or a bound action is — those are opaque, for the same reason a shortcut's
action is. The one exception is `locks`, a list of JSON pointers, because
`read`, `propose` and `patch` have to honour it and the caller cannot be
trusted to honour it against itself. `lock` is already on this module's list
of allowed identifiers, and a pointer is structure rather than a domain word,
so `domain_agnostic_test.go` has nothing to catch.

**Scope is a string, like a reference's.** `workbook:<id>` for a workbook's
own layout, `instance:<id>` for one published to everyone in an agency, a
person's key for their copy of a published one. The module compares it and
never parses it, so `domain_agnostic_test.go` has nothing to catch.

**Form, not two tables.** A whole layout and a patch over a published one
are the same row with `form` set, because a person's patch and the base it
applies to are listed, versioned and retired the same way. `base_version`
records what a patch was last resolved against, so a newer publish is
noticed rather than silently applied.

**A change stores its inverse when it is applied.** Computing the undo later
would need the layout as it was, which is gone by then. An undone change is
marked `reverted_at` and kept, so the history still says it happened.

**Proposing is not applying.** Nothing here stores a proposal. A model's
suggested change lives in the conversation until a person applies it, which
is why `layout_change.subject` is always a person.

**A choice is per person per workbook.** Two people sharing a workbook can
each work in their own arrangement of it, which is the question WS2 asks
about shortcuts, answered the other way here because an arrangement is taste
and a shortcut is a record.

`NewWorkspaceManager` does not require the new roles, so a deployed spec that
names none of them keeps working; `DeclaresLayouts` is the check and
`NewLayoutManager` is a second manager with its own operations file, dispatched
separately.

### Resolving a layout

`effective = locked(base) ∘ mergePatch(base, mine)`, computed in
`mergepatch.go`: apply the person's patch over the base with RFC 7396, then
write every locked pointer back from the base. Both halves are needed —
dropping locked paths out of the patch is not enough on its own, because a
patch may null out a parent of a locked path and take the locked child with
it.

Three calls WS6 left open, settled here:

**A published layout stays editable.** Publishing refuses only a second
publish and a patch that is not `whole`. `desks.md`'s own rules for conflict
("the admin changes an unlocked part a person had overridden", "the admin
locks a block a person had overridden") require an admin to change a
published desk, and its re-resolution happens when a person next opens it,
which `read` does. Locks are dropped from a **patch** form only, so an admin
is never stopped by the locks their own layout declares.

**`set_aside` is derived, never accumulated.** It is recomputed from the
stored patch against the base's current locks every time a layout resolves,
and the column is a cache of that. Union-ing it across writes would leave a
path listed as set aside for ever, and `desks.md` wants a lifted lock to
bring the person's value back.

**The stored patch keeps what the lock held back.** `patch` applies the whole
asked-for patch to the stored document and lets `read` drop the locked parts;
it does not store a pre-filtered patch. That is what makes the previous
paragraph possible.

**An undo writes a change row that is born `reverted_at`.** Otherwise the
next undo picks up the undo itself and undoes it, so ⌘Z twice would toggle
rather than walk back. A reset of one path is an ordinary change and stays
undoable. Redo is not built; `revert` names a `change_id` to reach further
back.

Only the last 20 changes per layout are kept, which is what `desks.md`
promises on *Your view*.

## Open

- Whether `use` records **every** act — an audit log nobody prunes — or only
  what earns a shortcut. Currently every act, which is the simpler rule and
  the more expensive table.
- Whether shortcuts rank per person or per workbook. Currently per workbook:
  two people sharing one see the same sidebar.
- Sharing by key exists; sharing by grant (`workbook:<id>` scope in
  `mwanachama-backend-permissions`) does not, and the two answer different
  questions — a grant needs the other person to have an account.
