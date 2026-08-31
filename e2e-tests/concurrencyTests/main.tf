# Concurrency limits and quotas on tenants and namespaces.
#
# A separate suite because it is conditional: these fields landed in Kestra 2.0 after
# 2.0.0-rc1, and an instance that predates them accepts the field and drops it silently
# rather than rejecting it. On such an instance the write appears to succeed, the read
# comes back without the block, and the follow-up plan is never empty — so this cannot
# live in surfaceTests, whose whole contract is that the plan is empty after an apply.
# ../../e2e-test.sh probes support once and skips this suite when the fields are absent.
#
# What it adds over the acceptance tests: those assert the blocks round-trip through
# state. This asserts they survive in a real dependency graph and, more importantly, that
# a limit changed outside Terraform shows up as drift rather than being silently
# overwritten — the failure mode the feature exists to close.

terraform {
  required_providers {
    kestra = {
      # No version constraint on purpose: run against the provider built from the working
      # tree via dev_overrides, which never evaluates one.
      source = "kestra-io/kestra"
    }
  }
}

provider "kestra" {
  url       = var.kestra_url
  username  = var.kestra_username
  password  = var.kestra_password
  tenant_id = var.tenant_id
}

resource "kestra_namespace" "concurrency" {
  namespace_id = "io.kestra.e2econcurrency"
  description  = "Namespace carrying a concurrency limit and quotas."

  concurrency {
    limit    = 5
    behavior = "QUEUE"
  }

  # more than one quota on purpose: quotas is an unbounded list, unlike concurrency which
  # is capped at a single block, so ordering and count both need to round-trip
  quotas {
    duration = "PT1H"
    limit    = 10
    behavior = "FAIL"
  }

  quotas {
    duration = "PT24H"
    limit    = 100
    behavior = "CANCEL"
  }
}

resource "kestra_tenant" "concurrency" {
  tenant_id = "e2e-concurrency"
  name      = "E2E Concurrency"

  # QUEUE is valid for a concurrency limit but not for a quota — the two are different
  # enums, so exercise the one that only concurrency accepts
  concurrency {
    limit    = 3
    behavior = "QUEUE"
  }

  quotas {
    duration = "PT30M"
    limit    = 4
    behavior = "CANCEL"
  }
}

data "kestra_namespace" "concurrency" {
  namespace_id = kestra_namespace.concurrency.namespace_id
}

data "kestra_tenant" "concurrency" {
  tenant_id = kestra_tenant.concurrency.tenant_id
}

# The data sources expose these as computed lists. Reading them back through a data
# source rather than the resource's own state is what catches a read path that only
# happens to work because the value was already in state.
check "data_sources_report_the_configured_values" {
  assert {
    condition     = one(data.kestra_namespace.concurrency.concurrency).limit == 5
    error_message = "namespace data source did not report the configured concurrency limit"
  }

  assert {
    condition     = length(data.kestra_namespace.concurrency.quotas) == 2
    error_message = "namespace data source did not report both quotas"
  }

  assert {
    condition     = one(data.kestra_tenant.concurrency.concurrency).limit == 3
    error_message = "tenant data source did not report the configured concurrency limit"
  }
}
