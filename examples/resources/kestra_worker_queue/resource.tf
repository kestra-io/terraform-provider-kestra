resource "kestra_worker_queue" "example" {
  queue_id    = "gpu-queue"
  name        = "GPU Queue"
  description = "Routes GPU workloads to dedicated workers"
  tags        = ["gpu", "high-memory"]

  # Optional: restrict the queue to specific tenants
  allowed_tenants = ["production"]
}
