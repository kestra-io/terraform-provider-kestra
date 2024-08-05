resource "kestra_namespace_secret" "example" {
  namespace    = "company.team"
  secret_key   = "MY_KEY"
  secret_value = "my-r34l-53cr37"
}

