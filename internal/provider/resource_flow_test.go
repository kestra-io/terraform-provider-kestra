package provider

import (
	"fmt"
	"path/filepath"
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
				PreConfig: func() {
					t1, _ := filepath.Abs("../resources/t1.yml")
					copyResource(t1, "t1.yml")

					flow, _ := filepath.Abs("../resources/flow.yml")
					copyResource(flow, "flow.yml")

					python, _ := filepath.Abs("../resources/flow.py")
					copyResource(python, "flow.py")

					bigint, _ := filepath.Abs("../resources/bigint.yml")
					copyResource(bigint, "bigint.yml")
				},
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"id: simple",
						"namespace: io.kestra.terraform",
						"revision: 13",
						"tasks:",
						"  - ${indent(4, file(\"/tmp/unit-test/t1.yml\"))}",
						"taskDefaults:",
						"  - type: io.kestra.core.tasks.debugs.Echo",
						"    values:",
						"      format: third {{flow.id}}",
						"inputs:",
						"  - name: my-value",
						"    type: STRING",
						"variables:",
						"  first: \"1\"",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_flow.new", "id", "io.kestra.terraform/simple",
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
			{
				Config: concat(
					"resource \"kestra_flow\" \"template\" {",
					"  namespace = \"io.kestra.terraform\"",
					"  flow_id = \"template\"",
					"  content = templatefile(\"/tmp/unit-test/flow.yml\", {})",
					"}",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_flow.template", "id", "io.kestra.terraform/template",
					),
				),
			},
			{
				Config: concat(
					"resource \"kestra_flow\" \"bigint\" {",
					"  namespace = \"io.kestra.terraform\"",
					"  flow_id = \"bigint\"",
					"  content = templatefile(\"/tmp/unit-test/bigint.yml\", {})",
					"}",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_flow.bigint", "content", regexp.MustCompile("(?s).*3600000.*"),
					),
					resource.TestCheckResourceAttr(
						"kestra_flow.bigint", "id", "io.kestra.terraform/bigint",
					),
				),
			},
		},
	})
}

func TestAccIncohrenceResourceFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"id: \"wrong-id\"",
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.core.tasks.debugs.Echo",
						"    format: first {{task.id}}",
						"    level: TRACE",
					),
				),
				ExpectError: regexp.MustCompile(".*incoherent resource id: simple.*"),
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
