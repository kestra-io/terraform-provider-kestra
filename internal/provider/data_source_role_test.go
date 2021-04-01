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
