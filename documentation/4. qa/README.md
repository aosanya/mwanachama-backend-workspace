# QA

`go test ./...` runs everything on SQLite in memory; `make test-pg` is
reserved for the Postgres pass that does not exist yet (WS4).

- `workspace_test.go` — the behaviour that matters: a listing never returns
  somebody else's workbooks and refuses an empty owner rather than treating
  it as a wildcard; touching one thing twice is one row; an act earns a
  shortcut and counts it; a malformed action is refused; a key opens a
  workbook until it is retired, and a wrong key answers exactly as an
  unknown workbook does.
- `domain_agnostic_test.go` — the rule that this module names no domain,
  enforced by parsing its own identifiers and its blueprint, plus a check
  that every emitted identifier fits Postgres's 63 bytes.
- `routes/routes_test.go` — the declared operations built against the
  manager interface, which is where a renamed method or a mistyped argument
  is caught at build rather than at the first request.

Two of these were mutation-checked when written: removing the owner filter
makes the isolation test fail, and removing the shortcut bump makes the
earning test fail.
