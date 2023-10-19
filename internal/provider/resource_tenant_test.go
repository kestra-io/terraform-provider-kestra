package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenant(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTenant(
					"custom",
					"My custom tenant",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_tenant.new", "tenant_id", "custom",
					),
					resource.TestCheckResourceAttr(
						"kestra_tenant.new", "name", "My custom tenant",
					),
				),
			},
			{
				ResourceName:      "kestra_tenant.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceTenant(id, name string) string {
	return fmt.Sprintf(
		`
        resource "kestra_tenant" "new" {
            tenant_id = "%s"
            name = "%s"
        }`,
		id,
		name,
	)
}
