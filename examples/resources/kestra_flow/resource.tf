resource "kestra_flow" "example" {
  namespace = "io.kestra.mynamespace"
  flow_id = "my-flow"
  content = <<EOT
inputs:
  - name: my-value
    type: STRING

variables:
  first: "1"

tasks:
  - id: t2
    type: io.kestra.core.tasks.debugs.Echo
    format: first {{task.id}}
    level: TRACE

taskDefaults:
  - type: io.kestra.core.tasks.debugs.Echo
    values:
      format: third {{flow.id}}
EOT
}
