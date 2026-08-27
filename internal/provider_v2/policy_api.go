package provider_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/kestra-io/client-sdk/go-sdk/kestra_api_client"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
	// yaml.v3 follows YAML 1.2: an unquoted `on:` key — used by every policy rule — stays
	// a string, while yaml.v2 (YAML 1.1) would resolve it to a boolean key.
	yaml "gopkg.in/yaml.v3"
)

// Policy scopes accepted by the provider. STATIC policies are declared in the Kestra
// configuration and are read-only through the API, so they are not manageable here.
const (
	policyScopeInstance  = "INSTANCE"
	policyScopeTenant    = "TENANT"
	policyScopeNamespace = "NAMESPACE"
)

// policyBasePath returns the collection path for a scope; the scope, tenant and namespace
// are carried by the URL, never by the policy source.
func policyBasePath(scope, tenantId, namespace string) string {
	switch scope {
	case policyScopeInstance:
		return "/api/v1/instance/policies"
	case policyScopeNamespace:
		return fmt.Sprintf("/api/v1/%s/namespaces/%s/policies", url.PathEscape(tenantId), url.PathEscape(namespace))
	default:
		return fmt.Sprintf("/api/v1/%s/policies", url.PathEscape(tenantId))
	}
}

// createPolicy and updatePolicy send the raw YAML source: the API parses and validates it,
// then persists the model with the source stored alongside for verbatim round-trips.
func createPolicy(ctx context.Context, client *kestra_api_client.APIClient, scope, tenantId, namespace, source string) (map[string]interface{}, int, error) {
	return sdk_client.RawYamlRequest(ctx, client, "POST", policyBasePath(scope, tenantId, namespace), source)
}

func readPolicy(ctx context.Context, client *kestra_api_client.APIClient, scope, tenantId, namespace, id string) (map[string]interface{}, int, error) {
	return sdk_client.RawRequest(ctx, client, "GET", policyBasePath(scope, tenantId, namespace)+"/"+url.PathEscape(id), nil)
}

func updatePolicy(ctx context.Context, client *kestra_api_client.APIClient, scope, tenantId, namespace, id, source string) (map[string]interface{}, int, error) {
	return sdk_client.RawYamlRequest(ctx, client, "PUT", policyBasePath(scope, tenantId, namespace)+"/"+url.PathEscape(id), source)
}

func deletePolicy(ctx context.Context, client *kestra_api_client.APIClient, scope, tenantId, namespace, id string) (int, error) {
	_, status, err := sdk_client.RawRequest(ctx, client, "DELETE", policyBasePath(scope, tenantId, namespace)+"/"+url.PathEscape(id), nil)
	return status, err
}

// parsePolicyContent parses the content YAML attribute into a JSON-compatible document and
// returns it with its id, so the configured policy_id can be checked against the source.
func parsePolicyContent(content string) (map[string]interface{}, string, error) {
	var parsed interface{}
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, "", fmt.Errorf("content is not valid YAML: %w", err)
	}
	document, ok := yamlToJSONCompatible(parsed).(map[string]interface{})
	if !ok || len(document) == 0 {
		return nil, "", fmt.Errorf("content must be a YAML document describing the policy (id, rules, ...)")
	}
	id, _ := document["id"].(string)
	if id == "" {
		return nil, "", fmt.Errorf("content must carry a string `id`")
	}
	return document, id, nil
}

// yamlToJSONCompatible rewrites any map[interface{}]interface{} trees produced by the YAML
// decoder into map[string]interface{} trees so they can be marshalled to JSON.
func yamlToJSONCompatible(value interface{}) interface{} {
	switch typed := value.(type) {
	case map[interface{}]interface{}:
		converted := make(map[string]interface{}, len(typed))
		for key, val := range typed {
			converted[fmt.Sprintf("%v", key)] = yamlToJSONCompatible(val)
		}
		return converted
	case map[string]interface{}:
		converted := make(map[string]interface{}, len(typed))
		for key, val := range typed {
			converted[key] = yamlToJSONCompatible(val)
		}
		return converted
	case []interface{}:
		converted := make([]interface{}, len(typed))
		for i, val := range typed {
			converted[i] = yamlToJSONCompatible(val)
		}
		return converted
	default:
		return value
	}
}

// jsonSemanticallyEqual compares two JSON-compatible values after a JSON round-trip, so
// YAML integers and JSON float64 numbers compare equal.
func jsonSemanticallyEqual(a, b interface{}) bool {
	normalizedA, errA := jsonNormalize(a)
	normalizedB, errB := jsonNormalize(b)
	if errA != nil || errB != nil {
		return false
	}
	return reflect.DeepEqual(normalizedA, normalizedB)
}

// policyModelContains reports whether the server policy model contains the configured
// document: same list shapes, and recursively every authored key equal to the server
// value. Keys only present server-side are ignored (stamped or defaulted fields such as
// scope, enforcement or `override: false` on Add rules). Only used for legacy policies
// persisted without a source, where a verbatim comparison is impossible.
func policyModelContains(server, configured interface{}) bool {
	normalizedServer, errServer := jsonNormalize(server)
	normalizedConfigured, errConfigured := jsonNormalize(configured)
	if errServer != nil || errConfigured != nil {
		return false
	}
	return jsonContains(normalizedServer, normalizedConfigured)
}

func jsonContains(server, configured interface{}) bool {
	switch configuredTyped := configured.(type) {
	case map[string]interface{}:
		serverTyped, ok := server.(map[string]interface{})
		if !ok {
			return false
		}
		for key, configuredValue := range configuredTyped {
			serverValue, present := serverTyped[key]
			if !present || !jsonContains(serverValue, configuredValue) {
				return false
			}
		}
		return true
	case []interface{}:
		serverTyped, ok := server.([]interface{})
		if !ok || len(serverTyped) != len(configuredTyped) {
			return false
		}
		for i, configuredValue := range configuredTyped {
			if !jsonContains(serverTyped[i], configuredValue) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(server, configured)
	}
}

func jsonNormalize(value interface{}) (interface{}, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized interface{}
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

// policyModelToYaml renders a policy API payload back to a YAML document, dropping the
// server-side fields that never belong to an authored source. Only used for legacy
// policies persisted without a source.
func policyModelToYaml(res map[string]interface{}) (string, error) {
	document := make(map[string]interface{}, len(res))
	for key, value := range res {
		switch key {
		case "scope", "tenantId", "namespace", "source", "deleted":
			// stamped from the URL or internal — never part of the authored source
		default:
			document[key] = value
		}
	}
	encoded, err := yaml.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("unable to render the policy as YAML: %w", err)
	}
	return string(encoded), nil
}
