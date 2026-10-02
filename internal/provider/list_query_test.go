package provider_test

import (
	"context"
	"fmt"
	"strings"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccQueryFlowList(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { queryTestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV5ProviderFactories: queryMuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform.bulkimport",
					"query-test",
					"id: query-test",
					"namespace: io.kestra.terraform.bulkimport",
					"tasks:",
					"  - id: hello",
					"    type: io.kestra.plugin.core.log.Log",
					"    message: query test",
				),
			},
			{
				Query: true,
				Config: `
provider "kestra" {}

list "kestra_flow" "all_flows" {
  provider         = kestra
  limit            = 1000
  include_resource = true
}

list "kestra_flow" "limited_flows" {
  provider = kestra
  limit    = 1
}
`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("kestra_flow.all_flows", map[string]knownvalue.Check{
						"namespace": knownvalue.StringExact("io.kestra.terraform.bulkimport"),
						"flow_id":   knownvalue.StringExact("query-test"),
					}),
					querycheck.ExpectResourceKnownValues(
						"kestra_flow.all_flows",
						queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
							"namespace": knownvalue.StringExact("io.kestra.terraform.bulkimport"),
							"flow_id":   knownvalue.StringExact("query-test"),
						}),
						[]querycheck.KnownValueCheck{
							{
								tfjsonpath.New("namespace"),
								knownvalue.StringExact("io.kestra.terraform.bulkimport"),
							},
							{
								tfjsonpath.New("flow_id"),
								knownvalue.StringExact("query-test"),
							},
						},
					),
					querycheck.ExpectLength("kestra_flow.limited_flows", 1),
				},
			},
		},
	})
}

func TestAccQueryNamespaceList(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() { queryTestAccPreCheck(t) },
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ProtoV5ProviderFactories: queryMuxProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFlow(
					"io.kestra.terraform.bulkimport",
					"namespace-seed",
					"id: namespace-seed",
					"namespace: io.kestra.terraform.bulkimport",
					"tasks:",
					"  - id: hello",
					"    type: io.kestra.plugin.core.log.Log",
					"    message: namespace seed",
				),
			},
			{
				Query: true,
				Config: `
provider "kestra" {}

list "kestra_namespace" "all_namespaces" {
  provider         = kestra
  limit            = 1000
  include_resource = true
}

list "kestra_namespace" "limited_namespaces" {
  provider = kestra
  limit    = 1
}
`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("kestra_namespace.all_namespaces", map[string]knownvalue.Check{
						"namespace_id": knownvalue.StringExact("io.kestra.terraform.bulkimport"),
					}),
					querycheck.ExpectResourceKnownValues(
						"kestra_namespace.all_namespaces",
						queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
							"namespace_id": knownvalue.StringExact("io.kestra.terraform.bulkimport"),
						}),
						[]querycheck.KnownValueCheck{
							{
								tfjsonpath.New("namespace_id"),
								knownvalue.StringExact("io.kestra.terraform.bulkimport"),
							},
						},
					),
					querycheck.ExpectLength("kestra_namespace.limited_namespaces", 1),
				},
			},
		},
	})
}


var queryMuxProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"kestra": func() (tfprotov5.ProviderServer, error) {
		ctx := context.Background()
		providers := []func() tfprotov5.ProviderServer{
			providerserver.NewProtocol5(provider_v2.New("test")()),
			provider.New("test", nil)().GRPCProvider,
		}
		mux, err := tf5muxserver.NewMuxServer(ctx, providers...)
		if err != nil {
			return nil, err
		}
		return mux.ProviderServer(), nil
	},
}

func queryTestAccPreCheck(t *testing.T) {
	if v := os.Getenv("KESTRA_URL"); v == "" {
		t.Fatal("KESTRA_URL must be set for acceptance tests")
	}

	hasBasicAuth := os.Getenv("KESTRA_USERNAME") != "" && os.Getenv("KESTRA_PASSWORD") != ""
	hasTokenAuth := os.Getenv("KESTRA_API_TOKEN") != "" || os.Getenv("KESTRA_JWT") != ""
	if !hasBasicAuth && !hasTokenAuth {
		t.Fatal("KESTRA_USERNAME/KESTRA_PASSWORD or KESTRA_API_TOKEN (or KESTRA_JWT) must be set for acceptance tests")
	}
}

func testAccResourceFlow(id, name string, content ...string) string {
	return fmt.Sprintf(
		`
resource "kestra_flow" "new" {
  namespace = "%s"
  flow_id   = "%s"
  content = <<EOT
%s
EOT
}`,
		id,
		name,
		queryConcat(content...),
	)
}

func queryConcat(s ...string) string {
	return strings.Join(s, "\n")
}
