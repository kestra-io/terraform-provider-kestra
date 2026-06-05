package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceWorkerGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkerGroup("1RgkLgU0oUXndtPswzaFku", "WorkerGroupKey-1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_worker_group.new", "id", regexp.MustCompile("1RgkLgU0oUXndtPswzaFku"),
					),
					resource.TestMatchResourceAttr(
						"data.kestra_worker_group.new", "key", regexp.MustCompile("WorkerGroupKey-1"),
					),
				),
			},
		},
	})
}

func testAccDataSourceWorkerGroup(id string, key string) string {
	return fmt.Sprintf(
		`
        data "kestra_worker_group" "new" {
            id = "%s"
            key = "%s"
        }`,
		id,
		key,
	)
}
