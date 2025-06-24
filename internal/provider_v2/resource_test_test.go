package provider_v2

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccTestResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccResourceTestSuite("io.kestra.terraform.data", "simple-return-test-suite-1-id",
					`
id: simple-return-test-suite-1-id
namespace: io.kestra.terraform.data
description: assert flow is returning the input value as output
flowId: simple
testCases:
  - id: test_case_1
    type: io.kestra.core.tests.flow.UnitTest
    fixtures:
      inputs:
        inputA: "Hi there"
    assertions:
      - value: "{{ outputs.return.value }}"
        equalTo: 'Hi there'
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_test.new", "namespace", "io.kestra.terraform.data"),
					resource.TestCheckResourceAttr("kestra_test.new", "test_id", "simple-return-test-suite-1-id"),
				),
			},
		},
	})
}
func testAccResourceTestSuite(namespace string, testid string, content string) string {
	return fmt.Sprintf(
		`
        resource "kestra_test" "new" {
			namespace = "%s"
			test_id = "%s"
            content = <<EOT
%s
EOT
        }`,
		namespace,
		testid,
		content,
	)
}
