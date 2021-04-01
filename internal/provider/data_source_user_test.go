package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUser("admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_user.new", "user_id", regexp.MustCompile("admin"),
					),
				),
			},
		},
	})
}

func testAccDataSourceUser(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_user" "new" {
            user_id = "%s"
        }`,
		id,
	)
}
