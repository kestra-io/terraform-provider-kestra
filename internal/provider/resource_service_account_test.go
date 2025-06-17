package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServiceAccount(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceServiceAccount(
					"sa-1",
					"group { group_id = kestra_group.group1.id }",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_service_account.new", "name", "sa-1",
					),
					resource.TestCheckResourceAttr(
						"kestra_service_account.new", "description", "Test description",
					),
					resource.TestMatchResourceAttr(
						"kestra_service_account.new", "group.0.group_id", regexp.MustCompile(".*"),
					),
				),
			},
			{
				Config: testAccResourceServiceAccount(
					"sa-2",
					"group { group_id = kestra_group.group1.id }",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_service_account.new", "name", "sa-2",
					),
					resource.TestCheckResourceAttr(
						"kestra_service_account.new", "description", "Test description",
					),
					resource.TestMatchResourceAttr(
						"kestra_service_account.new", "group.0.group_id", regexp.MustCompile(".*"),
					),
				),
			},
			{
				ResourceName:      "kestra_service_account.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceServiceAccount(name, group string) string {
	return fmt.Sprintf(
		`
        resource "kestra_role" "new" {
            name = "my user role"
			permissions {
			    type = "FLOW"
			    permissions = ["READ", "UPDATE"]
			}
        }

        resource "kestra_group" "group1" {
            name = "group 1"
        }

        resource "kestra_service_account" "new" {
            name = "%s"
			description = "Test description"
			%s
        }`,
		name,
		group,
	)
}
