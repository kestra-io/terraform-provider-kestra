package provider

import (
	"encoding/json"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func namespaceApiToSchema(r map[string]interface{}, d *schema.ResourceData, c *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("namespace_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if description, ok := r["description"].(string); ok {
		if err := d.Set("description", description); err != nil {
			return diag.FromErr(err)
		}
	}

	if variables, ok := r["variables"].(map[string]interface{}); ok {
		toYaml, err := toYaml(variables)
		if err != nil {
			return diag.FromErr(err)
		}

		if pointerToString(toYaml) != "{}\n" {
			if err := d.Set("variables", toYaml); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if allowedNamespaces, ok := r["allowedNamespaces"].([]interface{}); ok {
		allowedNamespacesList := make([]map[string]interface{}, len(allowedNamespaces))
		for i, ns := range allowedNamespaces {
			nsMap := ns.(map[string]interface{})
			allowedNamespacesList[i] = map[string]interface{}{
				"namespace": nsMap["namespace"].(string),
			}
		}
		if err := d.Set("allowed_namespaces", allowedNamespacesList); err != nil {
			return diag.FromErr(err)
		}
	}

	if workerSelector, ok := r["defaultWorkerSelector"].(map[string]interface{}); ok {
		workerSelectorDataList := includedWorkerSelectorApiToList(workerSelector)

		if err := d.Set("default_worker_selector", workerSelectorDataList); err != nil {
			return diag.FromErr(err)
		}
	}

	if storageType, ok := r["storageType"].(string); ok {
		if err := d.Set("storage_type", storageType); err != nil {
			return diag.FromErr(err)
		}
	}

	if storageConfiguration, ok := r["storageConfiguration"].(map[string]interface{}); ok {
		if err := d.Set("storage_configuration", storageConfiguration); err != nil {
			return diag.FromErr(err)
		}
	}

	if storageIsolation, ok := r["storageIsolation"].(map[string]interface{}); ok {
		storageIsolationMap := make(map[string]interface{})
		if enabled, ok := storageIsolation["enabled"].(bool); ok {
			// only set enabled when it's true to avoid writing default false into state
			if enabled {
				storageIsolationMap["enabled"] = enabled
			}
		}
		if deniedServices, ok := storageIsolation["deniedServices"].([]interface{}); ok {
			if len(deniedServices) > 0 {
				arr := make([]string, len(deniedServices))
				for i, v := range deniedServices {
					arr[i] = v.(string)
				}
				sort.Strings(arr)
				iArr := make([]interface{}, len(arr))
				for i, v := range arr {
					iArr[i] = v
				}
				storageIsolationMap["denied_services"] = iArr
			}
		}
		if len(storageIsolationMap) > 0 {
			if err := d.Set("storage_isolation", []interface{}{storageIsolationMap}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if secretIsolation, ok := r["secretIsolation"].(map[string]interface{}); ok {
		secretIsolationMap := make(map[string]interface{})
		if enabled, ok := secretIsolation["enabled"].(bool); ok {
			if enabled {
				secretIsolationMap["enabled"] = enabled
			}
		}
		if deniedServices, ok := secretIsolation["deniedServices"].([]interface{}); ok {
			if len(deniedServices) > 0 {
				arr := make([]string, len(deniedServices))
				for i, v := range deniedServices {
					arr[i] = v.(string)
				}
				sort.Strings(arr)
				iArr := make([]interface{}, len(arr))
				for i, v := range arr {
					iArr[i] = v
				}
				secretIsolationMap["denied_services"] = iArr
			}
		}
		if len(secretIsolationMap) > 0 {
			if err := d.Set("secret_isolation", []interface{}{secretIsolationMap}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if secretType, ok := r["secretType"].(string); ok {
		if err := d.Set("secret_type", secretType); err != nil {
			return diag.FromErr(err)
		}
	}

	if secretReadOnly, ok := r["secretReadOnly"].(bool); ok {
		if err := d.Set("secret_read_only", secretReadOnly); err != nil {
			return diag.FromErr(err)
		}
	}

	if secretConfiguration, ok := r["secretConfiguration"].(map[string]interface{}); ok {
		// The data source schema declares secret_configuration as map(string),
		// so nested object values from the API are JSON-encoded here to avoid
		// crashing the SDK's type check.
		flat := make(map[string]interface{}, len(secretConfiguration))
		for k, v := range secretConfiguration {
			if s, ok := v.(string); ok {
				flat[k] = s
				continue
			}
			if b, err := json.Marshal(v); err == nil {
				flat[k] = string(b)
			}
		}
		if err := d.Set("secret_configuration", flat); err != nil {
			return diag.FromErr(err)
		}
	}

	if outputs, ok := r["outputsInInternalStorage"].(bool); ok {
		if err := d.Set("outputs_in_internal_storage", outputs); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func namespaceSecretSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	secret := make(map[string]interface{}, 0)
	secret["key"] = d.Get("secret_key").(string)
	secret["value"] = d.Get("secret_value").(string)
	secret["description"] = d.Get("secret_description").(string)

	tagsByKey := d.Get("secret_tags").(map[string]interface{})
	tags := make([]interface{}, 0, len(tagsByKey))
	for key, value := range tagsByKey {
		tag := make(map[string]interface{}, 0)
		tag["key"] = key
		tag["value"] = value
		tags = append(tags, tag)
	}
	secret["tags"] = tags

	return secret, nil
}
