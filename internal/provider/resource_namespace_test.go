package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNamespace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNamespace(
					"io.kestra.terraform",
					"My Kestra Namespace",
					concat(
						"k1: 1",
						"k2:",
						"    v1: 1",
					),
					concat(
						"- type: io.kestra.core.tasks.log.Log",
						"  values:",
						"    message: first {{flow.id}}",
						"- type: io.kestra.core.tasks.debugs.Return",
						"  values:",
						"    format: first {{flow.id}}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "namespace_id", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "description", "My Kestra Namespace",
					),
				),
			},
			{
				Config: testAccResourceNamespace(
					"io.kestra.terraform",
					"My Kestra Namespace 2",
					concat(
						"k2:",
						"    v1: 1",
						"k1: 1",
					),
					concat(
						"- type: io.kestra.core.tasks.log.Log",
						"  values:",
						"    message: first {{flow.id}}",
						"- type: io.kestra.core.tasks.debugs.Return",
						"  values:",
						"    format: second {{flow.id}}",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "namespace_id", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "description", "My Kestra Namespace 2",
					),
					resource.TestMatchResourceAttr(
						"kestra_namespace.new", "plugin_defaults", regexp.MustCompile(".*format: second.*"),
					),
				),
			},
			{
				ResourceName:      "kestra_namespace.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceNamespace(id, description, variables, pluginDefaults string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace" "new" {
            namespace_id = "%s"
            description = "%s"
            variables = <<EOT
%s
EOT
            plugin_defaults = <<EOT
%s
EOT
        }`,
		id,
		description,
		variables,
		pluginDefaults,
	)
}
