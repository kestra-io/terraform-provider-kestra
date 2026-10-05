package provider

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		PreCheck:                 func() { testAccPreCheckConcurrency(t) },
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
				Config: testAccResourceTenantConcurrencyBare("concurrency-tenant"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "concurrency.#", "0"),
					resource.TestCheckResourceAttr("kestra_tenant.concurrency", "quotas.#", "0"),
				),
			},
		},
	})
}

// The bare config has to keep the same resource address as the steps before it.
// Switching to a different address with the same tenant_id makes Terraform create
// the new tenant before destroying the old one, which the API rejects with
// "Tenant id already exists" -- so the step would never reach its assertions.
func testAccResourceTenantConcurrencyBare(id string) string {
	return fmt.Sprintf(
		`
        resource "kestra_tenant" "concurrency" {
            tenant_id = "%s"
            name = "Concurrency tenant"
        }`,
		id,
	)
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

// TestAccTenantDestroyWithResources pins that destroying a tenant which already
// holds a flow, a KV and an execution succeeds and really removes the tenant.
// Tenant deletion used to answer 500 "tenantId cannot be null" on some backends
// (#215); the provider only forwards the DELETE, so this guards the end to end behavior.
func TestAccTenantDestroyWithResources(t *testing.T) {
	// unique per run: recreating a deleted tenant id gives 403 on the first calls on some backends
	tenantId := fmt.Sprintf("destroy-with-resources-%d", time.Now().UnixNano()%1000000)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		CheckDestroy: func(*terraform.State) error {
			status, err := tenantApi("GET", "/tenants/"+tenantId, "", "")
			if err != nil {
				return err
			}
			if status != http.StatusNotFound {
				return fmt.Errorf("tenant %s still exists after destroy, GET returned %d", tenantId, status)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceTenant(tenantId, "Destroy with resources"),
				Check: func(*terraform.State) error {
					calls := []struct{ method, path, contentType, body string }{
						{"POST", "/" + tenantId + "/flows", "application/x-yaml", "id: f1\nnamespace: tenant.destroy\ntasks:\n  - id: t\n    type: io.kestra.plugin.core.log.Log\n    message: hi\n"},
						{"PUT", "/" + tenantId + "/namespaces/tenant.destroy/kv/k1", "text/plain", `"v"`},
						{"POST", "/" + tenantId + "/executions/tenant.destroy/f1", "", ""},
					}
					for _, c := range calls {
						status, err := tenantApi(c.method, c.path, c.contentType, c.body)
						if err != nil {
							return err
						}
						if status != http.StatusOK {
							return fmt.Errorf("%s %s returned %d", c.method, c.path, status)
						}
					}
					return nil
				},
			},
		},
	})
}

func tenantApi(method, path, contentType, body string) (int, error) {
	req, err := http.NewRequest(method, strings.TrimSuffix(os.Getenv("KESTRA_URL"), "/")+"/api/v1"+path, strings.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.SetBasicAuth(os.Getenv("KESTRA_USERNAME"), os.Getenv("KESTRA_PASSWORD"))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	return res.StatusCode, nil
}
