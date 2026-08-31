output "tenant_id" {
  description = "Tenant the scenario is built in."
  value       = var.tenant_id
}

output "platform_admin_api_token" {
  description = "Credential stage 10 authenticates with."
  value       = kestra_user_api_token.platform_admin.full_token
  sensitive   = true
}

output "app_user_id" {
  description = "Passed to stage 10, which puts this user in the launcher group."
  value       = kestra_user.app_user.id
}

output "app_user_api_token" {
  description = "Credential scenarioTests/runtime acts as the app user with."
  value       = kestra_user_api_token.app_user.full_token
  sensitive   = true
}
