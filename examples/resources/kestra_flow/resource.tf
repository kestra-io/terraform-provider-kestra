resource "kestra_flow" "example" {
  namespace = "company.team"
  flow_id   = "my-flow"
  content   = <<EOT
inputs:
  - name: my-value
    type: STRING

variables:
  first: "1"

tasks:
  - id: t2
    type: io.kestra.core.tasks.log.Log
    message: first {{task.id}}
    level: TRACE

pluginDefaults:
  - type: io.kestra.core.tasks.log.Log
    values:
      message: third {{flow.id}}
EOT
}
