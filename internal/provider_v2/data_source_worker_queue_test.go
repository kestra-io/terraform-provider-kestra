package provider_v2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkerQueueDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkerQueue(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kestra_worker_queue.result", "queue_id", "tf-acc-queue-ds"),
					resource.TestCheckResourceAttr("data.kestra_worker_queue.result", "name", "Terraform Acceptance Queue Data Source"),
					resource.TestCheckResourceAttr("data.kestra_worker_queue.result", "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr("data.kestra_worker_queue.result", "tags.*", "datasource"),
				),
			},
		},
	})
}

func testAccDataSourceWorkerQueue() string {
	return `
        resource "kestra_worker_queue" "new" {
            queue_id = "tf-acc-queue-ds"
            name     = "Terraform Acceptance Queue Data Source"
            tags     = ["datasource"]
        }

        data "kestra_worker_queue" "result" {
            queue_id = kestra_worker_queue.new.queue_id
        }`
}
