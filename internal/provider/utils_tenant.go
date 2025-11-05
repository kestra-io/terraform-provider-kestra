package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	body["id"] = d.Get("tenant_id").(string)
	body["name"] = d.Get("name").(string)

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
		storageIsolationArray := storageIsolationList.([]interface{})
		if len(storageIsolationArray) > 0 {
			storageIsolationMap := storageIsolationArray[0].(map[string]interface{})
			storageIsolation := make(map[string]interface{})
			if enabled, ok := storageIsolationMap["enabled"].(bool); ok {
				storageIsolation["enabled"] = enabled
			}
			if ds, ok := storageIsolationMap["denied_services"].([]interface{}); ok && len(ds) > 0 {
				denied := make([]string, len(ds))
				for i, s := range ds {
					denied[i] = s.(string)
				}
				storageIsolation["deniedServices"] = denied
			}
			body["storageIsolation"] = storageIsolation
		}
	}

	if secretIsolationList, ok := d.GetOk("secret_isolation"); ok {
		secretIsolationArr := secretIsolationList.([]interface{})
		if len(secretIsolationArr) > 0 {
			secretIsolationMap := secretIsolationArr[0].(map[string]interface{})
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

	if v, ok := d.GetOk("require_existing_namespace"); ok {
		body["requireExistingNamespace"] = v.(bool)
	}

	if v, ok := d.GetOk("outputs_in_internal_storage"); ok {
		body["outputsInInternalStorage"] = v.(bool)
	}

	return body, nil
}

func tenantApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("tenant_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if name, ok := r["name"].(string); ok {
		if err := d.Set("name", name); err != nil {
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
		if ds, ok := storageIsolation["deniedServices"].([]interface{}); ok {
			arr := make([]interface{}, len(ds))
			for i, v := range ds {
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
		if ds, ok := secretIsolation["deniedServices"].([]interface{}); ok {
			arr := make([]interface{}, len(ds))
			for i, v := range ds {
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

	if requireExisting, ok := r["requireExistingNamespace"].(bool); ok {
		if err := d.Set("require_existing_namespace", requireExisting); err != nil {
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
