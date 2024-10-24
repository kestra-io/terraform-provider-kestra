package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccKv(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceKv(
					"io.kestra.terraform",
					"string",
					"stringValue",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "id", "io.kestra.terraform/string",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "key", "string",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "stringValue",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_kv.new", "type",
					),
				),
			},
			{
				ResourceName:      "kestra_kv.new",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "STRING",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "stringValue",
					),
				),
			},
			{
				Config: testAccResourceKv(
					"io.kestra.terraform",
					"int",
					"1",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "id", "io.kestra.terraform/int",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "key", "int",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "1",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_kv.new", "type",
					),
				),
			},
			{
				ResourceName:      "kestra_kv.new",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "NUMBER",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "1",
					),
				),
			},
			{
				Config: testAccResourceKvWithType(
					"io.kestra.terraform",
					"int",
					"1",
					"type = \"STRING\"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "id", "io.kestra.terraform/int",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "key", "int",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "1",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "STRING",
					),
				),
			},
			{
				ResourceName:      "kestra_kv.new",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "STRING",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "1",
					),
				),
			},
			{
				Config: testAccResourceKv(
					"io.kestra.terraform",
					"object",
					"{\\\"some\\\":\\\"json\\\"}",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "id", "io.kestra.terraform/object",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "key", "object",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "{\"some\":\"json\"}",
					),
					resource.TestCheckNoResourceAttr(
						"kestra_kv.new", "type",
					),
				),
			},
			{
				ResourceName:      "kestra_kv.new",
				ImportState:       true,
				ImportStateVerify: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "JSON",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "{\"some\":\"json\"}",
					),
				),
			},
			{
				Config: testAccResourceKvWithType(
					"io.kestra.terraform",
					"object",
					"{\\\"some\\\":\\\"json\\\"}",
					"type = \"STRING\"",
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "id", "io.kestra.terraform/object",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "key", "object",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "value", "{\"some\":\"json\"}",
					),
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "type", "STRING",
					),
				),
			},
		},
	})

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy: func(state *terraform.State) error {
			urlEnv := strings.TrimRight(os.Getenv("KESTRA_URL"), "/")
			usernameEnv := os.Getenv("KESTRA_USERNAME")
			passwordEnv := os.Getenv("KESTRA_PASSWORD")
			c, _ := NewClient(urlEnv, 10, &usernameEnv, &passwordEnv, nil, nil, nil, nil, nil)
			url := c.Url + fmt.Sprintf("%s/namespaces/io.kestra.terraform/kv/string", apiRoot(nil))
			request, _ := http.NewRequest("GET", url, nil)
			_, _, httpError := c.rawResponseRequest("GET", request)

			if httpError.StatusCode != http.StatusNotFound {
				return fmt.Errorf("resource 'string' should have been destroyed")
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccResourceKv(
					"io.kestra.terraform",
					"string",
					"stringValue",
				),
			},
		},
	})
}

func testAccResourceKv(namespace string, key string, value string) string {
	return testAccResourceKvWithType(namespace, key, value, "")
}

func testAccResourceKvWithType(namespace string, key string, value string, valueType string) string {
	return fmt.Sprintf(
		`
        resource "kestra_kv" "new" {
            namespace = "%s"
			key = "%s"
            value = "%s"
			%s
        }`,
		namespace,
		key,
		value,
		valueType,
	)
}
