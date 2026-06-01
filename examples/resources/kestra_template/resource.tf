resource "kestra_template" "example" {
  namespace   = "company.team"
  template_id = "my-template"
  content     = <<EOT
tasks:
  - id: t2
    type: io.kestra.plugin.core.log.Log
    message: first {{task.id}}
    level: TRACE
EOT
}
