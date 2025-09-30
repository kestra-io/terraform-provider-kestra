package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestExtraHeadersWork tests that extra headers are sent in HTTP requests
func TestExtraHeadersWork(t *testing.T) {
	// Mock server to capture headers
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Test with map[string]string (direct usage)
	extraHeaders := map[string]string{"X-Test-Header": "test-value"}
	var headersInterface interface{} = extraHeaders

	client, err := NewClient(server.URL, 10, nil, nil, nil, nil, &headersInterface, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, reqErr := client.request("GET", "/test", nil)
	if reqErr != nil {
		t.Fatal(reqErr.Err)
	}

	// Verify header was sent
	if capturedHeaders.Get("X-Test-Header") != "test-value" {
		t.Error("Extra header was not sent")
	}
}

// TestExtraHeadersWithTerraformType tests map[string]interface{} (Terraform's actual type)
func TestExtraHeadersWithTerraformType(t *testing.T) {
	// Mock server to capture headers
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	// Test with map[string]interface{} (what Terraform actually passes)
	extraHeaders := map[string]interface{}{"X-Terraform-Header": "terraform-value"}
	var headersInterface interface{} = extraHeaders

	client, err := NewClient(server.URL, 10, nil, nil, nil, nil, &headersInterface, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, reqErr := client.request("GET", "/test", nil)
	if reqErr != nil {
		t.Fatal(reqErr.Err)
	}

	// Debug: Print captured headers
	t.Logf("Captured headers: %+v", capturedHeaders)
	t.Logf("Client ExtraHeader: %+v", client.ExtraHeader)
	
	// Verify header was sent
	if capturedHeaders.Get("X-Terraform-Header") != "terraform-value" {
		t.Error("Terraform extra header was not sent")
	}
}

// TestExtraHeadersWithAuth tests headers work with authentication
func TestExtraHeadersWithAuth(t *testing.T) {
	// Mock server to capture headers
	var capturedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	username := "user"
	password := "pass"
	extraHeaders := map[string]string{"X-Auth-Header": "auth-value"}
	var headersInterface interface{} = extraHeaders

	client, err := NewClient(server.URL, 10, &username, &password, nil, nil, &headersInterface, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, reqErr := client.request("GET", "/test", nil)
	if reqErr != nil {
		t.Fatal(reqErr.Err)
	}

	// Verify both auth and extra headers are present
	if capturedHeaders.Get("Authorization") == "" {
		t.Error("Authorization header missing")
	}
	if capturedHeaders.Get("X-Auth-Header") != "auth-value" {
		t.Error("Extra header missing with auth")
	}
}
