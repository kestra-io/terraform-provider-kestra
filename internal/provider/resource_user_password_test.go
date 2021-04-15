package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserPassword(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceUserPassword(
					"PassPass3",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user_password.new", "password", "PassPass3",
					),
				),
			},
			{
				Config: testAccResourceUserPassword(
					"PassPass2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user_password.new", "password", "PassPass2",
					),
				),
			},
		},
	})
}

func testAccResourceUserPassword(password string) string {
	return fmt.Sprintf(
		`
        resource "kestra_user" "new" {
            username = "my-new-user"
        }

        resource "kestra_user_password" "new" {
            user_id = kestra_user.new.id
            password = "%s"
        }`,
		password,
	)
}
