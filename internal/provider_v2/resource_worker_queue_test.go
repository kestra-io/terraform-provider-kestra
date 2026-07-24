package provider_v2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkerQueueResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceWorkerQueue(
					"tf-acc-queue",
					"Terraform Acceptance Queue",
					"created by acceptance tests",
					`["gpu", "high-memory"]`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "queue_id", "tf-acc-queue"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "name", "Terraform Acceptance Queue"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "description", "created by acceptance tests"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr("kestra_worker_queue.new", "tags.*", "gpu"),
					resource.TestCheckTypeSetElemAttr("kestra_worker_queue.new", "tags.*", "high-memory"),
				),
			},
			// Update and Read testing
			{
				Config: testAccResourceWorkerQueue(
					"tf-acc-queue",
					"Terraform Acceptance Queue Updated",
					"updated by acceptance tests",
					`["gpu"]`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "queue_id", "tf-acc-queue"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "name", "Terraform Acceptance Queue Updated"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "description", "updated by acceptance tests"),
					resource.TestCheckResourceAttr("kestra_worker_queue.new", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("kestra_worker_queue.new", "tags.*", "gpu"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kestra_worker_queue.new",
				ImportState:       true,
				ImportStateId:     "tf-acc-queue",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceWorkerQueue(queueId, name, description, tags string) string {
	return fmt.Sprintf(
		`
        resource "kestra_worker_queue" "new" {
            queue_id    = "%s"
            name        = "%s"
            description = "%s"
            tags        = %s
        }`,
		queueId,
		name,
		description,
		tags,
	)
}
