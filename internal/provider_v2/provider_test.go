package provider_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"os"
	"testing"
)

const (
	// providerConfig is a shared configuration to combine with the actual
	// test configuration so the HashiCups client is properly configured.
	// It is also possible to use the HASHICUPS_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.
	providerConfig = `
provider "kestra" {
  url      = "test123"
  host     = "http://localhost:19090"
}
`
)

var (
	// testAccProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"kestra": providerserver.NewProtocol6WithError(New("acc_test_version")()),
	}
)

func testAccPreCheck(t *testing.T) {
	// You can add code here to run prior to any test case execution, for example assertions
	// about the appropriate environment variables being set are common to see in a pre-check
	// function.
	if os.Getenv("KESTRA_URL") == "" {
		t.Fatal("KESTRA_URL must be set for acceptance tests")
	}

	if os.Getenv("KESTRA_USERNAME") == "" {
		t.Fatal("KESTRA_USERNAME must be set for acceptance tests")
	}

	if os.Getenv("KESTRA_PASSWORD") == "" {
		t.Fatal("KESTRA_PASSWORD must be set for acceptance tests")
	}
}
