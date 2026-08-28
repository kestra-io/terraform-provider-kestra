# Stage 00 — bootstrap.
#
# Runs as the super-admin declared in `kestra.security.super-admin` (see
# .github/docker/application.yml). That identity is an *input* to Terraform, never a
# resource it manages: Kestra has no API for creating the first super-admin, and
# `kestra_user` has never exposed the privilege.
#
# This is the only stage that uses those break-glass credentials, which is also the only
# stage a customer would run with them. Everything after this point is applied by the
# platform admin created here.

terraform {
  required_providers {
    kestra = {
      source  = "kestra-io/kestra"
      version = "0.24.0"
    }
  }
}

provider "kestra" {
  url       = var.kestra_url
  username  = var.kestra_username
  password  = var.kestra_password
  tenant_id = var.tenant_id
}

# The platform admin. An ordinary user — the point is that stage 10 authenticates as
# this identity rather than as the super-admin, so every resource it manages is
# exercised through real RBAC instead of through break-glass access.
resource "kestra_user" "platform_admin" {
  email       = "e2e-platform-admin@kestra.io"
  first_name  = "E2E"
  last_name   = "Platform Admin"
  description = "Manages the scenario platform in stage 10. Created by scenarioTests/00-bootstrap."
}

resource "kestra_user_password" "platform_admin" {
  user_id  = kestra_user.platform_admin.id
  password = var.platform_admin_password
}

# Permissions this admin needs to build the platform in stage 10.
#
# Actions are the Kestra 2.0 fine-grained vocabulary, not the pre-2.0 CRUD verbs. The
# old READ/CREATE/UPDATE/DELETE names still round-trip through the API, so a role
# written with them looks correct in state and in a plan while granting nothing — see
# internal/provider/migrate_role_permissions.go for the full mapping.
resource "kestra_role" "platform_admin" {
  name        = "e2e-platform-admin"
  description = "Full management of the scenario namespaces and their IAM."

  resources {
    type    = "NAMESPACE"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE", "MANAGE_FILES", "EXPORT_PLUGIN_DEFAULTS", "IMPORT_PLUGIN_DEFAULTS"]
  }

  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "EXPORT", "CREATE", "IMPORT", "UPDATE", "EXECUTE", "DISABLE", "ENABLE", "VALIDATE", "DELETE"]
  }

  resources {
    type    = "EXECUTION"
    actions = ["VIEW", "LIST", "ACCESS_LOGS", "ACCESS_OUTPUTS", "ACCESS_FILES", "EXPORT", "FOLLOW", "CREATE", "UPDATE", "RESTART", "KILL", "REPLAY", "PAUSE", "RESUME", "CHANGE_LABELS", "UNQUEUE", "FORCE_RUN", "DELETE"]
  }

  resources {
    type    = "SECRET"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }

  resources {
    type    = "KVSTORE"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }

  resources {
    type    = "TESTSUITE"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }

  resources {
    type    = "ROLE"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }

  resources {
    type    = "BINDING"
    actions = ["VIEW", "LIST", "CREATE", "DELETE"]
  }

  resources {
    type    = "GROUP"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "MANAGE_MEMBERS", "DELETE"]
  }

  resources {
    type    = "USER"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "MANAGE_GROUP_MEMBERSHIP", "DELETE"]
  }
}

# Tenant-wide: no `namespace`, because stage 10 creates the namespaces and cannot be
# scoped to namespaces that do not exist yet.
resource "kestra_binding" "platform_admin" {
  type        = "USER"
  external_id = kestra_user.platform_admin.id
  role_id     = kestra_role.platform_admin.id
}

# How stage 10 authenticates. The binding must exist first or the token is issued to an
# identity with no permissions and stage 10 fails on its first read.
resource "kestra_user_api_token" "platform_admin" {
  user_id     = kestra_user.platform_admin.id
  name        = "e2e-platform-admin"
  description = "Used by scenarioTests/10-platform."
  max_age     = "PT2H"

  depends_on = [kestra_binding.platform_admin]
}
