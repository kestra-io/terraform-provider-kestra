package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNamespace(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNamespace(
					"io.kestra.terraform",
					"My Kestra Namespace",
					concat(
						"k1: 1",
						"k2:",
						"    v1: 1",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "namespace_id", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "description", "My Kestra Namespace",
					),
				),
			},
			{
				Config: testAccResourceNamespace(
					"io.kestra.terraform",
					"My Kestra Namespace 2",
					concat(
						"k2:",
						"    v1: 1",
						"k1: 1",
					),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "namespace_id", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "description", "My Kestra Namespace 2",
					),
					resource.TestMatchResourceAttr(
						"kestra_namespace.new", "variables", regexp.MustCompile(".*k1: 1.*"),
					),
				),
			},
			{
				Config: testAccResourceNamespaceWorkerSelector(
					"io.kestra.terraform",
					"My Kestra Namespace 3",
					concat(
						"k2:",
						"    v1: 1",
						"k1: 1",
					),
					"tf-acc-ns-queue",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "namespace_id", "io.kestra.terraform",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "description", "My Kestra Namespace 3",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "default_worker_selector.0.tags.#", "1",
					),
					resource.TestCheckResourceAttr(
						"kestra_namespace.new", "default_worker_selector.0.fallback", "WAIT",
					),
				),
			},
			{
				ResourceName:            "kestra_namespace.new",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"variables"},
			},
		},
	})
}

func TestAccResourceNamespaceNestedSecretConfiguration(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "kestra_namespace" "nested" {
						namespace_id = "io.kestra.terraform.nested"
						description  = "nested secret config v1"
						secret_configuration = {
							vault = {
								address = "https://vault.example.invalid"
								auth = {
									method = "approle"
									role   = "kestra"
								}
							}
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_namespace.nested", "namespace_id", "io.kestra.terraform.nested"),
				),
			},
			{
				Config: `
					resource "kestra_namespace" "nested" {
						namespace_id = "io.kestra.terraform.nested"
						description  = "nested secret config v2"
						secret_configuration = {
							vault = {
								address = "https://vault.example.invalid"
								auth = {
									method = "token"
								}
							}
						}
					}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_namespace.nested", "description", "nested secret config v2"),
				),
			},
		},
	})
}

func testAccResourceNamespace(id, description, variables string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace" "new" {
            namespace_id = "%s"
            description = "%s"
            variables = <<EOT
%s
EOT
        }`,
		id,
		description,
		variables,
	)
}

// The 2.0 namespace routes tasks through a tag set matched against Worker Queues, so the
// selector is exercised against a real kestra_worker_queue rather than a worker group key.
func testAccResourceNamespaceWorkerSelector(id, description, variables string, queueId string) string {
	return fmt.Sprintf(
		`
		resource "kestra_worker_queue" "new" {
			queue_id = "%s"
			tags     = ["%s"]
		}

        resource "kestra_namespace" "new" {
            namespace_id = "%s"
            description = "%s"
            variables = <<EOT
%s
EOT
			default_worker_selector {
				tags     = kestra_worker_queue.new.tags
				fallback = "WAIT"
			}
        }`,
		queueId,
		queueId,
		id,
		description,
		variables,
	)
}

// TestAccNamespaceConcurrencyAndQuotas covers the Kestra 2.0 concurrency limit
// and quotas. The write path replaces the whole namespace, so the last step drops
// both blocks to pin that removing them actually clears them server-side rather
// than leaving the previous values in place.
func TestAccNamespaceConcurrencyAndQuotas(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckConcurrency(t) },
		ProtoV5ProviderFactories: muxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceNamespaceConcurrency("io.kestra.terraform.concurrency", 5, "QUEUE", "PT1H", 10, "FAIL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "concurrency.0.limit", "5"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "concurrency.0.behavior", "QUEUE"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.0.duration", "PT1H"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.0.limit", "10"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.0.behavior", "FAIL"),
				),
			},
			{
				Config: testAccResourceNamespaceConcurrency("io.kestra.terraform.concurrency", 3, "FAIL", "PT30M", 7, "CANCEL"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "concurrency.0.limit", "3"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "concurrency.0.behavior", "FAIL"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.0.limit", "7"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.0.behavior", "CANCEL"),
				),
			},
			{
				Config: testAccResourceNamespaceConcurrencyBare("io.kestra.terraform.concurrency"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "concurrency.#", "0"),
					resource.TestCheckResourceAttr("kestra_namespace.concurrency", "quotas.#", "0"),
				),
			},
		},
	})
}

func testAccResourceNamespaceConcurrency(id string, limit int, behavior, quotaDuration string, quotaLimit int, quotaBehavior string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace" "concurrency" {
            namespace_id = "%s"

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

func testAccResourceNamespaceConcurrencyBare(id string) string {
	return fmt.Sprintf(
		`
        resource "kestra_namespace" "concurrency" {
            namespace_id = "%s"
        }`,
		id,
	)
}
