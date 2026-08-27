package provider_v2

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWorkerGroupDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkerGroup(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kestra_worker_group.result", "group_id", "tf-acc-group-ds"),
					resource.TestCheckResourceAttr("data.kestra_worker_group.result", "name", "Terraform Acceptance Group Data Source"),
					resource.TestCheckResourceAttr("data.kestra_worker_group.result", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("data.kestra_worker_group.result", "subscriptions.0.worker_queue_id", "default"),
				),
			},
		},
	})
}

func testAccDataSourceWorkerGroup() string {
	return `
        resource "kestra_worker_group" "new" {
            group_id = "tf-acc-group-ds"
            name     = "Terraform Acceptance Group Data Source"

            subscriptions {
                worker_queue_id = "default"
            }
        }

        data "kestra_worker_group" "result" {
            group_id = kestra_worker_group.new.group_id
        }`
}
