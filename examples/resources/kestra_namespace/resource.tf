resource "kestra_worker_queue" "gpu" {
  queue_id = "gpu"
  tags     = ["gpu", "linux"]
}

resource "kestra_namespace" "example" {
  namespace_id = "company.team"
  description  = "Friendly description"
  variables    = <<EOT
k1: 1
k2:
    v1: 1
EOT

  # route every task of the namespace to a Worker Queue matching these tags
  default_worker_selector {
    tags     = kestra_worker_queue.gpu.tags
    match    = "ALL"
    fallback = "WAIT"
  }
}
