package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceBinding(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceBinding("john"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_binding.new", "binding_id", "john",
					),
				),
			},
		},
	})
}

func testAccDataSourceBinding(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_binding" "new" {
            binding_id = "%s"
        }`,
		id,
	)
}
