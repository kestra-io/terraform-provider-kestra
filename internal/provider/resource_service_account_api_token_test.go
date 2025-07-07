package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServiceAccountApiTokenAccount(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceAccountApiToken(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_service_account_api_token.new", "name", "test-token",
					),
					resource.TestCheckResourceAttr(
						"kestra_service_account_api_token.new", "description", "Test token",
					),
					resource.TestMatchResourceAttr(
						"kestra_service_account_api_token.new", "full_token", regexp.MustCompile(".*"),
					),
				),
			},
			/**
			// not supported
			{
				ResourceName:      "kestra_service_account_api_token.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
			**/
		},
	})
}

func testAccResourceServiceAccountApiToken() string {
	return fmt.Sprintf(
		`
        resource "kestra_service_account" "new" {
            name = "my-service-account-1"
			description = "Test description"
		}

        resource "kestra_service_account_api_token" "new" {
			service_account_id = resource.kestra_service_account.new.id

            name = "test-token"
			description = "Test token"
			max_age = "PT1H"
			extended = false

			depends_on = [resource.kestra_service_account.new]
        }`,
	)
}
