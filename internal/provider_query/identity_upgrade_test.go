package provider_query_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// State written by earlier releases must plan cleanly with this build: 2.1.0 has no resource
// identity, so Terraform asks for an identity upgrade from version 0, and 2.2.0 already stores
// identities, which Terraform refuses to downgrade.
func TestAccIdentityUpgradeFromRelease(t *testing.T) {
	config := `
provider "kestra" {}

resource "kestra_namespace" "upgrade" {
  namespace_id = "tfidentityupgrade"
  description  = "created before resource identity"
}

resource "kestra_flow" "upgrade" {
  namespace = kestra_namespace.upgrade.namespace_id
  flow_id   = "identity-upgrade"
  content   = <<EOT
id: identity-upgrade
namespace: tfidentityupgrade
tasks:
  - id: hello
    type: io.kestra.plugin.core.log.Log
    message: identity upgrade
EOT
}
`

	for _, release := range []string{"2.1.0", "2.2.0"} {
		t.Run(release, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				PreCheck: func() { queryTestAccPreCheck(t) },
				TerraformVersionChecks: []tfversion.TerraformVersionCheck{
					tfversion.SkipBelow(tfversion.Version1_12_0),
				},
				Steps: []resource.TestStep{
					{
						ExternalProviders: map[string]resource.ExternalProvider{
							"kestra": {
								Source:            "kestra-io/kestra",
								VersionConstraint: release,
							},
						},
						Config: config,
					},
					{
						ProtoV5ProviderFactories: queryMuxProviderFactories,
						Config:                   config,
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectEmptyPlan(),
							},
						},
					},
				},
			})
		})
	}
}
