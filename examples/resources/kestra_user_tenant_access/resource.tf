resource "kestra_user_tenant_access" "example" {
  user_id = "4by6NvSLcPXFhCj8nwbZOM"
}

# Tenant access is a prerequisite for tenant-scoped resources: grant it first,
# then have those resources depend on it.
resource "kestra_user_group_membership" "example" {
  user_id  = "4by6NvSLcPXFhCj8nwbZOM"
  group_id = "2ZQwf5Lj9pUyKxRtBvNm8H"

  depends_on = [kestra_user_tenant_access.example]
}
