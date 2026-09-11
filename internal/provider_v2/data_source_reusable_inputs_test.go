package provider_v2

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceReusableInputs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheckReusableInputs(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceReusableInputsConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kestra_reusable_inputs.test", "reusable_inputs_id", "terraform-data-source-reusable-inputs"),
					resource.TestCheckResourceAttr("data.kestra_reusable_inputs.test", "namespace", "io.kestra.terraform.data"),
					resource.TestMatchResourceAttr("data.kestra_reusable_inputs.test", "content", regexp.MustCompile("read by terraform")),
					resource.TestCheckResourceAttrPair("data.kestra_reusable_inputs.test", "revision", "kestra_reusable_inputs.test", "revision"),
				),
			},
		},
	})
}

func testAccDataSourceReusableInputsConfig() string {
	return `
resource "kestra_reusable_inputs" "test" {
  reusable_inputs_id = "terraform-data-source-reusable-inputs"
  namespace          = "io.kestra.terraform.data"

  content = <<EOT
description: read by terraform
inputs:
  - id: environment
    type: STRING
EOT
}

data "kestra_reusable_inputs" "test" {
  reusable_inputs_id = kestra_reusable_inputs.test.reusable_inputs_id
  namespace          = kestra_reusable_inputs.test.namespace
}
`
}
