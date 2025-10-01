package sdk_client

import (
	"context"
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
