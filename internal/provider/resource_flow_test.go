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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			// TODO these test don't work well on macos
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

					sourceyaml, _ := filepath.Abs("../resources/source_yaml.yml")
					copyResource(sourceyaml, "source_yaml.yml")

					sourceyaml2, _ := filepath.Abs("../resources/source_yaml_2.yml")
					copyResource(sourceyaml2, "source_yaml_2.yml")
				},
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"id: simple",
						"namespace: io.kestra.terraform",
						"tasks:",
						"  - ${indent(4, file(\"/tmp/unit-test/t1.yml\"))}",
						"inputs:",
						"  - id: my-value",
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
			{
				Config: concat(
					"resource \"kestra_flow\" \"yaml_source\" {",
					"  namespace = \"io.kestra.terraform\"",
					"  flow_id = \"yaml_source\"",
					"  content = templatefile(\"/tmp/unit-test/source_yaml.yml\", {})",
					"}",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_flow.yaml_source", "content", regexp.MustCompile("# yaml source code must be kept"),
					),
					resource.TestMatchResourceAttr(
						"kestra_flow.yaml_source", "content", regexp.MustCompile("# even inside task"),
					),
				),
			},
			{
				Config: concat(
					"resource \"kestra_flow\" \"yaml_source\" {",
					"  namespace = \"io.kestra.terraform\"",
					"  flow_id = \"yaml_source\"",
					"  content = templatefile(\"/tmp/unit-test/source_yaml_2.yml\", {})",
					"}",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"kestra_flow.yaml_source", "content", regexp.MustCompile("# yaml source code must be kept"),
					),
					resource.TestMatchResourceAttr(
						"kestra_flow.yaml_source", "content", regexp.MustCompile("# only comment"),
					),
				),
			},
		},
	})
}

func TestAccIncohrenceResourceFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"id: \"wrong-id\"",
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.plugin.core.log.Log",
						"    message: first {{task.id}}",
						"    level: TRACE",
					),
				),
				ExpectError: regexp.MustCompile("Inconsistent flow identity"),
			},
		},
	})
}

func TestAccTenantResourceFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "kestra" { tenant_id = "unit_test" }` + testAccResourceFlow(
					"io.kestra.terraform",
					"simple",
					concat(
						"id: simple",
						"namespace: io.kestra.terraform",
						"tasks:",
						"  - id: t2",
						"    type: io.kestra.plugin.core.log.Log",
						"    message: first {{task.id}}",
						"    level: TRACE",
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
						"kestra_flow.new", "tenant_id", "unit_test",
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

// TestAccResourceFlowMetadata covers disabled, description and labels managed as attributes:
// each apply must leave a clean plan, an attribute may not repeat a key of content, and
// removing the attributes hands the keys back to content.
func TestAccResourceFlowMetadata(t *testing.T) {
	content := concat(
		"id: metadata",
		"namespace: io.kestra.terraform",
		"# comments in content are kept",
		"tasks:",
		"  - id: log",
		"    type: io.kestra.plugin.core.log.Log",
		"    message: hello",
	)
	config := func(attributes, content string) string {
		return fmt.Sprintf(`
resource "kestra_flow" "metadata" {
  namespace = "io.kestra.terraform"
  flow_id   = "metadata"
  %s
  content   = <<EOT
%s
EOT
}`, attributes, content)
	}

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config(`
  disabled    = true
  description = "Managed by Terraform"
  labels      = { team = "data" }`, content),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_flow.metadata", "disabled", "true"),
					resource.TestCheckResourceAttr("kestra_flow.metadata", "description", "Managed by Terraform"),
					resource.TestCheckResourceAttr("kestra_flow.metadata", "labels.team", "data"),
					resource.TestCheckResourceAttr("kestra_flow.metadata", "content", content+"\n"),
				),
			},
			{
				Config: config(`
  disabled    = false
  description = ""
  labels      = {}`, content),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_flow.metadata", "disabled", "false"),
					resource.TestCheckResourceAttr("kestra_flow.metadata", "description", ""),
					resource.TestCheckResourceAttr("kestra_flow.metadata", "labels.%", "0"),
				),
			},
			{
				Config:      config(`description = "Managed by Terraform"`, content+"\ndescription: Managed by Terraform"),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Conflicting flow metadata"),
			},
			{
				Config: config("", content+"\ndescription: Managed in YAML"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("kestra_flow.metadata", "disabled"),
					resource.TestCheckNoResourceAttr("kestra_flow.metadata", "description"),
					resource.TestCheckNoResourceAttr("kestra_flow.metadata", "labels.%"),
					resource.TestMatchResourceAttr("kestra_flow.metadata", "content", regexp.MustCompile("description: Managed in YAML")),
				),
			},
			{
				ResourceName:      "kestra_flow.metadata",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccResourceFlowMetadataConflictAtApply covers a content known only at apply. Terraform
// validates the configuration again once the content is known, so the duplicate key still
// fails before anything is written, also for an attribute set to an empty value.
func TestAccResourceFlowMetadataConflictAtApply(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "content" {
  input = "id: metadata-apply\nnamespace: io.kestra.terraform\nlabels:\n  team: data\ntasks:\n  - id: log\n    type: io.kestra.plugin.core.log.Log\n    message: hello\n"
}

resource "kestra_flow" "conflict" {
  labels  = {}
  content = terraform_data.content.output
}`,
				ExpectError: regexp.MustCompile("`labels` is set both as an attribute and in content"),
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
