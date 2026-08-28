# Stage 10 — platform.
#
# Applied as the platform admin created in stage 00, authenticating with that user's API
# token. Nothing here runs as the super-admin, which is the point: it is the only place
# in the repo where the provider is exercised through ordinary RBAC.
#
# Identities live in stage 00, not here: /api/v1/users/** is instance-level and
# super-admin only since 0.24, so a tenant admin cannot create a user, set its password
# or issue its token. What a tenant admin *can* do — and what this stage does — is grant
# an existing identity permissions, through a group, a role and a binding.

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
  api_token = var.platform_admin_api_token
  tenant_id = var.tenant_id
}

locals {
  namespace           = "io.kestra.e2escenario.allowed"
  forbidden_namespace = "io.kestra.e2escenario.forbidden"
  kv_value            = "e2e-kv-value"
}

#-------------------------------------------------------------------------------
# Namespaces
#-------------------------------------------------------------------------------
resource "kestra_namespace" "allowed" {
  namespace_id = local.namespace
  description  = "Namespace the app user is bound to."
}

resource "kestra_namespace" "forbidden" {
  namespace_id = local.forbidden_namespace
  description  = "Sibling namespace the app user must not reach."
}

#-------------------------------------------------------------------------------
# Namespace-scoped resources the flow consumes at runtime
#-------------------------------------------------------------------------------
resource "kestra_namespace_secret" "token" {
  namespace          = kestra_namespace.allowed.namespace_id
  secret_key         = "E2E_SCENARIO_TOKEN"
  secret_value       = "e2e-scenario-secret-value"
  secret_description = "Read by e2e_scenario_flow to prove secrets resolve at runtime."
  secret_tags        = { "E2E" : "true" }
}

resource "kestra_kv" "scenario" {
  namespace = kestra_namespace.allowed.namespace_id
  key       = "E2E_SCENARIO_KV"
  type      = "STRING"
  value     = local.kv_value
}

resource "kestra_namespace_file" "marker" {
  namespace = kestra_namespace.allowed.namespace_id
  filename  = "/scenario-marker.txt"
  content   = "e2e-scenario-file"
}

#-------------------------------------------------------------------------------
# Flows and the test suite
#-------------------------------------------------------------------------------
resource "kestra_flow" "scenario" {
  namespace = kestra_namespace.allowed.namespace_id
  flow_id   = "e2e_scenario_flow"
  content   = file("${path.module}/flows/scenario_flow.yml")

  # the flow reads all three at runtime; without this Terraform is free to create the
  # flow first and the first execution races the fixtures it depends on
  depends_on = [
    kestra_namespace_secret.token,
    kestra_kv.scenario,
    kestra_namespace_file.marker,
  ]
}

resource "kestra_flow" "forbidden" {
  namespace = kestra_namespace.forbidden.namespace_id
  flow_id   = "e2e_forbidden_flow"
  content   = file("${path.module}/flows/forbidden_flow.yml")
}

resource "kestra_test" "scenario" {
  namespace = kestra_namespace.allowed.namespace_id
  test_id   = "e2e_scenario_testsuite"
  content   = file("${path.module}/flows/scenario_testsuite.yml")

  depends_on = [kestra_flow.scenario]
}

#-------------------------------------------------------------------------------
# Authorization for the app user created in stage 00
#-------------------------------------------------------------------------------
resource "kestra_group" "app_users" {
  name        = "e2e-app-users"
  description = "Users allowed to launch the scenario flow."
}

# Deliberately narrow, so the negative assertions in scenarioTests/runtime have
# something real to bite on: read and run flows, read the resulting executions, and
# nothing else. No CREATE on FLOW, no SECRET, no KVSTORE.
#
# Every action below appears in internal/provider/migrate_role_permissions.go, which
# mirrors the API's own vocabulary. The API validates action-per-resource and rejects
# anything else with a 422, so that file — not the CRUD verbs in the older docs — is the
# reference to write roles against.
resource "kestra_role" "launcher" {
  name        = "e2e-launcher"
  description = "View and execute flows; view the resulting executions."

  resources {
    type    = "NAMESPACE"
    actions = ["VIEW", "LIST"]
  }

  # EXECUTE is a FLOW action in 2.0, while starting a run also needs CREATE on
  # EXECUTION — a role with only one of the two reads as correct and does not work.
  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "EXECUTE"]
  }

  resources {
    type    = "EXECUTION"
    actions = ["VIEW", "LIST", "CREATE", "ACCESS_LOGS", "ACCESS_OUTPUTS", "FOLLOW"]
  }
}

# Scoped to the allowed namespace only. Kestra grants a namespace binding to child
# namespaces too, which is why the forbidden namespace is a sibling rather than a child.
resource "kestra_binding" "app_users" {
  type        = "GROUP"
  external_id = kestra_group.app_users.id
  role_id     = kestra_role.launcher.id
  namespace   = kestra_namespace.allowed.namespace_id
}

# Membership rather than kestra_user.groups: that attribute is a full replacement and
# would fight anything else managing this user's groups. It is also the one piece of
# identity wiring a tenant admin may do, since it is tenant-scoped.
resource "kestra_user_group_membership" "app_user" {
  user_id  = var.app_user_id
  group_id = kestra_group.app_users.id
}
