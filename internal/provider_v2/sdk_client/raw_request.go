package sdk_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kestra-io/client-sdk/go-sdk/kestra_api_client"
)

// RawRequest performs a JSON HTTP call reusing the SDK client's server URL and
// default headers (auth, extra headers). It returns the decoded JSON body or an
// httpStatus on non-2xx responses so callers can react to 404s.
func RawRequest(ctx context.Context, c *kestra_api_client.APIClient, method, relPath string, body interface{}) (map[string]interface{}, int, error) {
	cfg := c.GetConfig()
	if len(cfg.Servers) == 0 {
		return nil, 0, fmt.Errorf("no server configured")
	}
	url := cfg.Servers[0].URL + relPath

	var bodyReader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	for k, v := range cfg.DefaultHeader {
		req.Header.Set(k, v)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("status %d: %s", resp.StatusCode, string(raw))
	}

	if resp.StatusCode == http.StatusNoContent || resp.ContentLength == 0 {
		return nil, resp.StatusCode, nil
	}

	out := map[string]interface{}{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil && err != io.EOF {
		return nil, resp.StatusCode, fmt.Errorf("decode response: %w", err)
	}
	return out, resp.StatusCode, nil
}
