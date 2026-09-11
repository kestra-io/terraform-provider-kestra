package provider_v2

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Reusable inputs need an EE build carrying the reusable-inputs routes; on an OSS or older
// build they are not registered and answer 404 before reaching a controller.
//
// Only "route absent" statuses (404, 501) skip: reusable inputs are gated by the
// REUSABLE_INPUTS ACL resource rather than a license entitlement, so a 403 is a permissions
// failure worth seeing. Every other status falls through and fails on its own terms.
var (
	reusableInputsProbeOnce   sync.Once
	reusableInputsProbeReason string
)

func testAccPreCheckReusableInputs(t *testing.T) {
	testAccPreCheck(t)

	reusableInputsProbeOnce.Do(func() {
		reusableInputsProbeReason = probeReusableInputsFeature()
	})
	if reusableInputsProbeReason != "" {
		t.Skipf("skipping reusable inputs acceptance tests: %s", reusableInputsProbeReason)
	}
}

func probeReusableInputsFeature() string {
	tenant := os.Getenv("KESTRA_TENANT_ID")
	if tenant == "" {
		tenant = "main"
	}

	// the namespaces listing is the only reusable-inputs route taking neither a namespace
	// nor an id, so a build carrying the feature answers it without any data set up
	url := fmt.Sprintf("%s/api/v1/%s/reusable-inputs/namespaces", strings.TrimSuffix(os.Getenv("KESTRA_URL"), "/"), tenant)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Sprintf("unable to build the probe request: %s", err)
	}
	if token := os.Getenv("KESTRA_API_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if user := os.Getenv("KESTRA_USERNAME"); user != "" {
		req.SetBasicAuth(user, os.Getenv("KESTRA_PASSWORD"))
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("the instance is unreachable at %s: %s", url, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))

	switch resp.StatusCode {
	case http.StatusNotFound, http.StatusNotImplemented:
		return fmt.Sprintf(
			"GET %s returned %d (%s) — the instance is not an EE build carrying reusable inputs",
			url, resp.StatusCode, strings.TrimSpace(string(body)),
		)
	default:
		return ""
	}
}
