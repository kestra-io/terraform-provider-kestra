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

// Policies are an EE feature gated by the `FEATURE_POLICIES` license entitlement. On an
// instance without it, `FeatureGateSecurityRule` rejects every policies endpoint with a
// bare 403 before the request reaches the controller, so the acceptance tests below can
// only run against a licensed EE instance.
//
// testAccPreCheckPolicies probes the instance once and skips when the feature is not
// available, so an unlicensed environment reports a skip instead of four opaque
// "status 403" failures. Any other probe outcome is left alone: the test then runs and
// fails on its own terms.
var (
	policiesProbeOnce   sync.Once
	policiesProbeReason string
)

func testAccPreCheckPolicies(t *testing.T) {
	testAccPreCheck(t)

	policiesProbeOnce.Do(func() {
		policiesProbeReason = probePoliciesFeature()
	})
	if policiesProbeReason != "" {
		t.Skipf("skipping policy acceptance tests: %s", policiesProbeReason)
	}
}

func probePoliciesFeature() string {
	url := strings.TrimSuffix(os.Getenv("KESTRA_URL"), "/") + "/api/v1/instance/policies"
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
	case http.StatusForbidden, http.StatusNotFound, http.StatusNotImplemented:
		// 403 is what an unlicensed instance returns; 404/501 cover an OSS build or a
		// version predating the policies API
		return fmt.Sprintf(
			"GET %s returned %d (%s) — the instance is not an EE build with the FEATURE_POLICIES entitlement",
			url, resp.StatusCode, strings.TrimSpace(string(body)),
		)
	default:
		return ""
	}
}
