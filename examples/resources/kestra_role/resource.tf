resource "kestra_role" "example" {
  namespace = "io.kestra.mynamespace"
  name        = "Friendly name"
  description = "Friendly description"

  permissions {
    type        = "FLOW"
    permissions = ["READ", "UPDATE"]
  }

  permissions {
    type        = "TEMPLATE"
    permissions = ["READ", "UPDATE"]
  }
}
