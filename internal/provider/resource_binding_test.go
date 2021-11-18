package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBinding(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceBinding(
					"GROUP",
					"admin",
					"admin",
					"namespace = \"io.kestra.terraform.data\"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "type", "GROUP",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "external_id", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "role_id", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "namespace", "io.kestra.terraform.data",
					),
				),
			},
			{
				Config: testAccResourceBinding(
					"USER",
					"john",
					"admin",
					"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "type", "USER",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "external_id", "john",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "role_id", "admin",
					),
				),
			},
			{
				ResourceName:            "kestra_binding.new",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"namespace"},
			},
		},
	})
}

func testAccResourceBinding(resourceType, externalId, roleId, namespace string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "my binding role"
        }

        resource "kestra_role" "new2" {
            name = "my binding role 2"
        }

        resource "kestra_binding" "new" {
            type = "%s"
            external_id = "%s"
			role_id = "%s"
			%s
        }`,
		resourceType,
		externalId,
		roleId,
		namespace,
	)

}
