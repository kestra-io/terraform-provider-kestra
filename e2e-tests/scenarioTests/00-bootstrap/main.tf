# Stage 00 — bootstrap.
#
# Runs as the super-admin declared in `kestra.security.super-admin` (see
# .github/docker/application.yml). That identity is an *input* to Terraform, never a
# resource it manages: Kestra has no API for creating the first super-admin, and
# `kestra_user` has never exposed the privilege.
#
# This stage owns every *identity*, because since 0.24 all of /api/v1/users/** is
# instance-level and super-admin only — creating a user, setting its password and
# issuing its API token cannot be done by a tenant admin. Stage 10 owns authorization
# (groups, roles, bindings), which is tenant-scoped and therefore delegable.

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

# Permissions stage 10 needs. Written out rather than looked up: Kestra creates an Admin
# role per tenant and the ES fixtures seed more, so `data.kestra_role` by name resolves
# to several matches in this environment and fails.
#
# The action vocabulary is the 2.0 fine-grained one and the API validates the
# resource/action *pair*, rejecting a bad one with a 422. Two traps are encoded below:
#
#   * SECRET has no CREATE. A secret is created through UPDATE — the permissions
#     reference says so explicitly, and the API returns
#     "Action CREATE is not valid for resource SECRET".
#   * BINDING has no UPDATE. Bindings are immutable; they are created and deleted.
#
# internal/provider/migrate_role_permissions.go is the closest in-repo reference for the
# vocabulary, but it is a migration map rather than the API's validation table — it emits
# SECRET CREATE, which the API refuses. Trust the API.
resource "kestra_role" "platform_admin" {
  name        = "e2e-platform-admin"
  description = "Manages the scenario namespaces and their IAM. Created by scenarioTests/00-bootstrap."

  resources {
    type    = "NAMESPACE"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE", "MANAGE_FILES"]
  }

  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE"]
  }

  resources {
    type    = "EXECUTION"
    actions = ["VIEW", "LIST"]
  }

  resources {
    type    = "SECRET"
    actions = ["VIEW", "LIST", "UPDATE", "DELETE"]
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
    actions = ["VIEW", "LIST", "CREATE", "UPDATE", "DELETE", "MANAGE_MEMBERS"]
  }
}

#-------------------------------------------------------------------------------
# The platform admin — the identity stage 10 applies as
#-------------------------------------------------------------------------------
resource "kestra_user" "platform_admin" {
  email       = "e2e-platform-admin@kestra.io"
  first_name  = "E2E"
  last_name   = "Platform Admin"
  description = "Applies scenarioTests/10-platform. Created by scenarioTests/00-bootstrap."
}

resource "kestra_user_password" "platform_admin" {
  user_id  = kestra_user.platform_admin.id
  password = var.platform_admin_password
}

resource "kestra_binding" "platform_admin" {
  type        = "USER"
  external_id = kestra_user.platform_admin.id
  role_id     = kestra_role.platform_admin.id
}

# The binding must exist before the token is issued, or stage 10 authenticates as an
# identity with no permissions and fails on its first read.
resource "kestra_user_api_token" "platform_admin" {
  user_id     = kestra_user.platform_admin.id
  name        = "e2e-platform-admin"
  description = "Used by scenarioTests/10-platform."
  max_age     = "PT2H"

  depends_on = [kestra_binding.platform_admin]
}

#-------------------------------------------------------------------------------
# The app user — a regular human. Deliberately gets no binding here: stage 10 grants
# its permissions through a group, so this identity starts with none.
#-------------------------------------------------------------------------------
resource "kestra_user" "app_user" {
  email       = "e2e-app-user@kestra.io"
  first_name  = "E2E"
  last_name   = "App User"
  description = "Regular user. Permissions come from the group stage 10 puts it in."
}

resource "kestra_user_password" "app_user" {
  user_id  = kestra_user.app_user.id
  password = var.app_user_password
}

resource "kestra_user_api_token" "app_user" {
  user_id     = kestra_user.app_user.id
  name        = "e2e-app-user"
  description = "Used by scenarioTests/runtime to act as the app user."
  max_age     = "PT2H"
}
