package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccUserTenantAccess(t *testing.T) {
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	config := fmt.Sprintf(`
		resource "kestra_user" "bob" {
			email = "bob-tenant-access-%[1]s@test.local"
			lifecycle {
				ignore_changes = [groups]
			}
		}

		resource "kestra_user_tenant_access" "bob" {
			user_id = kestra_user.bob.id
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
						"kestra_user_tenant_access.bob", "user_id",
						"kestra_user.bob", "id",
					),
					resource.TestCheckResourceAttrPair(
						"kestra_user_tenant_access.bob", "id",
						"kestra_user.bob", "id",
					),
				),
			},
			{
				ResourceName:      "kestra_user_tenant_access.bob",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestUnitParseUserTenantAccessImportId(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "user id", id: "6oYBldfatcg6nkR2TBKRv6", wantErr: false},
		{name: "empty", id: "", wantErr: true},
		{name: "tenant qualified", id: "main/6oYBldfatcg6nkR2TBKRv6", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := resourceUserTenantAccess().TestResourceData()
			d.SetId(tt.id)

			_, err := resourceUserTenantAccessImport(nil, d, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for id %q, got none", tt.id)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for id %q: %v", tt.id, err)
			}
			if got := d.Get("user_id").(string); got != tt.id {
				t.Fatalf("user_id = %q, want %q", got, tt.id)
			}
		})
	}
}
