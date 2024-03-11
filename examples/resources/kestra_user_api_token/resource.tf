resource "kestra_user_api_token" "example" {
  user_id = "4by6NvSLcPXFhCj8nwbZOM"

  name = "test-token"
  description = "Test token"
  max_age = "PT1H"
  extended = false
}
