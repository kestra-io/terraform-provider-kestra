package provider

import (
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
				Config: testAccResourceFlow,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_flow.foo", "namespace", regexp.MustCompile("^ba")),
				),
			},
		},
	})
}

const testAccResourceFlow = `
resource "kestra_flow" "foo" {
  namespace = "io.kestra.tests"
  flow_id = "logs"
  content = <<EOT
taskDefaults:
  - type: io.kestra.core.tasks.debugs.Echo
	values:
	  format: third {{flow.id}}

tasks:
- id: t1
  type: io.kestra.core.tasks.debugs.Echo
  format: first {{task.id}}
  level: TRACE
- id: t2
  type: io.kestra.core.tasks.debugs.Echo
  format: second {{task.type}}
  level: WARN
- id: t3
  type: io.kestra.core.tasks.debugs.Echo
  format: third {{flow.id}}
  level: ERROR
EOT
}
`
