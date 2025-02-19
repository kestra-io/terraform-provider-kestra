package provider

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var providerFactories = map[string]func() (*schema.Provider, error){
	"kestra": func() (*schema.Provider, error) {
		return New("dev", stringToPointer("lde"))(), nil
	},
}

// Provider with KeepOriginalSource to false
var providerFactoriesKOSFalse = map[string]func() (*schema.Provider, error){
	"kestra": func() (*schema.Provider, error) {
		provider := New("dev", nil)()
		provider.Schema["keep_original_source"].Default = false
		return provider, nil
	},
}

var providerTenantFactories = map[string]func() (*schema.Provider, error){
	"kestra": func() (*schema.Provider, error) {
		return New("dev", stringToPointer("unit_test"))(), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("dev", nil)().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if v := os.Getenv("KESTRA_URL"); v == "" {
		t.Fatal("KESTRA_URL must be set for acceptance tests")
	}

	if v := os.Getenv("KESTRA_USERNAME"); v == "" {
		t.Fatal("KESTRA_USERNAME must be set for acceptance tests")
	}

	if v := os.Getenv("KESTRA_PASSWORD"); v == "" {
		t.Fatal("KESTRA_PASSWORD must be set for acceptance tests")
	}
}

func concat(s ...string) string {
	return strings.Join(s, "\n")
}

//goland:noinspection GoUnhandledErrorResult
func copyResource(src, dst string) {
	tmpDir := os.TempDir() + "/unit-test/"

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		_ = os.Mkdir(os.TempDir()+"/unit-test/", 0777)
	}

	source, _ := os.Open(src)
	defer source.Close()

	destination, _ := os.Create(tmpDir + dst)
	defer destination.Close()

	io.Copy(destination, source)
}
