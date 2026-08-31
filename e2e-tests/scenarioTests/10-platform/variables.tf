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

variable "app_user_id" {
  description = "Id of the app user created by scenarioTests/00-bootstrap."
  type        = string
}
