package provider_v2

import (
	"testing"
)

func TestUnitPolicyBasePath(t *testing.T) {
	tests := []struct {
		name      string
		scope     string
		tenantId  string
		namespace string
		expected  string
	}{
		{"instance", policyScopeInstance, "", "", "/api/v1/instance/policies"},
		{"tenant", policyScopeTenant, "main", "", "/api/v1/main/policies"},
		{"namespace", policyScopeNamespace, "main", "io.kestra.terraform", "/api/v1/main/namespaces/io.kestra.terraform/policies"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policyBasePath(tt.scope, tt.tenantId, tt.namespace); got != tt.expected {
				t.Errorf("policyBasePath() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

func TestUnitParsePolicyContent(t *testing.T) {
	document, id, err := parsePolicyContent(`
id: my-policy
description: some policy
rules:
  - type: io.kestra.plugin.ee.rules.Deny
    on: PLUGIN
    where:
      - field: type
        operator: EQUAL_TO
        value: io.kestra.plugin.core.log.Log
`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if id != "my-policy" {
		t.Errorf("expected id my-policy, got %s", id)
	}
	rules, ok := document["rules"].([]interface{})
	if !ok || len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %v", document["rules"])
	}
	rule, ok := rules[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected a map rule, got %T", rules[0])
	}
	// regression: with YAML 1.1 parsing, the unquoted `on` key resolves to a boolean key
	if rule["on"] != "PLUGIN" {
		t.Errorf("expected the `on` key to be preserved as a string key, got rule: %v", rule)
	}
}

func TestUnitParsePolicyContentInvalid(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"not yaml", "\t- broken"},
		{"not a document", "- type: io.kestra.plugin.ee.rules.Deny"},
		{"missing id", "description: no id here"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, err := parsePolicyContent(tt.content); err == nil {
				t.Errorf("expected an error for content: %q", tt.content)
			}
		})
	}
}

func TestUnitJsonSemanticallyEqual(t *testing.T) {
	// YAML integers must compare equal to JSON float64 numbers
	yamlSide := []interface{}{map[string]interface{}{"max": 3, "on": "FLOW"}}
	jsonSide := []interface{}{map[string]interface{}{"max": 3.0, "on": "FLOW"}}
	if !jsonSemanticallyEqual(yamlSide, jsonSide) {
		t.Errorf("expected %v and %v to be semantically equal", yamlSide, jsonSide)
	}
	if jsonSemanticallyEqual(yamlSide, []interface{}{map[string]interface{}{"max": 4.0, "on": "FLOW"}}) {
		t.Errorf("expected different values to not be semantically equal")
	}
}

func TestUnitPolicyModelContains(t *testing.T) {
	configured := map[string]interface{}{
		"id": "my-policy",
		"rules": []interface{}{map[string]interface{}{
			"type": "io.kestra.plugin.ee.rules.Add",
			"on":   "PLUGIN",
			"values": map[string]interface{}{
				"options": map[string]interface{}{"readTimeout": "PT30S"},
			},
		}},
	}
	// the server stamps scope/enforcement and fills plugin defaults the user never authored
	serverEnriched := map[string]interface{}{
		"id":          "my-policy",
		"scope":       "TENANT",
		"tenantId":    "main",
		"enforcement": "ACTIVE",
		"deleted":     false,
		"rules": []interface{}{map[string]interface{}{
			"type":     "io.kestra.plugin.ee.rules.Add",
			"on":       "PLUGIN",
			"override": false,
			"values": map[string]interface{}{
				"options": map[string]interface{}{"readTimeout": "PT30S"},
			},
		}},
	}
	if !policyModelContains(serverEnriched, configured) {
		t.Errorf("expected server-side stamped fields and defaults to be ignored")
	}

	serverChangedValue := map[string]interface{}{
		"id": "my-policy",
		"rules": []interface{}{map[string]interface{}{
			"type": "io.kestra.plugin.ee.rules.Add",
			"on":   "PLUGIN",
			"values": map[string]interface{}{
				"options": map[string]interface{}{"readTimeout": "PT60S"},
			},
		}},
	}
	if policyModelContains(serverChangedValue, configured) {
		t.Errorf("expected a changed authored value to be reported as drift")
	}

	serverDroppedRule := map[string]interface{}{
		"id":    "my-policy",
		"rules": []interface{}{},
	}
	if policyModelContains(serverDroppedRule, configured) {
		t.Errorf("expected a dropped rule to be reported as drift")
	}
}

func TestUnitPolicyModelToYaml(t *testing.T) {
	res := map[string]interface{}{
		"id":          "my-policy",
		"scope":       "TENANT",
		"tenantId":    "main",
		"enforcement": "ACTIVE",
		"deleted":     false,
		"source":      nil,
		"rules": []interface{}{map[string]interface{}{
			"type": "io.kestra.plugin.ee.rules.Deny",
			"on":   "PLUGIN",
		}},
	}
	rendered, err := policyModelToYaml(res)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	document, id, err := parsePolicyContent(rendered)
	if err != nil {
		t.Fatalf("rendered YAML does not parse back: %s", err)
	}
	if id != "my-policy" {
		t.Errorf("expected id my-policy, got %s", id)
	}
	for _, stripped := range []string{"scope", "tenantId", "namespace", "deleted", "source"} {
		if _, present := document[stripped]; present {
			t.Errorf("expected server-side field %q to be stripped, got: %v", stripped, document)
		}
	}
}
