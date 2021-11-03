package provider

import (
	"fmt"
	"regexp"
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
					"admin",
					"My admin user",
					"[kestra_group.group1.id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new", "username", "admin",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "description", "My admin user",
					),
					resource.TestMatchResourceAttr(
						"kestra_user.new", "groups.1", regexp.MustCompile(".*"),
					),
				),
			},
			{
				Config: testAccResourceUser(
					"admin 2",
					"My admin user 2",
					"[]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new", "username", "admin 2",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new", "description", "My admin user 2",
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

func testAccResourceUser(name, description, groups string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "my user role"
        }

        resource "kestra_group" "group1" {
            name = "group 1"
        }

        resource "kestra_group" "group2" {
            name = "group 2"
        }

        resource "kestra_user" "new" {
            username = "%s"
            description = "%s"
            groups = %s
        }`,
		name,
		description,
		groups,
	)

}
