package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNamespace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNamespace("io.kestra.terraform.data"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_namespace.new", "namespace_id", regexp.MustCompile("io\\.kestra"),
					),
					resource.TestMatchResourceAttr(
						"data.kestra_namespace.new", "name", regexp.MustCompile("My Kestra Namespace"),
					),
				),
			},
		},
	})
}

func testAccDataSourceNamespace(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_namespace" "new" {
            namespace_id = "%s"
        }`,
		id,
	)
}
