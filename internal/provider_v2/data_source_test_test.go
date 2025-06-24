package provider_v2

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestAccTestDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDataSourceTestSuite("io.kestra.terraform.data", "test-suite-2-already-in-db"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.kestra_test.an_existing_test", "namespace", "io.kestra.terraform.data"),
					resource.TestCheckResourceAttr("data.kestra_test.an_existing_test", "test_id", "test-suite-2-already-in-db"),
				),
			},
		},
	})
}
func testAccDataSourceTestSuite(namespace string, testid string) string {
	return fmt.Sprintf(
		`
        data "kestra_test" "an_existing_test" {
			namespace = "%s"
			test_id = "%s"
        }`,
		namespace,
		testid,
	)
}
