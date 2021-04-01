package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNamespaceSecret(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNamespaceSecret(
					"my-secret",
					"my-value",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace_secret.new", "secret_key", "my-secret",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace_secret.new", "secret_value", "my-value",
					),
				),
			},
			{
				Config: testAccResourceNamespaceSecret(
					"my-secret",
					"my-value2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace_secret.new", "secret_key", "my-secret",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace_secret.new", "secret_value", "my-value2",
					),
				),
			},
		},
	})
}

func testAccResourceNamespaceSecret(key, value string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace" "new" {
            namespace_id = "io.kestra.terraform.secret"
        }

        resource "kestra_namespace_secret" "new" {
            namespace = kestra_namespace.new.id
            secret_key = "%s"
            secret_value = "%s"
        }`,
		key,
		value,
	)
}
