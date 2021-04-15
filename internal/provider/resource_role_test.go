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
						"permissions {",
						"  type = \"FLOW\"",
						"  permissions = [\"READ\", \"UPDATE\"]",
						"}",
						"permissions {",
						"  type = \"TEMPLATE\"",
						"  permissions = [\"READ\", \"UPDATE\"]",
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
					resource.TestCheckResourceAttr(
						"kestra_role.new", "permissions.0.type", "FLOW",
					),
					resource.TestCheckResourceAttr(
						"kestra_role.new", "permissions.0.permissions.0", "READ",
					),
					resource.TestCheckResourceAttr(
						"kestra_role.new", "permissions.0.permissions.1", "UPDATE",
					),
				),
			},
			{
				Config: testAccResourceRole(
					"admin 2",
					"My admin role 2",
					concat(
						"permissions {",
						"  type = \"FLOW\"",
						"  permissions = [\"READ\"]",
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
					resource.TestCheckResourceAttr(
						"kestra_role.new", "permissions.0.type", "FLOW",
					),
					resource.TestCheckResourceAttr(
						"kestra_role.new", "permissions.0.permissions.0", "READ",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_role.new", "permissions.1.permissions.0",
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

func testAccResourceRole(name, description, permissions string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "%s"
            description = "%s"
            %s
        }`,
		name,
		description,
		permissions,
	)
}
