# Sanity-check flows that need non-core plugins

These are kept verbatim as real-world examples but are **not loaded** by `locals.tf`:
`fileset(..., "*.yaml")` does not recurse, so moving them here excludes them.

The Kestra 2.0 EE image ships with `/app/plugins` empty — no plugins at all, only the
core task types bundled in the binary. Applying these flows fails at the API with
`No plugin registered for the defined type`:

| Flow | Needs |
|---|---|
| `alerting_sanitychecks.yaml` | `io.kestra.plugin.notifications.slack.SlackExecution` |
| `sync_sanitychecks.yaml` | `io.kestra.plugin.git.SyncFlows`, `io.kestra.plugin.serdes.json.IonToJson`, `io.kestra.plugin.scripts.python.*` |

They applied fine against `v2.0.0-rc1`, so this is an image packaging change rather than an
API or type-name change — `io.kestra.plugin.git.SyncFlows` still exists in the plugin
catalog.

To restore them, install the plugins in `init-tests-env.sh` (for example
`kestra plugins install io.kestra.plugin:plugin-git:LATEST`) and move the files back up
one directory.
