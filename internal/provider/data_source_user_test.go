package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceUser(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUser("john"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.kestra_user.new", "user_id", regexp.MustCompile("john"),
					),
				),
			},
		},
	})
}

func testAccDataSourceUser(id string) string {
	return fmt.Sprintf(
		`
        data "kestra_user" "new" {
            user_id = "%s"
        }`,
		id,
	)
}

func TestAccDataSourceUserByEmail(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserByEmail("data-source-email@example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.kestra_user.by_email", "user_id",
						"kestra_user.email_lookup", "id",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_user.by_email", "email", "data-source-email@example.com",
					),
				),
			},
		},
	})
}

func testAccDataSourceUserByEmail(email string) string {
	return fmt.Sprintf(
		`
        resource "kestra_user" "email_lookup" {
            email = "%s"
        }

        data "kestra_user" "by_email" {
            email = kestra_user.email_lookup.email
        }`,
		email,
	)
}
