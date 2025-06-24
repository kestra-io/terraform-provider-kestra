terraform {
  required_providers {
    kestra = {
      source = "kestra-io/kestra"
      version = "0.23.0"
    }
  }
}
provider "kestra" {
  # mandatory, the Kestra webserver/standalone URL
  url = "http://localhost:8088"

  # optional basic auth username
  username = "root@root.com"

  # optional basic auth password
  password = "Root!1234"
}

resource "kestra_user" "example" {
  username    = "my-username"
  namespace   = "company.team"
  description = "Friendly description"
  first_name  = "John"
  last_name   = "Doe"
  email       = "john@doe.com"
  groups      = ["4by6NvSLcPXFhCj8nwbZOM"]
}
resource "kestra_user" "example2" {
  username    = "my-username-racevedo"
  namespace   = "company.team"
  description = "Friendly descriptionracevedo"
  first_name  = "Johnracevedo"
  last_name   = "Doeracevedo"
  email       = "racevedo@example.com"
  groups      = ["4by6NvSLcPXFhCj8nwbZOM"]
}
resource "kestra_namespace_secret" "environment" {
  namespace = "predoc"
  secret_key = "ENVIRONMENT"
  secret_value = "my-secret-1"
  secret_description = "AUTO Environment name"
  secret_tags = { "AUTO" : "true", "TERRAFORM" : "true" }
}

resource "kestra_namespace_secret" "gitlab_token" {
  namespace = "predoc"
  secret_key = "GITLAB_TOKEN"
  secret_value = "my-secret-2ddd"
  secret_description = "AUTO Gitlab token used to pull artifacts from gitlab"
  secret_tags = { "AUTO" : "true", "TERRAFORM" : "true" }
}
# resource "kestra_service_account" "examplesa" {
#   name        = "my-service-account"
#   description = "Friendly description"
# }
resource "kestra_worker_group" "wkggg" {
  key = "myworkergrouP1"
}

resource "kestra_flow" "ekestra_flowxample" {
  namespace = "io.kestra.terraform.data"
  flow_id   = "my-flow"
  content   = file("my-flow.yml")
}

resource "kestra_namespace" "kestra_namespaceexample" {
  namespace_id    = "company.team"
  description     = "Friendly description"
  variables       = <<EOT
k1: 1
k2:
    v1: 1
EOT
  plugin_defaults = <<EOT
- type: io.kestra.core.tasks.log.Log
  values:
    message: first {{flow.id}}
- type: io.kestra.core.tasks.debugs.Return
  values:
    format: first {{flow.id}}
EOT
}
resource "kestra_group" "exfsfsample" {
  namespace   = "company.team"
  name        = "Friendly name"
  description = "Friendly description"
}
# resource "kestra_binding" "exgggample" {
#   type        = "GROUP"
#   external_id = "68xAawPfiJPkTkZJIPX6jQ"
#   role_id     = "3kcvnr27ZcdHXD2AUvIe7z"
#   namespace   = "company.team"
# }
resource "kestra_role" "exarrrmple" {
  namespace   = "company.team"
  name        = "Friendly name"
  description = "Friendly description"

  permissions {
    type        = "FLOW"
    permissions = ["READ", "UPDATE"]
  }

  permissions {
    type        = "TEMPLATE"
    permissions = ["READ", "UPDATE"]
  }
}
resource "kestra_template" "exhhhample" {
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
resource "kestra_tenant" "exahhhmple" {
  tenant_id = "my-tenant"
  name      = "My Tenant"
}
# resource "kestra_user_api_token" "examdfgple" {
#   user_id = "4DPVrcZKRZnCMGMYoTDRaj"
#
#   name        = "test-token"
#   description = "Test token"
#   max_age     = "PT1H"
#   extended    = false
# }
resource "kestra_user" "exahqmple" {
  username    = "my-username"
  namespace   = "company.team"
  description = "Friendly description"
  first_name  = "John"
  last_name   = "Doe"
  email       = "john@doe.com"
  groups      = ["4by6NvSLcPXFhCj8nwbZOM"]
}
# resource "kestra_user_password" "exampertrele" {
#   user_id  = "4DPVrcZKRZnCMGMYoTDRaj"
#   password = "6ZEUYT32fdsfmy-random-password"
# }
