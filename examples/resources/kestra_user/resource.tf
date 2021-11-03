resource "kestra_user" "example" {
  username = "my-username"
  namespace = "io.kestra.mynamespace"
  description = "Friendly description"
  first_name = "John"
  last_name = "Doe"
  email = "john@doe.com"
  groups = ["4by6NvSLcPXFhCj8nwbZOM"]
}
