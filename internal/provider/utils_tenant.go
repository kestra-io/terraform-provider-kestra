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

	if secretType := d.Get("secret_type").(string); secretType != "" {
		body["secretType"] = secretType
	}

	if secretConfiguration := d.Get("secret_configuration").(map[string]interface{}); len(secretConfiguration) > 0 {
		body["secretConfiguration"] = secretConfiguration
	}

	return body, nil
}

func tenantApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("tenant_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", r["name"].(string)); err != nil {
		return diag.FromErr(err)
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

	if secretType, ok := r["secretType"].(string); ok {
		if err := d.Set("secret_type", secretType); err != nil {
			return diag.FromErr(err)
		}
	}

	if secretConfiguration, ok := r["secretConfiguration"].(map[string]interface{}); ok {
		if err := d.Set("secret_configuration", secretConfiguration); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
