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
					"new_user_one_group",
					"admin@john.doe",
					"[kestra_group.group1.id]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "username", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "email", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "groups.#", "1", // counting groups
					),
				),
			},
			{
				Config: testAccResourceUser(
					"new_user_one_group",
					"admin@john.doe",
					"[]", // removing the user group
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "username", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "email", "admin@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_one_group", "groups.#", "0",
					),
				),
			},
			{
				Config: testAccResourceUser(
					"new_user_no_group",
					"admin2@john.doe",
					"[]",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_no_group", "username", "admin2@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_no_group", "email", "admin2@john.doe",
					),
					resource.TestCheckResourceAttr(
						"kestra_user.new_user_no_group", "groups.#", "0",
					),
				),
			},
			{
				ResourceName:      "kestra_user.new_user_no_group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceUser(tfstateid, email, groups string) string {
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

        resource "kestra_user" "%s" {
            email = "%s"
            groups = %s
        }`,
		tfstateid,
		email,
		groups,
	)

}
