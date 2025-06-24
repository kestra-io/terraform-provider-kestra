package sdk_client

import (
	"context"
	"encoding/base64"
	"github.com/kestra-io/client-sdk/go-sdk"
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

	defaultHeaders := map[string]string{}
	if (username != nil) && (password != nil) {
		auth := base64.StdEncoding.EncodeToString([]byte(*username + ":" + *password))
		defaultHeaders["Authorization"] = "Basic " + auth
	}
	if jwt != nil && *jwt != "" {
		cookieHeader := "JWT=" + *jwt
		defaultHeaders["Cookie"] = cookieHeader
	}
	if apiToken != nil && *apiToken != "" {
		defaultHeaders["Authorization"] = "Bearer " + *apiToken
	}
	if extraHeaders != nil {
		for k, v := range *extraHeaders {
			defaultHeaders[k] = v
		}
	}
	configuration.DefaultHeader = defaultHeaders

	apiClient := kestra_api_client.NewAPIClient(configuration)

	return apiClient, nil
}
