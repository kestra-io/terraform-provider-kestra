package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func workerGroupSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["key"] = d.Get("key").(string)
	body["description"] = d.Get("description").(string)
	body["allowedTenants"] = d.Get("allowed_tenants").([]interface{})

	return body, nil
}

func workerGroupApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("key", r["key"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["description"]; ok {
		if r["description"].(string) != "" {
			if err := d.Set("description", r["description"].(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, ok := r["allowedTenants"]; ok {
		if len(r["allowedTenants"].([]interface{})) > 0 {
			if err := d.Set("allowed_tenants", r["allowedTenants"].([]interface{})); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return diags
}

func includedWorkerGroupSchemaToApi(workerGroupList []interface{}) map[string]interface{} {
	var workerGroupData = make(map[string]interface{})

	if len(workerGroupList) > 0 {
		workerGroupMap := workerGroupList[0].(map[string]interface{})
		workerGroupData["key"] = workerGroupMap["key"].(string)

		if workerGroupMap["fallback"] != "" {
			workerGroupData["fallback"] = workerGroupMap["fallback"].(string)
		}
	}

	return workerGroupData
}

func includedWorkerGroupApiToList(workerGroup map[string]interface{}) []map[string]interface{} {
	var workerGroupData = make(map[string]interface{})
	workerGroupData["key"] = workerGroup["key"].(string)

	if workerGroup["fallback"] != nil {
		workerGroupData["fallback"] = workerGroup["fallback"].(string)
	}

	var workerGroupDataList []map[string]interface{}
	workerGroupDataList = append(workerGroupDataList, workerGroupData)

	return workerGroupDataList
}
