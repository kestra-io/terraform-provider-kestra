package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTemplate(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTemplate(
					"io.kestra.terraform",
					"simple",
					concat(
						"tasks:",
						"  - id: t1",
						"    type: io.kestra.core.tasks.log.Log",
						"    message: first {{task.id}}",
						"    level: TRACE",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_template.new", "id", "io.kestra.terraform/simple",
					),
					resource.TestCheckResourceAttr(
						"kestra_template.new", "namespace", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_template.new", "template_id", "simple",
					),
				),
			},
			{
				Config: testAccResourceTemplate(
					"io.kestra.terraform",
					"simple",
					concat(
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.core.tasks.log.Log",
						"    message: first {{task.id}}",
						"    level: TRACE",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_template.new", "content", regexp.MustCompile(".*id: t2\n.*"),
					),
				),
			},
			{
				ResourceName:      "kestra_template.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceTemplate(id, name, content string) string {
	return fmt.Sprintf(
		`
        resource "kestra_template" "new" {
            namespace = "%s"
            template_id = "%s"
            content = <<EOT
%s
EOT
        }`,
		id,
		name,
		content,
	)
}
