package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccTenant(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTenant(
					"custom",
					"My custom tenant",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_tenant.new", "tenant_id", "custom",
					),
					resource.TestCheckResourceAttr(
						"kestra_tenant.new", "name", "My custom tenant",
					),
				),
			},
			{
				ResourceName:      "kestra_tenant.new",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccResourceTenant(id, name string) string {
	return fmt.Sprintf(
		`
        resource "kestra_tenant" "new" {
            tenant_id = "%s"
            name = "%s"
        }`,
		id,
		name,
	)
}

// TestAccTenantConcurrencyAndQuotas covers the Kestra 2.0 concurrency limit and
// quotas. The write path replaces the whole tenant, so the last step drops both
// blocks to pin that removing them actually clears them server-side rather than
// leaving the previous values in place.
func TestAccTenantConcurrencyAndQuotas(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTenantConcurrency("concurrency-tenant", "QUEUE", 5, "PT1H", 10, "FAIL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "concurrency.0.limit", "5"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "concurrency.0.behavior", "QUEUE"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.0.duration", "PT1H"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.0.limit", "10"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.0.behavior", "FAIL"),
				),
			},
			{
				Config: testAccResourceTenantConcurrency("concurrency-tenant", "CANCEL", 2, "PT15M", 3, "CANCEL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "concurrency.0.limit", "2"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "concurrency.0.behavior", "CANCEL"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.0.duration", "PT15M"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.0.limit", "3"),
				),
			},
			{
				Config: testAccResourceTenant("concurrency-tenant", "My custom tenant"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_tenant.new", "concurrency.#", "0"),
					resource.TestCheckResourceAttr("kestra_tenant.new", "quotas.#", "0"),
				),
			},
		},
	})
}

func testAccResourceTenantConcurrency(id, behavior string, limit int, quotaDuration string, quotaLimit int, quotaBehavior string) string {
	return fmt.Sprintf(
		`
        resource "kestra_tenant" "concurrency" {
            tenant_id = "%s"
            name = "Concurrency tenant"

            concurrency {
                limit = %d
                behavior = "%s"
            }

            quotas {
                duration = "%s"
                limit = %d
                behavior = "%s"
            }
        }`,
		id, limit, behavior, quotaDuration, quotaLimit, quotaBehavior,
	)
}
