package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccWorkerGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceWorkerGroup(
					"test-key",
					"Test Worker Group",
					"test-description",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "key", "test-key",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "name", "Test Worker Group",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "description", "test-description",
					),
				),
			},
			{
				Config: testAccResourceWorkerGroup(
					"test-key",
					"Test Worker Group Updated",
					"test-description-2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "key", "test-key",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "name", "Test Worker Group Updated",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "description", "test-description-2",
					),
				),
			},
			{
				ResourceName:      "kestra_worker_group.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceWorkerGroup(key, name, description string) string {
	return fmt.Sprintf(
		`
        resource "kestra_worker_group" "new" {
            key = "%s"
            name = "%s"
            description = "%s"
        }`,
		key,
		name,
		description,
	)

}
