package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"tasks:",
						"  - id: t1",
						"    type: io.kestra.core.tasks.debugs.Echo",
						"    format: first {{task.id}}",
						"    level: TRACE",
						"taskDefaults:",
						"  - type: io.kestra.core.tasks.debugs.Echo",
						"    values:",
						"      format: third {{flow.id}}",
						"inputs:",
						"  - name: my-value",
						"    type: STRING",
						"    required: true",
						"variables:",
						"  first: \"1\"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_flow.new", "id", "io.kestra.terraform_simple",
					),
					resource.TestCheckResourceAttr(
						"kestra_flow.new", "namespace", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_flow.new", "flow_id", "simple",
					),
				),
			},
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.core.tasks.debugs.Echo",
						"    format: first {{task.id}}",
						"    level: TRACE",
						"taskDefaults:",
						"  - type: io.kestra.core.tasks.debugs.Echo",
						"    values:",
						"      format: third {{flow.id}}",
						"inputs:",
						"  - name: my-value",
						"    type: STRING",
						"    required: true",
						"variables:",
						"  first: \"1\"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_flow.new", "content", regexp.MustCompile(".*id: t2\n.*"),
					),
				),
			},
			{
				ResourceName:      "kestra_flow.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceFlow(id, name, content string) string {
	return fmt.Sprintf(
		`
        resource "kestra_flow" "new" {
            namespace = "%s"
            flow_id = "%s"
            content = <<EOT
%s
EOT
        }`,
		id,
		name,
		content,
	)
}
