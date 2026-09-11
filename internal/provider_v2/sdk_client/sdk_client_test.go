package sdk_client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClientExtraHeaders tests that extra headers are properly configured in the SDK client
func TestNewClientExtraHeaders(t *testing.T) {
	tests := []struct {
		name            string
		extraHeaders    *map[string]string
		expectedHeaders map[string]string
	}{
		{
			name: "valid extra headers",
			extraHeaders: &map[string]string{
				"X-Custom-Header":  "custom-value",
				"X-Another-Header": "another-value",
			},
			expectedHeaders: map[string]string{
				"X-Custom-Header":  "custom-value",
				"X-Another-Header": "another-value",
			},
		},
		{
			name:            "nil extra headers",
			extraHeaders:    nil,
			expectedHeaders: map[string]string{},
		},
		{
			name:            "empty extra headers",
			extraHeaders:    &map[string]string{},
			expectedHeaders: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with extra headers
			client, err := NewClient(
				context.Background(),
				"http://localhost:8080",
				10,  // timeout
				nil, // username
				nil, // password
				nil, // jwt
				nil, // apiToken
				tt.extraHeaders,
			)
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			// Check that the client was created successfully
			if client == nil {
				t.Fatal("Client should not be nil")
			}

			// Note: The SDK client doesn't expose the configuration directly,
			// so we can't easily test the headers without making actual requests.
			// The headers are set in the configuration.DefaultHeader which is used
			// by the generated SDK client for all requests.
		})
	}
}

// TestNewClientExtraHeadersWithAuth tests extra headers work with authentication
func TestNewClientExtraHeadersWithAuth(t *testing.T) {
	username := "testuser"
	password := "testpass"
	extraHeaders := map[string]string{
		"X-Custom-Header":  "custom-value",
		"X-Another-Header": "another-value",
	}

	// Create client with both auth and extra headers
	client, err := NewClient(
		context.Background(),
		"http://localhost:8080",
		10, // timeout
		&username,
		&password,
		nil, // jwt
		nil, // apiToken
		&extraHeaders,
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Check that the client was created successfully
	if client == nil {
		t.Fatal("Client should not be nil")
	}
}

func TestUnitDefaultHeaders(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		name     string
		username *string
		password *string
		jwt      *string
		apiToken *string
		extra    *map[string]string
		expected map[string]string
	}{
		{
			name:     "basic auth",
			username: str("user"),
			password: str("pass"),
			expected: map[string]string{"Authorization": "Basic dXNlcjpwYXNz"},
		},
		{
			name:     "jwt goes to the cookie",
			jwt:      str("a-jwt"),
			expected: map[string]string{"Cookie": "JWT=a-jwt"},
		},
		{
			name:     "api token wins over basic auth",
			username: str("user"),
			password: str("pass"),
			apiToken: str("a-token"),
			expected: map[string]string{"Authorization": "Bearer a-token"},
		},
		{
			name:     "extra headers win over auth",
			username: str("user"),
			password: str("pass"),
			extra:    &map[string]string{"Authorization": "Basic overridden"},
			expected: map[string]string{"Authorization": "Basic overridden"},
		},
		{
			name:     "empty api token does not clear basic auth",
			username: str("user"),
			password: str("pass"),
			apiToken: str(""),
			expected: map[string]string{"Authorization": "Basic dXNlcjpwYXNz"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultHeaders(tt.username, tt.password, tt.jwt, tt.apiToken, tt.extra)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
			for k, v := range tt.expected {
				if got[k] != v {
					t.Errorf("header %q: expected %q, got %q", k, v, got[k])
				}
			}
		})
	}
}

// TestUnitNewKestraClientSendsHeaders pins the precedence on the wire: the hand-written
// client applies auth as plain headers, so an Authorization supplied through extra
// headers still wins, as it does for the generated client.
func TestUnitNewKestraClientSendsHeaders(t *testing.T) {
	var got http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	username, password := "user", "pass"
	extra := map[string]string{"Authorization": "Basic overridden", "X-Custom": "custom-value"}

	client := NewKestraClient(server.URL, 10, &username, &password, nil, nil, &extra)
	if _, err := client.ReusableInputs().ListReusableInputsNamespaces(context.Background(), "main"); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if v := got.Get("Authorization"); v != "Basic overridden" {
		t.Errorf("expected the extra header to win, got %q", v)
	}
	if v := got.Get("X-Custom"); v != "custom-value" {
		t.Errorf("expected the custom header to be sent, got %q", v)
	}
}
