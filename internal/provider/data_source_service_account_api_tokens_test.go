package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceServiceAccountApiTokens(t *testing.T) {
	serviceAccountId := os.Getenv(testServiceAccountIdEnv)

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccRequireServiceAccountId(t, serviceAccountId)
		},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceServiceAccountApiTokens(serviceAccountId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "service_account_id", serviceAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.#", "1",
					),
					// Only the fields the create request controls can be asserted
					// exactly; the prefix and the timestamps are server generated.
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.name", "test",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.description", "test",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.extended", "false",
					),
					resource.TestMatchResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.token_prefix", regexp.MustCompile(`^.+$`),
					),
					resource.TestMatchResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.iat",
						regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`),
					),
					resource.TestMatchResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.exp",
						regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T`),
					),
				),
			},
		},
	})
}

func testAccDataSourceServiceAccountApiTokens(id string) string {
	return fmt.Sprintf(
		`
			data "kestra_service_account_api_tokens" "new" {
				service_account_id = "%s"

			}
			`,
		id,
	)
}
