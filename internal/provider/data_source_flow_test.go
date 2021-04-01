package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFlow("io.kestra.terraform.data", "simple"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_flow.new", "flow_id", "simple",
					),
				),
			},
		},
	})
}

func testAccDataSourceFlow(namespace, id string) string {
	return fmt.Sprintf(
		`
        data "kestra_flow" "new" {
            namespace = "%s"
            flow_id = "%s"
        }`,
		namespace,
		id,
	)
}
