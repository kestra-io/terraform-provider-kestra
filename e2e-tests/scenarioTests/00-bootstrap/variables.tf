variable "kestra_url" {
  description = "The Kestra webserver/standalone URL."
  type        = string
  default     = "http://localhost:8088"
}

variable "kestra_username" {
  description = "The super-admin declared in kestra.security.super-admin."
  type        = string
  default     = "root@root.com"
}

variable "kestra_password" {
  description = "The super-admin password."
  type        = string
  default     = "Root!1234"
  sensitive   = true
}

variable "tenant_id" {
  description = "Tenant the scenario is built in."
  type        = string
  default     = "main"
}

variable "platform_admin_password" {
  description = "Basic-auth password set on the platform admin, so the account is usable in the UI when debugging a failed run."
  type        = string
  default     = "E2ePlatformAdmin!1234"
  sensitive   = true
}
