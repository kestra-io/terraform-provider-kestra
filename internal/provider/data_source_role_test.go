package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceRole(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRole("admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_role.new", "role_id", regexp.MustCompile("admin"),
					),
				),
			},
		},
	})
}

func testAccDataSourceRole(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_role" "new" {
            role_id = "%s"
        }`,
		id,
	)
}

func TestAccDataSourceRoleByName(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoleByName("data-source-role-lookup"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.kestra_role.by_name", "role_id",
						"kestra_role.name_lookup", "id",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_role.by_name", "name", "data-source-role-lookup",
					),
				),
			},
		},
	})
}

func testAccDataSourceRoleByName(name string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "name_lookup" {
            name = "%s"
            resources {
                type = "FLOW"
                actions = ["VIEW", "LIST"]
            }
            is_default = false
        }

        data "kestra_role" "by_name" {
            name = kestra_role.name_lookup.name
        }`,
		name,
	)
}
