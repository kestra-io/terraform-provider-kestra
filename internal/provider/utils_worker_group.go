package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func workerGroupSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	key := d.Get("key").(string)
	body["id"] = key

	name := d.Get("name").(string)
	if name == "" {
		name = key
	}
	body["name"] = name

	body["description"] = d.Get("description").(string)
	body["subscriptions"] = []interface{}{}

	return body, nil
}

func workerGroupApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	id := r["id"].(string)
	d.SetId(id)

	if err := d.Set("key", id); err != nil {
		return diag.FromErr(err)
	}

	if name, ok := r["name"]; ok && name != nil {
		if err := d.Set("name", name.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if desc, ok := r["description"]; ok && desc != nil {
		if err := d.Set("description", desc.(string)); err != nil {
			return diag.FromErr(err)
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

	if key, ok := workerGroup["key"].(string); ok {
		workerGroupData["key"] = key
	} else {
		return nil
	}

	if workerGroup["fallback"] != nil {
		workerGroupData["fallback"] = workerGroup["fallback"].(string)
	}

	var workerGroupDataList []map[string]interface{}
	workerGroupDataList = append(workerGroupDataList, workerGroupData)

	return workerGroupDataList
}
