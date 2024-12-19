package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceApp(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceApp("new"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("kestra_app.new", "id"),
				),
			},
		},
	})
}

func testAccResourceApp(resourceId string) string {
	return fmt.Sprintf(
		`
        resource "kestra_flow" "new_flow" {
            namespace = "company.team"
            flow_id = "get_data"
            content = <<EOT
id: get_data
namespace: company.team

tasks:
- id: hello
  type: io.kestra.plugin.core.log.Log
  message: Hello World! 🚀
EOT
        }
        resource "kestra_app" "%s" {
			depends_on = [kestra_flow.new_flow]
            source = <<-EOF
id: test_tf
type: io.kestra.plugin.ee.apps.Execution
displayName: New display name
namespace: company.team
flowId: get_data
access: PRIVATE

layout:
  - on: OPEN
    blocks:
      - type: io.kestra.plugin.ee.apps.core.blocks.Markdown
        content: |
          ## Request data
          Select the dataset you want to download.
EOF
        }`,
		resourceId)
}
