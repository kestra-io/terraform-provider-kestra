resource "kestra_namespace" "example" {
  namespace_id    = "company.team"
  description     = "Friendly description"
  variables       = <<EOT
k1: 1
k2:
    v1: 1
EOT
  plugin_defaults = <<EOT
- type: io.kestra.plugin.core.log.Log
  values:
    message: first {{flow.id}}
- type: io.kestra.plugin.core.debug.Return
  values:
    format: first {{flow.id}}
EOT
}
