provider "kestra" {
  # mandatory, the url to kestra
  url = "http://localhost:8080"

  # optional basic auth username
  username = "john"

  # optional basic auth password
  password = "my-password"
}
