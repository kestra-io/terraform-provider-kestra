terraform {
  required_providers {
    kestra = {
      source = "kestra-io/kestra"
      version = "0.24.0"
    }
  }
}
provider "kestra" {
  # mandatory, the Kestra webserver/standalone URL
  url = "http://localhost:8080"

  # basic auth
  username = "root@root.com"
  password = "Root!1234"
}

resource "kestra_group" "my_group_1" {
  namespace   = "io.kestra.terraform.e2e.data"
  name        = "my-group-1"
  description = "my-group-1 description"
}
resource "kestra_user" "example" {
  namespace   = "io.kestra.terraform.e2e.data"
  description = "Friendly description"
  first_name  = "John"
  last_name   = "Doe"
  email       = "john@example.com"
}
resource "kestra_user" "example2" {
  namespace   = "io.kestra.terraform.e2e.data"
  description = "Friendly descriptionracevedo"
  first_name  = "Johnracevedo"
  last_name   = "Doeracevedo"
  email       = "racevedo@example.com"
  groups      = [kestra_group.my_group_1.id]
}
resource "kestra_namespace_secret" "environment" {
  namespace = "io.kestra.terraform.e2e.data"
  secret_key = "ENVIRONMENT"
  secret_value = "my-secret-1"
  secret_description = "AUTO Environment name"
  secret_tags = { "AUTO" : "true", "TERRAFORM" : "true" }
}

resource "kestra_namespace_secret" "gitlab_token" {
  namespace = "io.kestra.terraform.e2e.data"
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
  key = "tenant-worker-group"
}

resource "kestra_flow" "ekestra_flowxample" {
  namespace = "io.kestra.terraform.e2e.data"
  flow_id   = "return-flow"
  content   = file("my-flow.yml")
}

# this is just an example importing real world flows, sanity-checks are not actually runned/used
resource "kestra_flow" "sanity_checks_flows" {
  for_each = local.sanitycheck_files

  flow_id   = yamldecode(templatefile("${path.module}/example-sanity-checks/${each.value}", {}))["id"]
  namespace = yamldecode(templatefile("${path.module}/example-sanity-checks/${each.value}", {}))["namespace"]
  content   = templatefile("${path.module}/example-sanity-checks/${each.value}", {})
}



resource "kestra_test" "kestra_testsuite_example" {
  namespace = "io.kestra.terraform.e2e.data"
  test_id   = "simple-return-test-suite-1-id"
  content   = file("my-test-suite.yml")
  depends_on = [kestra_flow.ekestra_flowxample]
}

resource "kestra_namespace" "kestra_namespaceexample" {
  namespace_id    = "io.kestra.terraform.e2e.data.addednamespace"
  description     = "Friendly description"

  variables       = <<EOT
k1: 1
k2:
    v1: 1
EOT
  plugin_defaults = <<EOT
- type: io.kestra.core.tasks.log.Log
  forced: false
  values:
    message: first {{flow.id}}
- type: io.kestra.core.tasks.debugs.Return
  forced: false
  values:
    format: first {{flow.id}}
EOT

  allowed_namespaces {
    namespace = "io.kestra.terraform.e2e.allowed"
  }

  worker_group {
    key      = "tenant-worker-group"
    fallback = "WAIT"
  }

  storage_type = "s3"
  storage_configuration = {
    bucket = "my-namespace-bucket"
    region = "eu-west-1"
  }

  storage_isolation {
    enabled = true
    denied_services = ["EXECUTOR", "SCHEDULER"]
  }

  secret_isolation {
    enabled = false
    denied_services = []
  }

  secret_read_only = false
  secret_type = "aws-secret-manager"
  secret_configuration = {

    accessKeyId = "mysuperaccesskey"
    secretKeyId = "mysupersecretkey"
    sessionToken = "mysupersessiontoken"
    region = "us-east-1"
  }

  outputs_in_internal_storage = true
}
# resource "kestra_binding" "exgggample" {
#   type        = "GROUP"
#   external_id = "68xAawPfiJPkTkZJIPX6jQ"
#   role_id     = "3kcvnr27ZcdHXD2AUvIe7z"
#   namespace   = "io.kestra.terraform.e2e.data"
# }
resource "kestra_role" "exarrrmple" {
  namespace   = "io.kestra.terraform.e2e.data"
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
# resource "kestra_template" "exhhhample" {
#   namespace   = "io.kestra.terraform.e2e.data"
#   template_id = "my-template"
#   content     = <<EOT
# tasks:
#   - id: t2
#     type: io.kestra.core.tasks.log.Log
#     message: first {{task.id}}
#     level: TRACE
# EOT
# }
resource "kestra_tenant" "exahhhmple" {
  tenant_id = "my-tenant"
  name      = "My Tenant"

  worker_group {
    key      = kestra_worker_group.wkggg.key
    fallback = "FAIL"
  }

  storage_type = "s3"
  storage_configuration = {
    bucket = "my-tenant-bucket"
    region = "eu-west-1"
  }

  storage_isolation {
    enabled = false
    denied_services = []
  }

  secret_isolation {
    enabled = false
    denied_services = []
  }

  secret_type = "aws-secret-manager"
  secret_read_only = true
  secret_configuration = {
    accessKeyId = "mysuperaccesskey"
    secretKeyId = "mysupersecretkey"
    sessionToken = "mysupersessiontoken"
    region = "us-east-1"
  }

  require_existing_namespace = false
  outputs_in_internal_storage = false
}
# resource "kestra_user_api_token" "examdfgple" {
#   user_id = "4DPVrcZKRZnCMGMYoTDRaj"
#
#   name        = "test-token"
#   description = "Test token"
#   max_age     = "PT1H"
#   extended    = false
# }

# resource "kestra_user_password" "exampertrele" {
#   user_id  = "4DPVrcZKRZnCMGMYoTDRaj"
#   password = "6ZEUYT32fdsfmy-random-password"
# }
