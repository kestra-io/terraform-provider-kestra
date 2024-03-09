package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserApiTokenAccount(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUserApiToken(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user_api_token.new", "name", "test-token",
					),
					resource.TestCheckResourceAttr(
						"kestra_user_api_token.new", "description", "Test token",
					),
					resource.TestMatchResourceAttr(
						"kestra_user_api_token.new", "full_token", regexp.MustCompile(".*"),
					),
				),
			},
			/**
			// not supported
			{
				ResourceName:      "kestra_user_api_token.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
			**/
		},
	})
}

func testAccResourceUserApiToken() string {
	return fmt.Sprintf(
		`
        resource "kestra_service_account" "new" {
            username = "test-service-account"
			description = "Test description"
		}

        resource "kestra_user_api_token" "new" {
			user_id = resource.kestra_service_account.new.id

            name = "test-token"
			description = "Test token"
			max_age = "PT1H"
			extended = false

			depends_on = [resource.kestra_service_account.new]
        }`,
	)
}
