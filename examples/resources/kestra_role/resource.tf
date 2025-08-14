resource "kestra_role" "example" {
  name        = "Friendly name"
  description = "Friendly description"

  permissions {
    type        = "FLOW"
    permissions = ["READ", "UPDATE"]
  }

  permissions {
    type        = "EXECUTION"
    permissions = ["READ", "UPDATE"]
  }
}
