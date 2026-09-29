---
page_title: "Getting started with Kestra Enterprise"
---

# Getting started with Kestra Enterprise

This guide shows how to manage a Kestra Enterprise Edition instance with Terraform, using a dedicated service account instead of a personal user.

-> The resources used here (`kestra_service_account`, `kestra_service_account_api_token`, `kestra_role`, `kestra_binding`) are only available on the [Enterprise Edition](https://kestra.io/enterprise).

## 1. Start an instance

Use a fresh Kestra Enterprise instance, or an existing one. The first time, you need an account that can create service accounts (for example the basic auth super admin created during the instance setup).

```hcl
provider "kestra" {
  url      = "http://localhost:8080"
  username = var.kestra_admin_username
  password = var.kestra_admin_password
}
```

## 2. Create a service account with super admin rights

```hcl
resource "kestra_service_account" "terraform" {
  name        = "terraform"
  description = "Used by Terraform to manage this instance"
  super_admin = true
}

resource "kestra_service_account_api_token" "terraform" {
  service_account_id = kestra_service_account.terraform.id
  name               = "terraform"
  description        = "Terraform provider token"
  max_age            = "P365D"
}

output "terraform_api_token" {
  value     = kestra_service_account_api_token.terraform.full_token
  sensitive = true
}
```

Apply it, then read the token once with `terraform output -raw terraform_api_token` and store it in your secret manager.

## 3. Give the service account access to tenant data

A super admin manages the instance (tenants, service accounts, users...), but it does not see the data inside a tenant such as flows, namespaces or executions. If you want to terraform tenant data, bind a role to the service account:

```hcl
resource "kestra_role" "admin" {
  name = "terraform-admin"

  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }
}

resource "kestra_binding" "terraform" {
  type        = "USER"
  external_id = kestra_service_account.terraform.id
  role_id     = kestra_role.admin.id
}
```

Adapt the resource types and actions to what you manage. See the [RBAC migration guide](./rbac-migration.md) for the available permissions.

## 4. Configure the provider with the service account

Switch the provider to the API token, either in the configuration or with the `KESTRA_API_TOKEN` environment variable:

```hcl
provider "kestra" {
  url       = "http://localhost:8080"
  api_token = var.kestra_api_token

  # optional, defaults to "main"
  tenant_id = "my-tenant"
}
```

-> Use `api_token` to send a bearer token. If you rather set an `authorization` entry in `extra_headers`, keep in mind it overrides the API token.

## 5. Terraform your instance

You can now manage tenants, namespaces, flows and the other resources of the provider, for example:

```hcl
resource "kestra_flow" "hello" {
  namespace = "company.team"
  flow_id   = "hello"
  content   = <<-EOT
    id: hello
    namespace: company.team
    tasks:
      - id: log
        type: io.kestra.plugin.core.log.Log
        message: Hello from Terraform
  EOT
}
```
