variable "kestra_url" {
  description = "The Kestra webserver/standalone URL."
  type        = string
  default     = "http://localhost:8088"
}

variable "tenant_id" {
  description = "Tenant the scenario is built in."
  type        = string
  default     = "main"
}

variable "platform_admin_api_token" {
  description = "API token of the platform admin created by scenarioTests/00-bootstrap."
  type        = string
  sensitive   = true
}

variable "app_user_password" {
  description = "Basic-auth password set on the app user, so the account is usable in the UI when debugging a failed run."
  type        = string
  default     = "E2eAppUser!1234"
  sensitive   = true
}
