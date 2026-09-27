# mwanachama-backend-workspace

Workbooks: the unit of work. A person opens one, works in it across whatever
capabilities the work needs, and it accumulates what it touched and what was
done. Shortcuts are earned, never authored.

A Go library backed by Postgres or SQLite through GORM, with its objects
declared rather than written — `Migrate(db, spec)` creates the tables and
`NewWorkspaceManager(db, spec)` returns the manager. There is no service to
run.

## Five objects

| role | holds |
| --- | --- |
| `workbook` | a container of work, owned by one person |
| `reference` | one thing it has touched, remembered well enough to outlive that thing |
| `use` | one act performed here, named by the operation's own declared action |
| `shortcut` | derived from the uses — an act worth offering again |
| `share` | a key that opens it for somebody without an account here |

```go
w, _ := mgr.CreateWorkbook(ctx, workspace.Workbook{Owner: uid, Title: "Q4 planning"})
mgr.Touch(ctx, workspace.Reference{WorkbookID: w.ID, Module: "catalog", Item: slug, Label: "Mwanachama-Wakala"})
mgr.Record(ctx, workspace.Use{WorkbookID: w.ID, Action: "agency.goal.create", Subject: uid})

mgr.ListShortcuts(ctx, w.ID)   // [{Action: "agency.goal.create", Label: "Create goal", Uses: 1}]
```

## A shortcut is earned

"Create goal" appears because somebody created a goal here. No menu is
maintained per capability, and the sidebar shows what this work consists of
rather than what the product can do. What a `use` records is a **declared
action** — the string every spec-driven operation already carries — so a
module converted tomorrow contributes its shortcuts without this one
changing.

## Ownership is a column, not a filter

`ListWorkbooks` takes the owner as its own argument, not a filter field, so
listing somebody else's is not something a caller can do by forgetting to
fill one in. An empty owner is refused rather than treated as a wildcard.

## It names no domain

`spec/examples/` carries two domains that disagree about every noun — a
studio's workbooks and a clinic's case files — and
`domain_agnostic_test.go` fails the build on a word that means something in
only one of them.

Reasoning is in [documentation/2. design/workspace.md](documentation/2.%20design/workspace.md);
the cross-repo record is `developer`'s `architecture-workbooks.md`.
