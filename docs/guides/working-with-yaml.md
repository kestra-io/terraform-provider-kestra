---
page_title: "Working with Yaml"
---


# Working with Yaml

Most of Kestra resource need to be described as Yaml like [kestra_flow](../resources/flow.md) & [kestra_template](../resources/template.md).

We have chosen to use a full yaml in terraform definition since the structure is recursive and dynamic, so it can't be described using terraform internal schema.

A flow is sent to Kestra as written, so its comments and indentation are kept. Changes that only touch formatting or comments do not update the flow. The `keep_original_source` provider setting is deprecated and has no effect.

There is in terraform a lot of function that allow to work properly with this yaml content :

## Simple multiline string example

You can use simple terraform multiline string with [Heredoc String](https://www.terraform.io/docs/language/expressions/strings.html#heredoc-strings) :

```terraform
resource "kestra_flow" "example" {
  namespace = "company.team"
  flow_id = "my-flow"
  content = <<EOT
id: my-flow
namespace: company.team  
inputs:
  - name: my-value
    type: STRING
    required: true

tasks:
  - id: t2
    type: io.kestra.core.tasks.log.Log
    message: first {{task.id}}
    level: TRACE
EOT
}
```

## External files

Better will be to use a [file function](https://www.terraform.io/docs/language/functions/file.html). Just create a file `.yml` near your terraform `.tf` and include the whole file in your resource:

```yaml
id: my-flow
namespace: company.team
inputs:
  - name: my-value
    type: STRING
    required: true

tasks:
  - id: t2
    type: io.kestra.core.tasks.log.Log
    message: first {{task.id}}
    level: TRACE
EOT
```

```terraform
resource "kestra_flow" "example" {
  namespace = "company.team"
  flow_id = "my-flow"
  content = file("my-flow.yml")
}
```

## External files with template

Even better will be to use a [templatefile function](https://www.terraform.io/docs/language/functions/templatefile.html) that will allow more complex flows to be more readable. You can include some external external and this one can also include other file.

!> Take care about the indent functon that need to fit your actual flow ident. Terraform don't know anything about your yaml (it's a simple string), so you need to handle properly the indent count by yourself using the [indent function](https://www.terraform.io/docs/language/functions/indent.html)

### Dealing with included yaml string

Imagine a flow that will query an external database. Embedding the full query can lead to very long flow definition. In the case you can use `templatefile` to allow inclusion of an external files from the yaml.

Create a sql file:

```sql
SELECT *
FROM ....
```

Create the yaml file for the flow:

```yaml
tasks:
  - id: "query"
    type: "io.kestra.plugin.jdbc.mysql.Query"
    url: jdbc:postgresql://127.0.0.1:56982/
    username: postgres
    password: pg_passwd
    sql: |
      ${indent(6, file("my-query.sql"))}
    fetchOne: true
```

And finally create the resource invoking the `templatefile`:

```terraform
resource "kestra_flow" "example" {
  namespace = "company.team"
  flow_id = "my-flow"
  content = templatefile("my-flow.yaml", {})
}
```

The `tf` files will required the `yaml` files that will require the `sql` files and the final flow will be:

```yaml
tasks:
  - id: "query"
    type: "io.kestra.plugin.jdbc.mysql.Query"
    url: jdbc:postgresql://127.0.0.1:56982/
    username: postgres
    password: pg_passwd
    sql: |
      SELECT *
      FROM ....
    fetchOne: true
```

### Include full yaml part

By the same way, you can also include a full yaml specs inside another one.

Create 2 yaml files:

```yaml
id: t1
type: io.kestra.core.tasks.log.Log
message: first {{task.id}}
level: TRACE
```

```yaml
id: t2
type: io.kestra.core.tasks.log.Log
message: second {{task.id}}
level: TRACE
```

Create the yaml file for the flow:

```yaml
tasks:
  - ${indent(4, file("t1.yml"))}
  - ${indent(4, file("t2.yml"))}
```

## Flow metadata as attributes

`description`, `disabled` and `labels` can also be set as attributes of `kestra_flow`, which lets them come from Terraform variables or other resources without templating the yaml:

```terraform
resource "kestra_flow" "example" {
  namespace   = "company.team"
  flow_id     = "my-flow"
  description = "Loads the daily orders"
  disabled    = var.environment != "production"
  labels = {
    team = "data"
    env  = var.environment
  }
  content = file("my-flow.yml")
}
```

The attributes are appended to the yaml sent to Kestra and left out of `content` when the flow is read back. A key is managed either by its attribute or in the yaml, never both: setting an attribute that also appears in `content` is an error at plan time, even when both values are equal. Remove an attribute to hand its key back to the yaml.
