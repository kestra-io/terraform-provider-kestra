package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWorkerGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkerGroup("worker-group-data-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_worker_group.new", "id", "worker-group-data-1",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_worker_group.new", "key", "worker-group-data-1",
					),
				),
			},
		},
	})
}

func testAccDataSourceWorkerGroup(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_worker_group" "new" {
            id = "%s"
        }`,
		id,
	)
}
