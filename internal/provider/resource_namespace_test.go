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
