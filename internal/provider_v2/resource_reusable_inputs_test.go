package provider_v2

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccReusableInputsResource(t *testing.T) {
	// delete persists a new revision rather than dropping the row, so a block recreated on a
	// reused instance does not restart at 1: assert the bump, not the absolute value
	var createdRevision string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheckReusableInputs(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccReusableInputsConfig("terraform-reusable-inputs", "io.kestra.terraform.data", "created by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_reusable_inputs.test", "reusable_inputs_id", "terraform-reusable-inputs"),
					resource.TestCheckResourceAttr("kestra_reusable_inputs.test", "namespace", "io.kestra.terraform.data"),
					resource.TestCheckResourceAttrSet("kestra_reusable_inputs.test", "tenant_id"),
					resource.TestMatchResourceAttr("kestra_reusable_inputs.test", "content", regexp.MustCompile("created by terraform")),
					resource.TestCheckResourceAttrWith("kestra_reusable_inputs.test", "revision", func(value string) error {
						createdRevision = value
						return nil
					}),
				),
			},
			// Update and Read testing
			{
				Config: testAccReusableInputsConfig("terraform-reusable-inputs", "io.kestra.terraform.data", "updated by terraform"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kestra_reusable_inputs.test", "reusable_inputs_id", "terraform-reusable-inputs"),
					resource.TestMatchResourceAttr("kestra_reusable_inputs.test", "content", regexp.MustCompile("updated by terraform")),
					resource.TestCheckResourceAttrWith("kestra_reusable_inputs.test", "revision", func(value string) error {
						created, err := strconv.Atoi(createdRevision)
						if err != nil {
							return fmt.Errorf("unreadable revision %q from the create step: %s", createdRevision, err)
						}
						updated, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("unreadable revision %q: %s", value, err)
						}
						if updated != created+1 {
							return fmt.Errorf("expected the update to bump revision %d to %d, got %d", created, created+1, updated)
						}
						return nil
					}),
				),
			},
			// ImportState testing
			{
				ResourceName: "kestra_reusable_inputs.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["kestra_reusable_inputs.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["tenant_id"], rs.Primary.Attributes["namespace"], rs.Primary.Attributes["reusable_inputs_id"]), nil
				},
				// the API round-trips the source verbatim, so even content matches; the
				// resource has no `id` attribute, so match on reusable_inputs_id
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "reusable_inputs_id",
			},
		},
	})
}

func TestAccReusableInputsInheritanceGuard(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheckReusableInputs(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccReusableInputsConfig("terraform-inherited-reusable-inputs", "io.kestra.terraform.data", "defined in the parent namespace"),
			},
			// the block exists only in the parent, and the GET resolves inheritance: importing it
			// under a child namespace must report a missing object rather than bind the parent block
			{
				ResourceName: "kestra_reusable_inputs.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["kestra_reusable_inputs.test"]
					if !ok {
						return "", fmt.Errorf("resource not found in state")
					}
					return fmt.Sprintf("%s/%s.child/%s", rs.Primary.Attributes["tenant_id"], rs.Primary.Attributes["namespace"], rs.Primary.Attributes["reusable_inputs_id"]), nil
				},
				ExpectError: regexp.MustCompile("non-existent remote object"),
			},
		},
	})
}

func testAccReusableInputsConfig(id, namespace, description string) string {
	return fmt.Sprintf(`
resource "kestra_reusable_inputs" "test" {
  reusable_inputs_id = %[1]q
  namespace          = %[2]q

  content = <<EOT
description: %[3]s
inputs:
  - id: environment
    type: SELECT
    values:
      - dev
      - prod
    defaults: dev
EOT
}`, id, namespace, description)
}

func TestUnitParseReusableInputsImportId(t *testing.T) {
	valid := []struct {
		name     string
		id       string
		expected reusableInputsImportId
	}{
		{
			name:     "well formed",
			id:       "main/company.team/my-block",
			expected: reusableInputsImportId{TenantId: "main", Namespace: "company.team", ReusableInputsId: "my-block"},
		},
	}
	for _, tt := range valid {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseReusableInputsImportId(tt.id)
			if err != nil {
				t.Fatalf("expected %q to parse, got error: %s", tt.id, err)
			}
			if parsed != tt.expected {
				t.Errorf("expected %+v, got %+v", tt.expected, parsed)
			}
		})
	}

	invalid := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"too few parts", "main/my-block"},
		{"too many parts", "main/company.team/my-block/extra"},
		{"empty tenant", "/company.team/my-block"},
		{"empty namespace", "main//my-block"},
		{"empty id", "main/company.team/"},
	}
	for _, tt := range invalid {
		t.Run(tt.name, func(t *testing.T) {
			if parsed, err := parseReusableInputsImportId(tt.id); err == nil {
				t.Errorf("expected %q to be rejected, got %+v", tt.id, parsed)
			}
		})
	}
}
