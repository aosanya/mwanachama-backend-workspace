# Hub import drift

This page lists every difference between the hub's in-tree copy of the workspace module (`projects/hub-steward/code/wakala-api/modules/workspace`) and this repository, measured on 2026-11-19 before the hub imports this module and deletes its copy. It is the input to the merge: each row has a disposition. The method is the one of [hub-copy-diff.md](hub-copy-diff.md).

## Method

1. `scripts/hub-copy-diff.sh ../projects/hub-steward/code/wakala-api/modules/workspace` printed only `Only in backend` lines and exited 0.
2. As a second, stricter check, the hub folder was copied, its two import prefixes rewritten (`wakalaserver/modules/shared/` to `github.com/aosanya/mwanachama-backend-shared/`, `wakalaserver/modules/workspace` to `github.com/aosanya/mwanachama-backend-workspace`) and every file compared byte for byte with the file of the same path here, comments and layout included. No file differs.
3. `git status` of the hub folder was clean, and the newest hub commit that touches it is `ebef7ae3` (2026-11-11), which is already carried over here (see hub-copy-diff.md).
4. `go build ./...` and `go test ./...` pass here on in-memory SQLite.

## Only in the hub

| Item | Disposition |
|---|---|
| A file in `modules/workspace` that this repository lacks | None. Every one of the 29 files of the hub folder exists here, byte-identical after the import rewrite. Nothing to move. |
| A behaviour that this repository lacks | None. Same files, same behaviour. |

## Differing beyond the import path

None. No `.go` or `.json` file of the module differs from its counterpart once the import prefixes are rewritten.

## Only in this repository

| Item | Disposition |
|---|---|
| `domain_agnostic_test.go`, `layout_manager_test.go`, `layout_test.go`, `testdb_test.go`, `workspace_test.go` | Keep. Tests of code both sides share; the hub copy of 2026-11-04 was made without them. |
| `routes/layout_routes_test.go`, `routes/layout_stub_test.go`, `routes/routes_cover_test.go`, `routes/routes_test.go`, `routes/stub_test.go` | Keep. Same reason. |
| `spec/examples/clinic.workspace.json` | Keep. The second example domain that the domain-agnostic test relies on. |
| `spec/examples/wakala.workspace.json` | Keep. Identical to the hub's `internal/specs/wakala.workspace.json`, so the hub can keep its own spec file or point at this example. |
| `cmd/ddl`, `scripts/`, `Makefile`, `documentation/`, `CLAUDE.md`, `README.md`, `LICENSE`, `go.mod`, `go.sum` | Keep. Repository files that a folder inside the hub never carried. |

## Hub-only wiring that is not part of the module

These stay in the hub when it imports this module. They are listed so the import step knows its work.

| Hub file | What it does | Disposition |
|---|---|---|
| `cmd/server/workspace.go` | Reads `WORKSPACE_DATABASE_URL` or `DATABASE_URL`, opens Postgres, calls `Migrate`, `NewWorkspaceManager`, `DeclaresLayouts` and `NewLayoutManager`. | Stays in the hub. Its import changes to this module's path. |
| `cmd/server/mountspec.go` | The `specMount` helper whose `load` and `parse` fields are typed with the hub's in-tree `wakalaserver/modules/shared/spec`. | Stays in the hub. After the import, `workspace.LoadSpec` and `ParseSpec` return `*spec.Spec` of `mwanachama-backend-shared/spec`, which is a different Go type from the in-tree one, so the build breaks until the hub's own shared import is moved the same way. A dependency of the import step, not of the merge. |
| `cmd/server/main.go` (lines about `buildWorkspace`) | Builds the managers and hands them to `Deps`. | Stays in the hub. |
| `internal/api/http/deps.go` | `Deps.Workspace` and `Deps.Layouts` fields. | Stays in the hub. Import path changes. |
| `internal/api/http/workspace_routes.go` | Mounts the public, owner and layout routes under `/workspace`, with the `module:workspace` scope and the session reader as the mount. | Stays in the hub. Import path changes. This is the permission and gating layer. |
| `internal/api/http/operation_routes.go` | Lists `workspace.Operations()` and `workspace.LayoutOperations()` among the mounted modules. | Stays in the hub. Import path changes. |
| `internal/api/http/misuse_shareauthor_test.go`, `internal/api/http/misuse_workbookowner_test.go` | Hub tests that drive the mounted workspace routes through the hub's HTTP layer. | Stay in the hub. Import path changes. |
| `internal/specs/wakala.workspace.json` | The hub's workspace spec. | Stays in the hub. |

No hub tool filter or permission catalogue entry names the workspace module other than the `module:workspace` scope constant in `workspace_routes.go`.

## Owner questions

None. No operation behaves differently on the two sides.

## Result

This repository already is the union of both sides. The merge step changes no source file; it only records this page.
