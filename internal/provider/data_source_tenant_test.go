package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTenant(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTenant("admin"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_tenant.new", "id", regexp.MustCompile("admin"),
					),
				),
			},
		},
	})
}

func testAccDataSourceTenant(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_tenant" "new" {
            tenant_id = "%s"
        }`,
		id,
	)
}
