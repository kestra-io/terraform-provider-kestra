#!/bin/bash
#===============================================================================
# SCRIPT: e2e-test.sh
#
# DESCRIPTION:
#   Runs the two end-to-end suites against the docker env from ./init-tests-env.sh,
#   using a locally built provider (never the registry).
#
#     e2e-tests/surfaceTests/   "can this be expressed?"  — a wide, flat config.
#                               Apply, then require an empty plan. No behavioural
#                               assertions by design.
#
#     e2e-tests/scenarioTests/  "does it actually work?"   — a customer bootstrap in
#                               identity-scoped stages, ending in real assertions
#                               made over the API as the users Terraform created.
#
# USAGE: ./e2e-test.sh
#===============================================================================

set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
E2E="$ROOT/e2e-tests"

# Single source of truth: must match the `version` pinned in each suite's
# required_providers block. dev_overrides ignores the constraint, so a mismatch is
# silent — keep them equal so the configs stay honest.
PROVIDER_VERSION="0.24.0"

KESTRA_URL="${KESTRA_URL:-http://localhost:8088}"
KESTRA_USERNAME="${KESTRA_USERNAME:-root@root.com}"
KESTRA_PASSWORD="${KESTRA_PASSWORD:-Root!1234}"
KESTRA_TENANT_ID="${KESTRA_TENANT_ID:-main}"

PLUGIN_DIR="$E2E/.terraform/plugins/local/kestra-io/kestra"
TFRC="$E2E/.terraformrc"
export TF_CLI_CONFIG_FILE="$TFRC"
export TF_IN_AUTOMATION=1

step() { printf '\n\033[1m== %s\033[0m\n' "$1"; }

step "build the provider under test"
rm -f "$ROOT/terraform-trace.log"
mkdir -p "$PLUGIN_DIR"
go build -o "$PLUGIN_DIR/terraform-provider-kestra_$PROVIDER_VERSION" "$ROOT"

# dev_overrides paths are resolved against the working directory, and `terraform
# -chdir` changes it per suite. One relative path cannot serve two suites, so write an
# absolute one. A stale path here does NOT error — Terraform silently falls back to the
# registry and you end up testing a published provider instead of this build.
cat > "$TFRC" <<EOF
provider_installation {
  dev_overrides {
    "kestra-io/kestra" = "$PLUGIN_DIR"
  }
}
EOF

tf() {
  local dir="$1"; shift
  TF_LOG_PATH="$ROOT/terraform-trace.log" TF_LOG=DEBUG \
    terraform -chdir="$dir" "$@"
}

# apply, then require the follow-up plan to be empty
apply_and_assert_idempotent() {
  local dir="$1" label="$2"

  step "$label: apply"
  tf "$dir" apply -auto-approve

  step "$label: plan must be empty"
  local code=0
  tf "$dir" plan -detailed-exitcode || code=$?
  case $code in
    0) echo "✅ no changes to apply" ;;
    2) echo "❌ plan is not empty after apply — a resource does not round-trip"; exit 2 ;;
    *) echo "❌ plan failed"; exit 1 ;;
  esac
}

#-------------------------------------------------------------------------------
# Suite A — surface
#-------------------------------------------------------------------------------
SURFACE="$E2E/surfaceTests"
rm -f "$SURFACE"/terraform.tfstate*
apply_and_assert_idempotent "$SURFACE" "surfaceTests"

#-------------------------------------------------------------------------------
# Suite B — scenario
#-------------------------------------------------------------------------------
BOOTSTRAP="$E2E/scenarioTests/00-bootstrap"
PLATFORM="$E2E/scenarioTests/10-platform"

cleanup_scenario() {
  step "scenarioTests: destroy (reverse order)"
  tf "$PLATFORM" destroy -auto-approve || echo "⚠️  10-platform destroy failed"
  tf "$BOOTSTRAP" destroy -auto-approve || echo "⚠️  00-bootstrap destroy failed"
}

export TF_VAR_kestra_url="$KESTRA_URL"
export TF_VAR_tenant_id="$KESTRA_TENANT_ID"
export TF_VAR_kestra_username="$KESTRA_USERNAME"
export TF_VAR_kestra_password="$KESTRA_PASSWORD"

apply_and_assert_idempotent "$BOOTSTRAP" "scenarioTests/00-bootstrap"

# from here on a failure must still tear the instance back down
trap cleanup_scenario EXIT

PLATFORM_ADMIN_TOKEN="$(tf "$BOOTSTRAP" output -raw platform_admin_api_token)"
export TF_VAR_platform_admin_api_token="$PLATFORM_ADMIN_TOKEN"

apply_and_assert_idempotent "$PLATFORM" "scenarioTests/10-platform"

step "scenarioTests/runtime: assert behaviour over the API"
KESTRA_E2E_URL="$KESTRA_URL" \
KESTRA_E2E_TENANT="$KESTRA_TENANT_ID" \
KESTRA_E2E_ADMIN_TOKEN="$PLATFORM_ADMIN_TOKEN" \
KESTRA_E2E_USER_TOKEN="$(tf "$PLATFORM" output -raw app_user_api_token)" \
KESTRA_E2E_NAMESPACE="$(tf "$PLATFORM" output -raw namespace)" \
KESTRA_E2E_FORBIDDEN_NAMESPACE="$(tf "$PLATFORM" output -raw forbidden_namespace)" \
KESTRA_E2E_FLOW_ID="$(tf "$PLATFORM" output -raw flow_id)" \
KESTRA_E2E_FORBIDDEN_FLOW_ID="$(tf "$PLATFORM" output -raw forbidden_flow_id)" \
KESTRA_E2E_TESTSUITE_ID="$(tf "$PLATFORM" output -raw testsuite_id)" \
KESTRA_E2E_KV_VALUE="$(tf "$PLATFORM" output -raw kv_value)" \
  go test -tags e2e -count=1 -v "$ROOT/e2e-tests/scenarioTests/runtime"

step "all e2e suites passed"
