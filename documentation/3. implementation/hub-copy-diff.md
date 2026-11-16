# Hub copy diff

`mwanachama-backend-workspace` is the custodian of the workspace module. The hub (`projects/hub-steward/code/wakala-api/modules/workspace`) holds a copy made on 2026-11-04 and changed it afterwards. Every change made in that copy exists here, so the two differ by the import path only (hub issue WS6, the workspace half of ISS-105).

## Command

```
scripts/hub-copy-diff.sh [path-to-hub-modules-workspace]
```

The default path is `../projects/hub-steward/code/wakala-api/modules/workspace`, relative to this repository's parent folder. The exit code is 1 when the trees differ, 0 when they do not. The script and `scripts/hubdiff` are the ones of `mwanachama-backend-git` (G78), with the module name changed.

## What it does

1. Copies the `*.go` and `*.json` files of this repository and of the hub folder into two temporary folders, leaving out `cmd/`, `documentation/`, `scripts/` and `.git/`.
2. Rewrites the hub's imports: `wakalaserver/modules/shared/` becomes `github.com/aosanya/mwanachama-backend-shared/`, and `wakalaserver/modules/workspace` becomes `github.com/aosanya/mwanachama-backend-workspace`.
3. Runs every Go file on both sides through `scripts/hubdiff`, which parses it, drops all comments and prints it with `gofmt`, so comments and layout do not show as differences.
4. Runs `diff -r backend hub`.

## Result of 2026-11-16

Every non-test Go file and every JSON file the hub carries is identical after the rewrite, and every test file of the hub exists here and is identical. Nothing in the output is `diff` or `Only in hub`. What remains is `Only in backend` lines:

| Difference | Reason |
|---|---|
| Only here: `domain_agnostic_test.go`, `layout_manager_test.go`, `layout_test.go`, `testdb_test.go`, `workspace_test.go`, `routes/layout_routes_test.go`, `routes/layout_stub_test.go`, `routes/routes_test.go`, `routes/stub_test.go` | The hub copy of 2026-11-04 was made without these test files. They test code that is identical in both. |
| Only here: `spec/examples/` | Example workspace documents the hub never carried. |

## The hub commits carried over

One commit here per hub commit, each naming the hub hash. The tests of both were written first and failed to compile before the change.

| Hub commit | Here | What |
|---|---|---|
| f2f44c31 (MUN137) | 698dbe6 | `routes.Mount` and the `BuildFor`, `RoutesFor`, `SplitFor`, `OwnerRoutesFor`, `BuildToolsFor` and `ToolsFor` entry points, which carry the session; a workbook share key is issued by the caller, not the body. |
| ebef7ae3 (MUN198) | b90eae0 | A workbook is owned and listed for the session, a use names the session as its subject, and every layout act is signed by the caller; the layout routes take a `Mount` (`BuildLayoutRoutesFor`, `LayoutRoutesFor`, `BuildLayoutToolsFor`). |

The older entry points (`Build`, `BuildWith`, `Routes`, `RoutesWith`, `Split`, `SplitWith`, `OwnerRoutes`, `OwnerRoutesWith`, `BuildTools`, `BuildToolsWith`, `Tools`, `ToolsWith`) stay and delegate to the new ones with a `Mount` that holds only the authorizer.
