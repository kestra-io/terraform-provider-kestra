resource "kestra_namespace_secret" "example" {
  namespace    = "io.kestra.mynamespace"
  secret_key   = "MY_KEY"
  secret_value = "my-r34l-53cr37"
}

