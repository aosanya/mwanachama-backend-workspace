# mwanachama-backend-workspace (Go)

Open tasks only — 🚀 In Progress · 📋 Not Started · ⏸️ Blocked.
Everything else is in [todo_done.md](todo_done.md).

The cross-repo rows that put this module into service are `DSN-1750` onward
on `developer`'s design board; the design is
[architecture-workbooks.md](../../../developer/documentation/2.%20design/architecture-workbooks.md).

| Task | Title | Status | Notes |
|------|-------|--------|-------|
| WS1 | 👤 **Does `use` record every act, or only what earns a shortcut?** | 📋 | Currently every act: the simpler rule and the more expensive table. A workbook worked in daily writes a row per operation forever and nothing prunes it, so this becomes a retention question rather than a modelling one. The alternative — record only what a shortcut would rank — means deciding at write time what counts, and loses the workbook's history of what was actually done, which is half of what makes it a record rather than a bookmark folder. Decide before anything depends on the log being complete. |
| WS2 | 👤 **Do shortcuts rank per person or per workbook?** | 📋 | Per workbook today: two people sharing one see the same sidebar, which is one shared artefact and fewer rows. Per person is more useful and multiplies `shortcut` rows by the number of collaborators. `use` already carries `subject`, so the data to decide either way is being written. |
| WS3 | 🔌 **Sharing by grant, beside sharing by key.** | 📋 | `share` covers somebody without an account here. Somebody *with* one is a `mwanachama-backend-permissions` grant on scope `workbook:<id>` — its `scopePattern` already admits that shape and `GET /me/actions` already returns the scopes a caller holds, which is the workbook list's own query. Nothing is wired. Decide whether both exist or only one, because two ways to share one thing is two places to revoke it. |
| WS4 | 🧪 **No Postgres integration test.** | 📋 | Everything runs on SQLite. Catalog carries `postgres_integration_test.go` and overrides `Instance` to a random `pg<hex>` per run so it cannot touch a real deployment — copy that discipline, not just the file. What matters here is the composite primary key on `shortcut` and the `not_deleted` partial indexes, both of which SQLite is laxer about. |
| WS5 | 🃏 **`labelFor` is mechanical, and says so.** | 📋 | It turns `agency.goal.create` into "Create goal" from the action's own words, because this module cannot know what an action means. That is honest but plain, and a sidebar full of "Create goal", "Create role", "Create objective" reads worse than a capability's own wording would. The fix is not to teach this module the words: it is for a caller to supply a label when it records a use, which the shape already allows and nothing yet does. |

_Completed rows move to [todo_done.md](todo_done.md)._
