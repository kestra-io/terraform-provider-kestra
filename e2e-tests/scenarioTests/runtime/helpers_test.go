//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

// env reads a required KESTRA_E2E_* value. These come from `terraform output` in
// e2e-test.sh; a missing one means the driver and the Terraform outputs have drifted,
// which is a failure rather than something to skip over.
func env(t *testing.T, name string) string {
	t.Helper()
	v := os.Getenv(name)
	if v == "" {
		t.Fatalf("%s is not set — run this through ./e2e-test.sh, which feeds it from terraform output", name)
	}
	return v
}

type client struct {
	baseURL  string
	tenant   string
	who      string
	token    string // bearer, when acting as a Terraform-created user
	username string // basic auth, when acting as the configured super-admin
	password string
	http     *http.Client
}

func base(t *testing.T, who string) *client {
	t.Helper()
	return &client{
		baseURL: strings.TrimRight(env(t, "KESTRA_E2E_URL"), "/"),
		tenant:  env(t, "KESTRA_E2E_TENANT"),
		who:     who,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

// asUser authenticates with an API token issued to a user Terraform created.
func asUser(t *testing.T, who, tokenEnv string) *client {
	t.Helper()
	c := base(t, who)
	c.token = env(t, tokenEnv)
	return c
}

// asSuperAdmin authenticates with the credentials from kestra.security.super-admin.
// Used only where the assertion is about a resource round-tripping rather than about
// RBAC, so an unrelated permission gap cannot turn into a confusing failure.
func asSuperAdmin(t *testing.T) *client {
	t.Helper()
	c := base(t, "super-admin")
	c.username = env(t, "KESTRA_E2E_USERNAME")
	c.password = env(t, "KESTRA_E2E_PASSWORD")
	return c
}

type request struct {
	method string
	path   string // appended to /api/v1/{tenant}, or to /api/v1 when instance is set
	// instance targets a non-tenant-scoped route, e.g. /tenants or /users. Kestra moved
	// these out from under the tenant segment in 0.23.
	instance bool
	query    url.Values
	form     map[string]string // sent as multipart/form-data
	raw      string
	rawType  string
}

func (c *client) send(t *testing.T, r request) (int, []byte) {
	t.Helper()

	u := fmt.Sprintf("%s/api/v1/%s%s", c.baseURL, c.tenant, r.path)
	if r.instance {
		u = fmt.Sprintf("%s/api/v1%s", c.baseURL, r.path)
	}
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	var body io.Reader
	contentType := ""

	switch {
	case r.form != nil:
		// Flow inputs travel as multipart/form-data. Only wait, labels, revision,
		// scheduleDate, breakpoints and kind are query parameters — passing an input as
		// a query parameter is silently ignored and the flow falls back to its default.
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		for k, v := range r.form {
			if err := w.WriteField(k, v); err != nil {
				t.Fatalf("as %s: write form field %q: %v", c.who, k, err)
			}
		}
		if err := w.Close(); err != nil {
			t.Fatalf("as %s: close multipart writer: %v", c.who, err)
		}
		body = &buf
		contentType = w.FormDataContentType()
	case r.raw != "":
		body = strings.NewReader(r.raw)
		contentType = r.rawType
	}

	req, err := http.NewRequest(r.method, u, body)
	if err != nil {
		t.Fatalf("as %s: build %s %s: %v", c.who, r.method, r.path, err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else {
		req.SetBasicAuth(c.username, c.password)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("as %s: %s %s: %v", c.who, r.method, r.path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("as %s: read body of %s %s: %v", c.who, r.method, r.path, err)
	}
	return resp.StatusCode, raw
}

// expectOK fails with the response body when the call did not succeed. Kestra puts the
// reason in the body, so it belongs in the failure message.
func (c *client) expectOK(t *testing.T, r request) []byte {
	t.Helper()
	code, body := c.send(t, r)
	if code < 200 || code >= 300 {
		t.Fatalf("as %s: %s %s returned %d, want 2xx\nbody: %s", c.who, r.method, r.path, code, truncate(body))
	}
	return body
}

// expectDenied accepts 401, 403 and 404: Kestra hides some resources rather than
// admitting they exist, and which it picks is not part of the contract under test.
//
// Every caller sends a well-formed request, so a 4xx outside that set means something
// other than authorization rejected it and the assertion proved nothing — that is
// reported as a failure rather than quietly counted as a denial.
func (c *client) expectDenied(t *testing.T, r request) {
	t.Helper()
	code, body := c.send(t, r)
	switch code {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		t.Logf("as %s: %s %s correctly denied with %d", c.who, r.method, r.path, code)
	default:
		if code >= 200 && code < 300 {
			t.Fatalf("as %s: %s %s SUCCEEDED with %d — the role or binding is wider than intended\nbody: %s",
				c.who, r.method, r.path, code, truncate(body))
		}
		t.Fatalf("as %s: %s %s returned %d, which is neither success nor a denial — the request was rejected for some other reason, so this assertion proved nothing\nbody: %s",
			c.who, r.method, r.path, code, truncate(body))
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
