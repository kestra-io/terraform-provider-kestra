package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Namespace and tenant concurrency limits and quotas landed in Kestra 2.0 after
// 2.0.0-rc1. An instance that predates them accepts the field and drops it
// silently rather than rejecting the request, so the only way to tell is to write
// one and see whether it comes back.
//
// testAccPreCheckConcurrency probes once and skips when the fields are not
// supported, so an older instance reports a skip instead of an opaque "block
// count changed from 1 to 0" apply failure. Any other probe outcome is left
// alone: the test then runs and fails on its own terms.
var (
	concurrencyProbeOnce   sync.Once
	concurrencyProbeReason string
)

func testAccPreCheckConcurrency(t *testing.T) {
	testAccPreCheck(t)

	concurrencyProbeOnce.Do(func() {
		concurrencyProbeReason = probeConcurrencySupport()
	})
	if concurrencyProbeReason != "" {
		t.Skipf("skipping concurrency acceptance tests: %s", concurrencyProbeReason)
	}
}

func probeConcurrencySupport() string {
	root := strings.TrimSuffix(os.Getenv("KESTRA_URL"), "/")
	tenant := os.Getenv("KESTRA_TENANT_ID")
	if tenant == "" {
		tenant = "main"
	}
	namespaces := fmt.Sprintf("%s/api/v1/%s/namespaces", root, tenant)
	id := "io.kestra.terraform.concurrencyprobe"

	body, _ := json.Marshal(map[string]interface{}{
		"id":      id,
		"deleted": false,
		"concurrency": map[string]interface{}{
			"limit":    1,
			"behavior": "QUEUE",
		},
	})

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := doProbeRequest(client, http.MethodPost, namespaces, body)
	if err != nil {
		return fmt.Sprintf("the instance is unreachable at %s: %s", namespaces, err)
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	// Clean up regardless of the outcome; the probe namespace is not a fixture.
	defer func() {
		if del, err := doProbeRequest(client, http.MethodDelete, namespaces+"/"+id, nil); err == nil {
			del.Body.Close()
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Sprintf("POST %s returned %d (%s)", namespaces, resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return fmt.Sprintf("could not decode the probe response: %s", err)
	}
	if _, ok := decoded["concurrency"]; !ok {
		return "the instance accepted a namespace concurrency limit but did not return it, so it predates the field"
	}
	return ""
}

func doProbeRequest(client *http.Client, method, url string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token := os.Getenv("KESTRA_API_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	} else if jwt := os.Getenv("KESTRA_JWT"); jwt != "" {
		req.AddCookie(&http.Cookie{Name: "JWT", Value: jwt})
	} else if user := os.Getenv("KESTRA_USERNAME"); user != "" {
		req.SetBasicAuth(user, os.Getenv("KESTRA_PASSWORD"))
	}
	return client.Do(req)
}
