package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"os"
	"strconv"
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
					resource.TestCheckResourceAttr(
						"kestra_kv.new", "description", "",
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
			defaultMainTenantId := "main"

			c, _ := NewClient(urlEnv, 10, &usernameEnv, &passwordEnv, nil, nil, nil, &defaultMainTenantId, nil)
			url := c.Url + fmt.Sprintf("%s/namespaces/io.kestra.terraform/kv/string", apiRoot(&defaultMainTenantId))
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

func TestAccKvDescription(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckKvDescription(t)
		},
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceKvWithDescription(
					"io.kestra.terraform",
					"description",
					"value",
					"A description with café 東京 🚀",
				),
				Check: resource.TestCheckResourceAttr(
					"kestra_kv.new", "description", "A description with café 東京 🚀",
				),
			},
			{
				Config: testAccResourceKvWithDescription(
					"io.kestra.terraform",
					"description",
					"value",
					"Updated description: naïve façade",
				),
				Check: resource.TestCheckResourceAttr(
					"kestra_kv.new", "description", "Updated description: naïve façade",
				),
			},
			{
				Config: testAccResourceKv("io.kestra.terraform", "description", "value"),
				Check:  resource.TestCheckResourceAttr("kestra_kv.new", "description", ""),
			},
		},
	})
}

func testAccPreCheckKvDescription(t *testing.T) {
	version := strings.TrimPrefix(os.Getenv("KESTRA_VERSION"), "v")
	if version == "" || version == "develop" {
		return
	}

	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		t.Fatalf("KESTRA_VERSION %q is not a valid Kestra version", version)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		t.Fatalf("KESTRA_VERSION %q is not a valid Kestra version: %v", version, err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		t.Fatalf("KESTRA_VERSION %q is not a valid Kestra version: %v", version, err)
	}
	if major < 2 || (major == 2 && minor < 1) {
		t.Skipf("KV description headers require Kestra 2.1.0 or later; skipping on %s", version)
	}
}

func testAccResourceKv(namespace string, key string, value string) string {
	return testAccResourceKvWithType(namespace, key, value, "")
}

func testAccResourceKvWithDescription(namespace, key, value, description string) string {
	return fmt.Sprintf(`
        resource "kestra_kv" "new" {
            namespace = %q
			key = %q
            value = %q
			description = %q
        }`, namespace, key, value, description)
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
