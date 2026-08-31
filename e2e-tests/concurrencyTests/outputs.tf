output "namespace" {
  description = "Namespace the drift assertion mutates out of band."
  value       = kestra_namespace.concurrency.namespace_id
}

output "concurrency_limit" {
  description = "The configured limit, so the drift assertion can pick a different one."
  value       = one(kestra_namespace.concurrency.concurrency).limit
}
