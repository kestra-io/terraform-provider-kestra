package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceSecurityIntegration(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceSecurityIntegration(
					"test-integration",
					"SCIM",
					"Test description",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_security_integration.new", "name", "test-integration",
					),
					resource.TestCheckResourceAttr(
						"kestra_security_integration.new", "type", "SCIM",
					),
					resource.TestCheckResourceAttr(
						"kestra_security_integration.new", "description", "Test description",
					),
				),
			},
			{
				ResourceName:      "kestra_security_integration.new",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"uri",
					"secret_token",
				},
			},
		},
	})
}

func testAccResourceSecurityIntegration(name, integrationType, description string) string {
	return fmt.Sprintf(
		`
        resource "kestra_security_integration" "new" {
            name        = "%s"
            type        = "%s"
            description = "%s"
        }`,
		name,
		integrationType,
		description,
	)
}
