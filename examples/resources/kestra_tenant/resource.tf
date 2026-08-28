resource "kestra_tenant" "example" {
  tenant_id = "my-tenant"
  name      = "My Tenant"

  # cap how many executions of the whole tenant run at once
  concurrency {
    limit    = 10
    behavior = "QUEUE"
  }

  # and how many may start inside a sliding window
  quotas {
    duration = "PT1H"
    limit    = 100
    behavior = "FAIL"
  }
}
