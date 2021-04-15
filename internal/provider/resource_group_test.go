package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceGroup(
					"admin",
					"My admin group",
					"[kestra_role.new.id]",
					concat(
						"namespace_roles {",
						"  namespace = \"io.kestra.terraform.space1\"",
						"  roles = kestra_role.new.id",
						"}",
						"namespace_roles {",
						"  namespace = \"io.kestra.terraform.space2\"",
						"  roles = kestra_role.new.id",
						"}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_group.new", "name", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "description", "My admin group",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "namespace_roles.1.namespace", "io.kestra.terraform.space2",
					),
				),
			},
			{
				Config: testAccResourceGroup(
					"admin 2",
					"My admin group 2",
					"[]",
					concat(
						"namespace_roles {",
						"  namespace = \"io.kestra.terraform.space1\"",
						"  roles = kestra_role.new.id",
						"}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_group.new", "name", "admin 2",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "description", "My admin group 2",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_group.new", "namespace_roles.1",
					),
				),
			},
			{
				ResourceName:      "kestra_group.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceGroup(name, description, globalRoles, namespaceRoles string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "my group role"
        }

        resource "kestra_role" "new2" {
            name = "my group role 2"
        }

        resource "kestra_group" "new" {
            name = "%s"
            description = "%s"
            global_roles = %s
            %s
        }`,
		name,
		description,
		globalRoles,
		namespaceRoles,
	)

}
