package provider_v2

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccPolicyResourceTenant(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPolicyTenant("terraform-tenant-policy", "created by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_policy.tenant", "scope", "TENANT"),
					resource.TestCheckResourceAttr("kestra_policy.tenant", "policy_id", "terraform-tenant-policy"),
					resource.TestCheckResourceAttrSet("kestra_policy.tenant", "tenant_id"),
				),
			},
			// Update and Read testing
			{
				Config: testAccPolicyTenant("terraform-tenant-policy", "updated by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_policy.tenant", "policy_id", "terraform-tenant-policy"),
					resource.TestMatchResourceAttr("kestra_policy.tenant", "content", regexp.MustCompile("updated by terraform")),
				),
			},
			// ImportState testing
			{
				ResourceName: "kestra_policy.tenant",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["kestra_policy.tenant"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("TENANT/%s/%s", rs.Primary.Attributes["tenant_id"], rs.Primary.Attributes["policy_id"]), nil
				},
				// the API round-trips the source verbatim, so even content matches;
				// the resource has no `id` attribute, so match on policy_id
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "policy_id",
			},
		},
	})
}

func TestAccPolicyResourceNamespace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyNamespace("terraform-namespace-policy", "io.kestra.terraform.data"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_policy.namespace", "scope", "NAMESPACE"),
					resource.TestCheckResourceAttr("kestra_policy.namespace", "policy_id", "terraform-namespace-policy"),
					resource.TestCheckResourceAttr("kestra_policy.namespace", "namespace", "io.kestra.terraform.data"),
				),
			},
		},
	})
}

func TestAccPolicyResourceInstance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyInstance("terraform-instance-policy"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_policy.instance", "scope", "INSTANCE"),
					resource.TestCheckResourceAttr("kestra_policy.instance", "policy_id", "terraform-instance-policy"),
					resource.TestCheckNoResourceAttr("kestra_policy.instance", "tenant_id"),
				),
			},
		},
	})
}

func TestUnitPolicyResourceValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		expectError string
	}{
		{
			name: "namespace on tenant scope",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "TENANT"
  policy_id = "invalid"
  namespace = "company.team"
  content   = "id: invalid\nrules:\n  - type: io.kestra.plugin.ee.rules.Deny\n    on: PLUGIN\n"
}`,
			expectError: `namespace must not be set for a TENANT scope policy`,
		},
		{
			name: "tenant_id on instance scope",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "INSTANCE"
  policy_id = "invalid"
  tenant_id = "main"
  content   = "id: invalid\nrules:\n  - type: io.kestra.plugin.ee.rules.Deny\n    on: PLUGIN\n"
}`,
			expectError: `tenant_id must not be set for an INSTANCE scope policy`,
		},
		{
			name: "missing namespace on namespace scope",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "NAMESPACE"
  policy_id = "invalid"
  content   = "id: invalid\nrules:\n  - type: io.kestra.plugin.ee.rules.Deny\n    on: PLUGIN\n"
}`,
			expectError: `namespace is required for a NAMESPACE scope policy`,
		},
		{
			name: "content id mismatch",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "TENANT"
  policy_id = "invalid"
  content   = "id: another-id\nrules:\n  - type: io.kestra.plugin.ee.rules.Deny\n    on: PLUGIN\n"
}`,
			expectError: `must match policy_id`,
		},
		{
			name: "content is not a document",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "TENANT"
  policy_id = "invalid"
  content   = "- type: io.kestra.plugin.ee.rules.Deny"
}`,
			expectError: `content must be a YAML document`,
		},
		{
			name: "content without id",
			config: `
resource "kestra_policy" "invalid" {
  scope     = "TENANT"
  policy_id = "invalid"
  content   = "description: no id"
}`,
			expectError: "content must carry a string `id`",
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

const testUnitProviderConfig = `
provider "kestra" {
  url = "http://localhost:8088"
}
`

func testAccPolicyTenant(policyId, description string) string {
	return fmt.Sprintf(`
resource "kestra_policy" "tenant" {
  scope     = "TENANT"
  policy_id = "%s"

  content = <<EOT
id: %s
description: %s
target:
  namespaces:
    - io.kestra.terraform
rules:
  - type: io.kestra.plugin.ee.rules.Deny
    on: PLUGIN
    action: BLOCK
    errorMessage: The Log plugin is denied by terraform
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.core.log.Log
EOT
}`,
		policyId,
		policyId,
		description,
	)
}

func testAccPolicyNamespace(policyId, namespace string) string {
	return fmt.Sprintf(`
resource "kestra_policy" "namespace" {
  scope     = "NAMESPACE"
  policy_id = "%s"
  namespace = "%s"

  content = <<EOT
id: %s
rules:
  - type: io.kestra.plugin.ee.rules.Require
    on: FLOW
    properties:
      - labels.owner
    errorMessage: Flows must carry an owner label
EOT
}`,
		policyId,
		namespace,
		policyId,
	)
}

func testAccPolicyInstance(policyId string) string {
	return fmt.Sprintf(`
resource "kestra_policy" "instance" {
  scope     = "INSTANCE"
  policy_id = "%s"

  content = <<EOT
id: %s
rules:
  - type: io.kestra.plugin.ee.rules.Deny
    on: PLUGIN
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.scripts.shell.Commands
EOT
}`,
		policyId,
		policyId,
	)
}
