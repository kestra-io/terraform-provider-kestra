package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceServiceAccountApiTokens(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceServiceAccountApiTokens("2EPi5XC0oluKRCVF56gcC"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "service_account_id", "2EPi5XC0oluKRCVF56gcC",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.name", "test",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.description", "test",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.token_prefix", "TCAMX5",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.iat", "2024-01-01T00:00:00Z",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.exp", "2024-01-02T00:00:00Z",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account_api_tokens.new", "api_tokens.0.extended", "false",
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
