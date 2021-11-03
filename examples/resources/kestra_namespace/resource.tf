resource "kestra_namespace" "example" {
  namespace_id = "io.kestra.mynamespace"
  description = "Friendly description"
  variables = <<EOT
k1: 1
k2:
    v1: 1
EOT
  task_defaults = <<EOT
- type: io.kestra.core.tasks.debugs.Echo
  values:
    format: first {{flow.id}}
- type: io.kestra.core.tasks.debugs.Return
  values:
    format: first {{flow.id}}
EOT
}
