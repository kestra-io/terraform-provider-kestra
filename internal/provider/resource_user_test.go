package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUser(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUser(
					"admin@john.doe",
					"[kestra_group.group1.id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new", "username", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "email", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "groups.1", "[kestra_group.group1.id]",
					),
				),
			},
			{
				Config: testAccResourceUser(
					"admin2@john.doe",
					"[]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new", "username", "admin2@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "email", "admin2@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "groups", "admin2@john.doe",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_role.new", "groups.1",
					),
				),
			},
			{
				ResourceName:      "kestra_user.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceUser(email, groups string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "my user role"
			permissions {
			    type = "FLOW"
			    permissions = ["READ", "UPDATE"]
			}
        }

        resource "kestra_group" "group1" {
            name = "group 1"
        }

        resource "kestra_group" "group2" {
            name = "group 2"
        }

        resource "kestra_user" "new" {
            email = "%s"
            groups = %s
        }`,
		email,
		groups,
	)

}
