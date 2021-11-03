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
					"namespace = \"io.kestra\"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_group.new", "name", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "description", "My admin group",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "namespace", "io.kestra",
					),
				),
			},
			{
				Config: testAccResourceGroup(
					"admin 2",
					"My admin group 2",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_group.new", "name", "admin 2",
					),
					resource.TestCheckResourceAttr(
						"kestra_group.new", "description", "My admin group 2",
					),
				),
			},
			{
				ResourceName:            "kestra_group.new",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"namespace"},
			},
		},
	})
}

func testAccResourceGroup(name, description, namespace string) string {
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
			%s
        }`,
		name,
		description,
		namespace,
	)

}
