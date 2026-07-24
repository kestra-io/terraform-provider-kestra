package provider_v2

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkerGroupResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceWorkerGroup(
					"tf-acc-group",
					"Terraform Acceptance Group",
					"created by acceptance tests",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_worker_group.new", "group_id", "tf-acc-group"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "name", "Terraform Acceptance Group"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "description", "created by acceptance tests"),
				),
			},
			// Update with Worker Queue subscriptions
			{
				Config: testAccResourceWorkerGroupWithSubscriptions(
					"tf-acc-group",
					"Terraform Acceptance Group Updated",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_worker_group.new", "group_id", "tf-acc-group"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "name", "Terraform Acceptance Group Updated"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.#", "2"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.0.worker_queue_id", "default"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.0.reserved_percent", "-1"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.0.mode", "STRICT"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.1.worker_queue_id", "tf-acc-group-queue"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.1.reserved_percent", "50"),
					resource.TestCheckResourceAttr("kestra_worker_group.new", "subscriptions.1.mode", "ELASTIC"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "kestra_worker_group.new",
				ImportState:       true,
				ImportStateId:     "tf-acc-group",
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceWorkerGroup(groupId, name, description string) string {
	return fmt.Sprintf(
		`
        resource "kestra_worker_group" "new" {
            group_id    = "%s"
            name        = "%s"
            description = "%s"
        }`,
		groupId,
		name,
		description,
	)
}

func testAccResourceWorkerGroupWithSubscriptions(groupId, name string) string {
	return fmt.Sprintf(
		`
        resource "kestra_worker_queue" "queue" {
            queue_id = "%s-queue"
            tags     = ["tf-acc"]
        }

        resource "kestra_worker_group" "new" {
            group_id = "%s"
            name     = "%s"

            subscriptions {
                worker_queue_id = "default"
            }

            subscriptions {
                worker_queue_id  = kestra_worker_queue.queue.queue_id
                reserved_percent = 50
                mode             = "ELASTIC"
            }
        }`,
		groupId,
		groupId,
		name,
	)
}
