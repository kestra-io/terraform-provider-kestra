resource "kestra_template" "example" {
  namespace   = "io.kestra.mynamespace"
  template_id = "my-template"
  content     = <<EOT
tasks:
  - id: t2
    type: io.kestra.core.tasks.log.Log
    message: first {{task.id}}
    level: TRACE
EOT
}
