package provider_v2

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
func TestUnitKestraTestLocals(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKestraTestWithLocals(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						`kestra_test.tests["solutions.team/tests/test-1.yaml"]`,
						tfjsonpath.New("namespace"),
						knownvalue.StringExact("solutions.team"),
					),
					statecheck.ExpectKnownValue(
						`kestra_test.tests["solutions.team/tests/test-1.yaml"]`,
						tfjsonpath.New("test_id"),
						knownvalue.StringExact("test-1"),
					),
					statecheck.ExpectKnownValue(
						`kestra_test.tests["solutions.team/tests/test-2.yaml"]`,
						tfjsonpath.New("namespace"),
						knownvalue.StringExact("solutions.team"),
					),
					statecheck.ExpectKnownValue(
						`kestra_test.tests["solutions.team/tests/test-2.yaml"]`,
						tfjsonpath.New("test_id"),
						knownvalue.StringExact("test-2"),
					),
				},
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccKestraTestWithLocals() string {
	return `
provider "kestra" {
  url = "http://localhost:8088"
}

locals {
  test_files = toset([
    "solutions.team/tests/test-1.yaml",
    "solutions.team/tests/test-2.yaml",
  ])

  unit_tests = {
    for f in local.test_files : f => {
      namespace = regex("^(.+)/tests", f)[0]
      test_id   = trimsuffix(basename(f), ".yaml")
      content   = join("\n", [
        format("id: %s", trimsuffix(basename(f), ".yaml")),
        format("namespace: %s", regex("^(.+)/tests", f)[0]),
        "testCases: []",
      ])
    }
  }
}

resource "kestra_test" "tests" {
  for_each  = local.unit_tests
  namespace = each.value.namespace
  test_id   = each.value.test_id
  content   = each.value.content
}`
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
