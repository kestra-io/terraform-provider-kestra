package provider_v2

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourcePolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePolicyConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kestra_policy.tenant", "scope", "TENANT"),
					resource.TestCheckResourceAttr("data.kestra_policy.tenant", "policy_id", "terraform-data-source-policy"),
					resource.TestMatchResourceAttr("data.kestra_policy.tenant", "content", regexp.MustCompile("read by terraform")),
				),
			},
		},
	})
}

func TestUnitDataSourcePolicyValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError string
	}{
		{
			name: "namespace on tenant scope",
			config: `
data "kestra_policy" "invalid" {
  scope     = "TENANT"
  policy_id = "some-policy"
  namespace = "company.team"
}`,
			expectError: `namespace must not be set for a TENANT scope policy`,
		},
		{
			name: "tenant_id on instance scope",
			config: `
data "kestra_policy" "invalid" {
  scope     = "INSTANCE"
  policy_id = "some-policy"
  tenant_id = "main"
}`,
			expectError: `tenant_id must not be set for an INSTANCE scope policy`,
		},
		{
			name: "missing namespace on namespace scope",
			config: `
data "kestra_policy" "invalid" {
  scope     = "NAMESPACE"
  policy_id = "some-policy"
}`,
			expectError: `namespace is required for a NAMESPACE scope policy`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      testUnitProviderConfig + tt.config,
						ExpectError: regexp.MustCompile(tt.expectError),
					},
				},
			})
		})
	}
}

func testAccDataSourcePolicyConfig() string {
	return `
resource "kestra_policy" "tenant" {
  scope     = "TENANT"
  policy_id = "terraform-data-source-policy"

  content = <<EOT
id: terraform-data-source-policy
description: read by terraform
rules:
  - type: io.kestra.plugin.ee.rules.Deny
    on: PLUGIN
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.core.log.Log
EOT
}

data "kestra_policy" "tenant" {
  scope     = "TENANT"
  policy_id = kestra_policy.tenant.policy_id
}
`
}
