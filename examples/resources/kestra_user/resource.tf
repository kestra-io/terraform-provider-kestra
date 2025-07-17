resource "kestra_user" "example" {
  email       = "john@doe.com"
  namespace   = "company.team"
  description = "Friendly description"
  first_name  = "John"
  last_name   = "Doe"
  groups      = ["4by6NvSLcPXFhCj8nwbZOM"]
}
