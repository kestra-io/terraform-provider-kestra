package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceServiceAccount(t *testing.T) {
	serviceAccountId := os.Getenv(testServiceAccountIdEnv)

	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccRequireServiceAccountId(t, serviceAccountId)
		},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceServiceAccount(serviceAccountId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_service_account.new", "id", serviceAccountId,
					),
					resource.TestCheckResourceAttr(
						"data.kestra_service_account.new", "name", "test-sa",
					),
				),
			},
		},
	})
}

func testAccDataSourceServiceAccount(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_service_account" "new" {
            id = "%s"
        }`,
		id,
	)
}
