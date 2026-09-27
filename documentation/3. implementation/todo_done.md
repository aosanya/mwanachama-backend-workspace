# mwanachama-backend-workspace — completed

Rows are cut here from [todo.md](todo.md) when they close.

| Task | Title | Closed | Notes |
|------|-------|--------|-------|
| WS0 | 🧱 **The module: blueprint, manager, generated routes and tools.** | 2026-09-27 | Built on catalog's shape after reusing catalog itself was tried and measured against it — see `documentation/2. design/workspace.md` for why it lost. Five declared objects, 15 operations dispatched into routes and MCP tools with per-field prose read from the blueprint, two example domains that disagree about every noun, `cmd/ddl`, a Makefile. Ownership is an argument rather than a filter field, which `dispatch` enforced during the build: declaring `Owner` on both made two arguments arrive as `owner` and the tool build refused it. Two claims mutation-checked — removing the owner filter makes the isolation test list both people's workbooks, and removing the shortcut bump makes the earning test find none. |
