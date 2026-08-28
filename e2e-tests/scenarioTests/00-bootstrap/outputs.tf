output "tenant_id" {
  description = "Tenant the scenario is built in."
  value       = var.tenant_id
}

output "platform_admin_user_id" {
  value = kestra_user.platform_admin.id
}

output "platform_admin_api_token" {
  description = "Credential stage 10 authenticates with."
  value       = kestra_user_api_token.platform_admin.full_token
  sensitive   = true
}
