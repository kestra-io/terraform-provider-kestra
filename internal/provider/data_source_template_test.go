package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceTemplate(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceTemplate("io.kestra.terraform.data", "simple"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_template.new", "template_id", "simple",
					),
				),
			},
		},
	})
}

func testAccDataSourceTemplate(namespace, id string) string {
	return fmt.Sprintf(
		`
        data "kestra_template" "new" {
            namespace = "%s"
            template_id = "%s"
        }`,
		namespace,
		id,
	)
}
