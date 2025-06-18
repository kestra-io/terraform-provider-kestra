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
				Config: testAccResourceBinding("GROUP", "admin", "admin_main", "namespace = \"io.kestra.terraform.data\"", "new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "type", "GROUP",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "external_id", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "role_id", "admin_main",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "namespace", "io.kestra.terraform.data",
					),
				),
			},
			{
				Config: testAccResourceBinding("USER", "john", "launcher_main", "", "new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "type", "USER",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "external_id", "john",
					),
					resource.TestCheckResourceAttr(
						"kestra_binding.new", "role_id", "launcher_default",
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

func testAccResourceBinding(resourceType, externalId, roleId, namespace, id string) string {
	return fmt.Sprintf(
		`
        resource "kestra_binding" "%s" {
            type = "%s"
            external_id = "%s"
			role_id = "%s"
			%s
        }`,
		id,
		resourceType,
		externalId,
		roleId,
		namespace,
	)

}
