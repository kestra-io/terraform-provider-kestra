resource "kestra_worker_queue" "gpu" {
  queue_id = "gpu-queue"
  tags     = ["gpu"]
}

resource "kestra_worker_group" "example" {
  group_id    = "gpu-workers"
  name        = "GPU Workers"
  description = "Worker group dedicated to GPU workloads"

  # Subscribe to the global default queue with no reservation
  subscriptions {
    worker_queue_id = "default"
  }

  # Reserve 50% of each worker's slots for the GPU queue
  subscriptions {
    worker_queue_id  = kestra_worker_queue.gpu.queue_id
    reserved_percent = 50
    mode             = "ELASTIC"
  }
}
