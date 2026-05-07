resource "kestra_role" "example" {
  name        = "Friendly name"
  description = "Friendly description"

  resources {
    type    = "FLOW"
    actions = ["VIEW", "LIST", "UPDATE", "EXECUTE"]
  }

  resources {
    type    = "EXECUTION"
    actions = ["VIEW", "LIST", "ACCESS_LOGS"]
  }
}
