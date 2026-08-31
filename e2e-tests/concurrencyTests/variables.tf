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
  description = "Tenant the namespace is created in."
  type        = string
  default     = "main"
}
