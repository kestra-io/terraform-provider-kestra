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
					"test-description",
					"[\"test-tenant\"]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "key", "test-key",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "description", "test-description",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "allowed_tenants.#", "1",
					),
				),
			},
			{
				Config: testAccResourceWorkerGroup(
					"test-key",
					"test-description",
					"[]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "key", "test-key",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "description", "test-description",
					),
					resource.TestCheckResourceAttr(
						"kestra_worker_group.new", "allowed_tenants.#", "0",
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

func testAccResourceWorkerGroup(key, description, tenants string) string {
	return fmt.Sprintf(
		`
        resource "kestra_worker_group" "new" {
            key = "%s"
            description = "%s"
			allowed_tenants = %s
        }`,
		key,
		description,
		tenants,
	)

}
