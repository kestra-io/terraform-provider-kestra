package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

func namespaceSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	body["id"] = d.Get("namespace_id").(string)
	body["description"] = d.Get("description").(string)

	variables := make(map[string]interface{}, 0)
	err := yaml.Unmarshal([]byte(d.Get("variables").(string)), &variables)
	if err != nil {
		return nil, err
	}
	body["variables"] = variables

	var pluginDefaults interface{}
	err = yaml.Unmarshal([]byte(d.Get("plugin_defaults").(string)), &pluginDefaults)
	if err != nil {
		return nil, err
	}
	body["pluginDefaults"] = pluginDefaults

	allowedNamespaces := d.Get("allowed_namespaces").([]interface{})
	allowedNamespacesList := make([]map[string]interface{}, len(allowedNamespaces))
	for i, ns := range allowedNamespaces {
		nsMap := ns.(map[string]interface{})
		allowedNamespacesList[i] = map[string]interface{}{
			"namespace": nsMap["namespace"].(string),
		}
	}
	body["allowedNamespaces"] = allowedNamespacesList

	if workerGroup, ok := d.GetOk("worker_group"); ok {
		body["workerGroup"] = includedWorkerGroupSchemaToApi(workerGroup.([]interface{}))
	}

	if storageType := d.Get("storage_type").(string); storageType != "" {
		body["storageType"] = storageType
	}

	if storageConfiguration := d.Get("storage_configuration").(map[string]interface{}); len(storageConfiguration) > 0 {
		body["storageConfiguration"] = storageConfiguration
	}

	if storageIsolationList, ok := d.GetOk("storage_isolation"); ok {
		storageIsolationArr := storageIsolationList.([]interface{})
		if len(storageIsolationArr) > 0 {
			storageIsolationMap := storageIsolationArr[0].(map[string]interface{})
			storageIsolation := make(map[string]interface{})
			if enabled, ok := storageIsolationMap["enabled"].(bool); ok {
				storageIsolation["enabled"] = enabled
			}
			if deniedServices, ok := storageIsolationMap["denied_services"].([]interface{}); ok && len(deniedServices) > 0 {
				denied := make([]string, len(deniedServices))
				for i, s := range deniedServices {
					denied[i] = s.(string)
				}
				storageIsolation["deniedServices"] = denied
			}
			body["storageIsolation"] = storageIsolation
		}
	}

	if secretIsolationList, ok := d.GetOk("secret_isolation"); ok {
		secretIsolationStorage := secretIsolationList.([]interface{})
		if len(secretIsolationStorage) > 0 {
			secretIsolationMap := secretIsolationStorage[0].(map[string]interface{})
			secretIsolation := make(map[string]interface{})
			if enabled, ok := secretIsolationMap["enabled"].(bool); ok {
				secretIsolation["enabled"] = enabled
			}
			if ds, ok := secretIsolationMap["denied_services"].([]interface{}); ok && len(ds) > 0 {
				denied := make([]string, len(ds))
				for i, s := range ds {
					denied[i] = s.(string)
				}
				secretIsolation["deniedServices"] = denied
			}
			body["secretIsolation"] = secretIsolation
		}
	}

	if secretType := d.Get("secret_type").(string); secretType != "" {
		body["secretType"] = secretType
	}

	if v, ok := d.GetOk("secret_read_only"); ok {
		body["secretReadOnly"] = v.(bool)
	}

	if secretConfiguration := d.Get("secret_configuration").(map[string]interface{}); len(secretConfiguration) > 0 {
		body["secretConfiguration"] = secretConfiguration
	}

	if v, ok := d.GetOk("outputs_in_internal_storage"); ok {
		body["outputsInInternalStorage"] = v.(bool)
	}

	return body, nil
}

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

	if pluginDefaults, ok := r["pluginDefaults"].(interface{}); ok {
		toYaml, err := toYaml(pluginDefaults)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("plugin_defaults", toYaml); err != nil {
			return diag.FromErr(err)
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

	if workerGroup, ok := r["workerGroup"].(map[string]interface{}); ok {
		workerGroupDataList := includedWorkerGroupApiToList(workerGroup)

		if err := d.Set("worker_group", workerGroupDataList); err != nil {
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
			storageIsolationMap["enabled"] = enabled
		}
		if deniedServices, ok := storageIsolation["deniedServices"].([]interface{}); ok {
			arr := make([]interface{}, len(deniedServices))
			for i, v := range deniedServices {
				arr[i] = v.(string)
			}
			storageIsolationMap["denied_services"] = arr
		}
		if err := d.Set("storage_isolation", []interface{}{storageIsolationMap}); err != nil {
			return diag.FromErr(err)
		}
	}

	if secretIsolation, ok := r["secretIsolation"].(map[string]interface{}); ok {
		secretIsolationMap := make(map[string]interface{})
		if enabled, ok := secretIsolation["enabled"].(bool); ok {
			secretIsolationMap["enabled"] = enabled
		}
		if deniedServices, ok := secretIsolation["deniedServices"].([]interface{}); ok {
			arr := make([]interface{}, len(deniedServices))
			for i, v := range deniedServices {
				arr[i] = v.(string)
			}
			secretIsolationMap["denied_services"] = arr
		}
		if err := d.Set("secret_isolation", []interface{}{secretIsolationMap}); err != nil {
			return diag.FromErr(err)
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
		if err := d.Set("secret_configuration", secretConfiguration); err != nil {
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
