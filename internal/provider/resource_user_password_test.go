package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserPasswordFlow(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserPasswordFlow(
					"pass1",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user_password.new", "password", "pass1",
					),
				),
			},
			{
				Config: testAccUserPasswordFlow(
					"pass2",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_user_password.new", "password", "pass2",
					),
				),
			},
		},
	})
}

func testAccUserPasswordFlow(password string) string {
	return fmt.Sprintf(
		`
        resource "kestra_user" "new" {
            username = "my-user"
        }


        resource "kestra_user_password" "new" {
            user_id = kestra_user.new.id
            password = "%s"
        }`,
		password,
	)
}
