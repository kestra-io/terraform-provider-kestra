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

# Only the filename of the locally built plugin, never a released version: the suites
# declare no version constraint because dev_overrides never evaluates one.
PROVIDER_VERSION="0.0.0-dev"

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
# State is deliberately kept between runs. The original script deleted it on every run,
# which orphaned everything the previous run created and made the suite impossible to
# re-run against an instance that was not freshly wiped. Keeping it lets Terraform
# refresh: on a fresh instance the reads 404, the resources drop out of state, and the
# apply recreates them.
apply_and_assert_idempotent "$SURFACE" "surfaceTests"

#-------------------------------------------------------------------------------
# Suite B — scenario
#-------------------------------------------------------------------------------
BOOTSTRAP="$E2E/scenarioTests/00-bootstrap"
PLATFORM="$E2E/scenarioTests/10-platform"

# A failed teardown is a finding, not a warning: it means a delete path is broken, and it
# leaves resources behind that make the next run collide. Record it and fail the run.
destroy_failed=0
cleanup_scenario() {
  step "scenarioTests: destroy (reverse order)"
  tf "$PLATFORM" destroy -auto-approve || { echo "❌ 10-platform destroy failed"; destroy_failed=1; }
  tf "$BOOTSTRAP" destroy -auto-approve || { echo "❌ 00-bootstrap destroy failed"; destroy_failed=1; }
  [ "$destroy_failed" -eq 0 ] || exit 1
}

# Unlike surfaceTests, the scenario stages are created and destroyed within a single run,
# so their state has no value across runs — carrying it over only risks handing the
# runtime a credential from a previous run whose server-side token is long gone.
rm -f "$BOOTSTRAP"/terraform.tfstate* "$PLATFORM"/terraform.tfstate*

export TF_VAR_kestra_url="$KESTRA_URL"
export TF_VAR_tenant_id="$KESTRA_TENANT_ID"
export TF_VAR_kestra_username="$KESTRA_USERNAME"
export TF_VAR_kestra_password="$KESTRA_PASSWORD"

apply_and_assert_idempotent "$BOOTSTRAP" "scenarioTests/00-bootstrap"

# from here on a failure must still tear the instance back down
trap cleanup_scenario EXIT

# stage 00 owns every identity: /api/v1/users/** is super-admin only, so the app user and
# its token are issued here and merely authorized by stage 10
PLATFORM_ADMIN_TOKEN="$(tf "$BOOTSTRAP" output -raw platform_admin_api_token)"
APP_USER_TOKEN="$(tf "$BOOTSTRAP" output -raw app_user_api_token)"
export TF_VAR_platform_admin_api_token="$PLATFORM_ADMIN_TOKEN"
export TF_VAR_app_user_id="$(tf "$BOOTSTRAP" output -raw app_user_id)"

apply_and_assert_idempotent "$PLATFORM" "scenarioTests/10-platform"

step "scenarioTests/runtime: assert behaviour over the API"
KESTRA_E2E_URL="$KESTRA_URL" \
KESTRA_E2E_TENANT="$KESTRA_TENANT_ID" \
KESTRA_E2E_USERNAME="$KESTRA_USERNAME" \
KESTRA_E2E_PASSWORD="$KESTRA_PASSWORD" \
KESTRA_E2E_ADMIN_TOKEN="$PLATFORM_ADMIN_TOKEN" \
KESTRA_E2E_USER_TOKEN="$APP_USER_TOKEN" \
KESTRA_E2E_NAMESPACE="$(tf "$PLATFORM" output -raw namespace)" \
KESTRA_E2E_FORBIDDEN_NAMESPACE="$(tf "$PLATFORM" output -raw forbidden_namespace)" \
KESTRA_E2E_FLOW_ID="$(tf "$PLATFORM" output -raw flow_id)" \
KESTRA_E2E_FORBIDDEN_FLOW_ID="$(tf "$PLATFORM" output -raw forbidden_flow_id)" \
KESTRA_E2E_TESTSUITE_ID="$(tf "$PLATFORM" output -raw testsuite_id)" \
KESTRA_E2E_KV_VALUE="$(tf "$PLATFORM" output -raw kv_value)" \
  go test -tags e2e -count=1 -v "$ROOT/e2e-tests/scenarioTests/runtime"

step "all e2e suites passed"
