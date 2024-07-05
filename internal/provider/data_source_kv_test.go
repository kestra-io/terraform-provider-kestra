package provider

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceKv(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "string"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "string",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "stringValue",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "int"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "int",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "1",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "double"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "double",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "1.5",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "falseBoolean"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "falseBoolean",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "false",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "trueBoolean"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "trueBoolean",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "true",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "dateTime"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "dateTime",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "2022-05-01T03:02:01Z",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "date"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "date",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "2022-05-01",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "duration"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "duration",
					),
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "value", "PT75H2M1S",
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "object"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "object",
					),
					resource.TestCheckResourceAttrWith(
						"data.kestra_kv.new", "value", func(value string) error {
							var obj map[string]string
							err := json.Unmarshal([]byte(value), &obj)
							if err != nil {
								return err
							}
							if obj["some"] == "value" && obj["in"] == "object" {
								return nil
							}
							return fmt.Errorf("unexpected value: %s", value)
						},
					),
				),
			},
			{
				Config: testAccDataSourceKv("io.kestra.terraform.data", "array"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.kestra_kv.new", "key", "array",
					),
					resource.TestCheckResourceAttrWith(
						"data.kestra_kv.new", "value", func(value string) error {
							var obj []map[string]string
							err := json.Unmarshal([]byte(value), &obj)
							if err != nil {
								return err
							}
							if obj[0]["some"] == "value" && obj[0]["in"] == "object" &&
								obj[1]["yet"] == "another" && obj[1]["array"] == "object" {
								return nil
							}
							return fmt.Errorf("unexpected value: %s", value)
						},
					),
				),
			},
		},
	})
}

func testAccDataSourceKv(namespace, key any) string {
	return fmt.Sprintf(
		`
        data "kestra_kv" "new" {
            namespace = "%s"
            key = "%s"
        }`,
		namespace,
		key,
	)
}
