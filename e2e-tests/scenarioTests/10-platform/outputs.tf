output "namespace" {
  value = kestra_namespace.allowed.namespace_id
}

output "forbidden_namespace" {
  value = kestra_namespace.forbidden.namespace_id
}

output "flow_id" {
  value = kestra_flow.scenario.flow_id
}

output "forbidden_flow_id" {
  value = kestra_flow.forbidden.flow_id
}

output "testsuite_id" {
  value = kestra_test.scenario.test_id
}

output "kv_value" {
  description = "Expected tail of the flow output, so the assertion and the fixture cannot drift apart."
  value       = local.kv_value
}

output "app_user_id" {
  value = kestra_user.app_user.id
}

output "app_user_api_token" {
  description = "Credential scenarioTests/runtime acts as the app user with."
  value       = kestra_user_api_token.app_user.full_token
  sensitive   = true
}
