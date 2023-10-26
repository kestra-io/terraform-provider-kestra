package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNamespaceFile(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNamespaceFile(
					"io.kestra.terraform",
					"/path/simple.yml",
					concat(
						"tasks:",
						"  - id: t1",
						"    type: io.kestra.core.tasks.debugs.Echo",
						"    format: first {{task.id}}",
						"    level: TRACE",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace_file.new", "id", "io.kestra.terraform//path/simple.yml",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace_file.new", "destination_path", "io.kestra.terraform",
					),
					resource.TestMatchResourceAttr(
						"kestra_namespace_file.new", "content", regexp.MustCompile(".*id: t1\n.*"),
					),
				),
			},
			{
				Config: testAccResourceNamespaceFile(
					"io.kestra.terraform",
					"/path/simple.yml",
					concat(
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.core.tasks.debugs.Echo",
						"    format: first {{task.id}}",
						"    level: TRACE",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_namespace_file.new", "content", regexp.MustCompile(".*id: t2\n.*"),
					),
				),
			},
			{
				ResourceName:      "kestra_namespace_file.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceNamespaceFile(namespace, fileName, content string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace_file" "new" {
            namespace = "%s"
			destination_path = "%s"
            content = <<EOT
%s
EOT
        }`,
		namespace,
		fileName,
		content,
	)
}
