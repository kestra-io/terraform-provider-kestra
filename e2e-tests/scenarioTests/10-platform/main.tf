# Stage 10 — platform.
#
# Applied as the platform admin created in stage 00, authenticating with that user's API
# token. Nothing here runs as the super-admin, which is the point: it is the only place
# in the repo where the provider is exercised through ordinary RBAC.
#
# Identities live in stage 00, not here: /api/v1/users/** is instance-level and
# super-admin only since Kestra 0.24, so a tenant admin cannot create a user, set its
# password or issue its token. What a tenant admin *can* do — and what this stage does —
# is grant an existing identity permissions, through a group, a role and a binding.

terraform {
  required_providers {
    kestra = {
      # No version constraint on purpose. These suites always run against the provider
      # built from the working tree via dev_overrides, which ignores version constraints
      # entirely — so a pin here is never evaluated and can only misinform. (It used to
      # say 0.24.0, which was never a release of this provider; the latest is 1.x.)
      source = "kestra-io/kestra"
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
# The action sets below are taken from Kestra's own managed roles, which are the
# authoritative reference: GET /api/v1/{tenant}/roles/launcher_{tenant} (and
# admin_{tenant}) return exactly the valid resource/action pairs. The API rejects
# anything else with a 422, and neither the published permissions reference (still on the
# pre-2.0 CRUD verbs) nor migrate_role_permissions.go (a migration map — it emits
# SECRET CREATE and EXECUTION CREATE, both refused) can be trusted for this.
#
# The one that catches people: EXECUTION has no CREATE. Starting a flow run is gated
# solely by FLOW EXECUTE. NAMESPACE is absent here for the same reason it is absent from
# the managed Launcher role — viewing flows does not require it.
resource "kestra_role" "launcher" {
  name        = "e2e-launcher"
  description = "View and execute flows; view the resulting executions."

  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "EXECUTE"]
  }

  resources {
    type    = "EXECUTION"
    actions = ["VIEW", "LIST", "ACCESS_LOGS", "ACCESS_OUTPUTS", "FOLLOW"]
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
