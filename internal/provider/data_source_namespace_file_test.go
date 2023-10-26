package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNamespaceFile(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceNamespaceFile("io.kestra.terraform.data", "/flow.py"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_namespace_file.new", "destination_path", "/flow.py",
					),
				),
			},
		},
	})
}

func testAccDataSourceNamespaceFile(namespace, filename string) string {
	return fmt.Sprintf(
		`
        data "kestra_namespace_file" "new" {
            namespace = "%s"
            destination_path = "%s"
        }`,
		namespace,
		filename,
	)
}
