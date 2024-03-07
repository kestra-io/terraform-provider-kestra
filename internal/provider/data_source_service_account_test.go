package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceServiceAccount(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceServiceAccount("2EPi5XC0oluKRCVF56gcC"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_service_account.new", "id", regexp.MustCompile("2EPi5XC0oluKRCVF56gcC"),
					),
				),
			},
		},
	})
}

func testAccDataSourceServiceAccount(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_service_account" "new" {
            id = "%s"
        }`,
		id,
	)
}
