package sdk_client

import (
	"context"
	"encoding/base64"
	"github.com/kestra-io/client-sdk/go-sdk/v2/kestra_api_client"
	"net/http"
	"time"
)

func NewClient(ctx context.Context, url string, timeout int64, username *string, password *string, jwt *string, apiToken *string, extraHeaders *map[string]string) (*kestra_api_client.APIClient, error) {
	configuration := kestra_api_client.NewConfiguration()
	configuration.HTTPClient = &http.Client{Timeout: time.Duration(timeout) * time.Second}

	configuration.Servers = []kestra_api_client.ServerConfiguration{
		{
			URL: url,
		},
	}

	configuration.DefaultHeader = defaultHeaders(username, password, jwt, apiToken, extraHeaders)

	apiClient := kestra_api_client.NewAPIClient(configuration)

	return apiClient, nil
}

// defaultHeaders builds the headers applied to every request. Later entries win:
// an api token overrides basic auth, and extra headers override both.
func defaultHeaders(username *string, password *string, jwt *string, apiToken *string, extraHeaders *map[string]string) map[string]string {
	headers := map[string]string{}
	if (username != nil) && (password != nil) {
		auth := base64.StdEncoding.EncodeToString([]byte(*username + ":" + *password))
		headers["Authorization"] = "Basic " + auth
	}
	if jwt != nil && *jwt != "" {
		headers["Cookie"] = "JWT=" + *jwt
	}
	if apiToken != nil && *apiToken != "" {
		headers["Authorization"] = "Bearer " + *apiToken
	}
	if extraHeaders != nil {
		for k, v := range *extraHeaders {
			headers[k] = v
		}
	}
	return headers
}

// NewKestraClient builds the hand-written client from the same configuration as NewClient.
// Auth goes through plain headers because WithBasicAuth/WithTokenAuth outrank an
// Authorization entry from extraHeaders, inverting the precedence used above.
func NewKestraClient(url string, timeout int64, username *string, password *string, jwt *string, apiToken *string, extraHeaders *map[string]string) *kestra_api_client.KestraClient {
	return kestra_api_client.NewClient(
		url,
		kestra_api_client.WithHTTPClient(&http.Client{Timeout: time.Duration(timeout) * time.Second}),
		kestra_api_client.WithHeaders(defaultHeaders(username, password, jwt, apiToken, extraHeaders)),
	)
}
