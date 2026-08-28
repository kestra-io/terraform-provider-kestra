package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserGroupMembership(t *testing.T) {
	// Kestra 2.0 (kestra-ee #10493) stopped auto-granting tenant access when a user
	// is added to a group: addUserToGroup now resolves the user through a
	// hasTenantAccess filter and 404s with "User does not exist for id" otherwise.
	// The provider has no way to grant tenant access -- kestra_user has no `tenants`
	// attribute and there is no tenant-access resource -- so this cannot be fixed in
	// the config. Unskip once the provider can grant tenant access.
	t.Skip("kestra_user cannot grant tenant access, which Kestra 2.0 now requires before a group add")

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	config := fmt.Sprintf(`
		resource "kestra_group" "platform" {
			name = "platform-team-tf-%[1]s"
		}

		resource "kestra_group" "data" {
			name = "data-team-tf-%[1]s"
		}

		resource "kestra_user" "alice" {
			email = "alice-membership-%[1]s@test.local"
			lifecycle {
				ignore_changes = [groups]
			}
		}

		resource "kestra_user_group_membership" "alice_platform" {
			user_id  = kestra_user.alice.id
			group_id = kestra_group.platform.id
		}

		resource "kestra_user_group_membership" "alice_data" {
			user_id  = kestra_user.alice.id
			group_id = kestra_group.data.id
		}
	`, suffix)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"kestra_user_group_membership.alice_platform", "user_id",
						"kestra_user.alice", "id",
					),
					resource.TestCheckResourceAttrPair(
						"kestra_user_group_membership.alice_platform", "group_id",
						"kestra_group.platform", "id",
					),
					resource.TestCheckResourceAttrPair(
						"kestra_user_group_membership.alice_data", "group_id",
						"kestra_group.data", "id",
					),
				),
			},
			{
				ResourceName:      "kestra_user_group_membership.alice_platform",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
