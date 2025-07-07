package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serviceAccountSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["name"] = d.Get("name").(string)
	body["description"] = d.Get("description").(string)
	body["superAdmin"] = d.Get("super_admin").(bool)

	// Convert group data from schema.Set to a slice of maps
	var groupList []map[string]interface{}
	if groups, ok := d.Get("groups").(*schema.Set); ok {
		for _, item := range groups.List() {
			if schemaGroup, ok := item.(map[string]interface{}); ok {
				var apiGroup = map[string]interface{}{}
				apiGroup["id"] = schemaGroup["id"]
				groupList = append(groupList, apiGroup)
			}
		}
	}
	body["groups"] = groupList

	return body, nil
}

func serviceAccountApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("name", r["name"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["namespaceId"]; ok {
		if r["namespaceId"].(string) != "" {
			if err := d.Set("namespace", r["namespaceId"].(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, ok := r["description"]; ok {
		if r["description"].(string) != "" {
			if err := d.Set("description", r["description"].(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, ok := r["superAdmin"]; ok {
		if err := d.Set("super_admin", r["superAdmin"].(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Convert group data from a slice of maps to a schema.Set
	if groupList, ok := r["groups"].([]interface{}); ok {
		groups := make([]map[string]interface{}, len(groupList))
		for i, apiGroup := range groupList {
			var schemaGroup = map[string]interface{}{}
			schemaGroup["id"] = apiGroup.(map[string]interface{})["id"]
			groups[i] = schemaGroup
		}
		if err := d.Set("groups", groups); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}
