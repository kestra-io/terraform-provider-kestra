package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGroup(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGroup("admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_group.new", "group_id", regexp.MustCompile("admin"),
					),
				),
			},
		},
	})
}

func testAccDataSourceGroup(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_group" "new" {
            group_id = "%s"
        }`,
		id,
	)
}
