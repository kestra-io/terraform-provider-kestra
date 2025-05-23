---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "kestra_template Resource - terraform-provider-kestra"
subcategory: ""
description: |-
  Manages a Kestra Template.
---

# kestra_template (Resource)

Manages a Kestra Template.

## Example Usage

```terraform
resource "kestra_template" "example" {
  namespace   = "company.team"
  template_id = "my-template"
  content     = <<EOT
tasks:
  - id: t2
    type: io.kestra.core.tasks.log.Log
    message: first {{task.id}}
    level: TRACE
EOT
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `content` (String) The template full content in yaml string.
- `namespace` (String) The template namespace.
- `template_id` (String) The template id.

### Read-Only

- `id` (String) The ID of this resource.
- `tenant_id` (String) The tenant id.

## Import

Import is supported using the following syntax:

```shell
terraform import kestra_template.example {{namespace}}/{{template_id}}
```
