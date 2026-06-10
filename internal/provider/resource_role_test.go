package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccRole(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceRole(
					"admin",
					"My admin role",
					concat(
						"resources {",
						"  type = \"FLOW\"",
						"  actions = [\"VIEW\", \"LIST\", \"UPDATE\"]",
						"}",
						"resources {",
						"  type = \"NAMESPACE\"",
						"  actions = [\"VIEW\", \"LIST\"]",
						"}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_role.new", "name", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_role.new", "description", "My admin role",
					),
				),
			},
			{
				Config: testAccResourceRole(
					"admin 2",
					"My admin role 2",
					concat(
						"resources {",
						"  type = \"FLOW\"",
						"  actions = [\"VIEW\", \"LIST\"]",
						"}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_role.new", "name", "admin 2",
					),
					resource.TestCheckResourceAttr(
						"kestra_role.new", "description", "My admin role 2",
					),
				),
			},
			{
				ResourceName:      "kestra_role.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceRole(name, description, resources string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "%s"
            description = "%s"
            %s
            is_default = false
        }`,
		name,
		description,
		resources,
	)
}
