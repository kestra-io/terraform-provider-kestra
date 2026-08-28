//go:build e2e

package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// env reads a required KESTRA_E2E_* value. These come from `terraform output` in
// e2e-test.sh; a missing one means the driver and the Terraform outputs have drifted,
// which is a failure and not something to skip over.
func env(t *testing.T, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s is not set — run this through ./e2e-test.sh, which feeds it from terraform output", name)
	}
	return v
}

type client struct {
	baseURL string
	tenant  string
	token   string
	who     string
	http    *http.Client
}

func newClient(t *testing.T, who, tokenEnv string) *client {
	t.Helper()
	return &client{
		baseURL: strings.TrimRight(env(t, "KESTRA_E2E_URL"), "/"),
		tenant:  env(t, "KESTRA_E2E_TENANT"),
		token:   env(t, tokenEnv),
		who:     who,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// do issues a tenant-scoped request. relPath is appended to /api/v1/{tenant}.
func (c *client) do(t *testing.T, method, relPath string, query url.Values) (int, []byte) {
	t.Helper()

	u := fmt.Sprintf("%s/api/v1/%s%s", c.baseURL, c.tenant, relPath)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		t.Fatalf("as %s: build %s %s: %v", c.who, method, relPath, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("as %s: %s %s: %v", c.who, method, relPath, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("as %s: read body of %s %s: %v", c.who, method, relPath, err)
	}
	return resp.StatusCode, body
}

// expectOK fails with the response body when the call did not succeed. The body is
// where Kestra puts the reason, so it belongs in the failure message.
func (c *client) expectOK(t *testing.T, method, relPath string, query url.Values) []byte {
	t.Helper()
	code, body := c.do(t, method, relPath, query)
	if code < 200 || code >= 300 {
		t.Fatalf("as %s: %s %s returned %d, want 2xx\nbody: %s", c.who, method, relPath, code, truncate(body))
	}
	return body
}

// expectDenied accepts 403 and 404: Kestra hides some resources rather than admitting
// they exist, and which one it picks is not part of the contract under test. Anything
// else — a 2xx above all — is a real failure.
func (c *client) expectDenied(t *testing.T, method, relPath string, query url.Values) {
	t.Helper()
	code, body := c.do(t, method, relPath, query)
	switch code {
	case http.StatusForbidden, http.StatusNotFound:
		t.Logf("as %s: %s %s correctly denied with %d", c.who, method, relPath, code)
	default:
		t.Fatalf("as %s: %s %s returned %d, want 403 or 404 — the role or binding is wider than intended\nbody: %s",
			c.who, method, relPath, code, truncate(body))
	}
}

func decode[T any](t *testing.T, body []byte, what string) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode %s: %v\nbody: %s", what, err, truncate(body))
	}
	return out
}

func truncate(b []byte) string {
	const max = 4000
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "…(truncated)"
}
