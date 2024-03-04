package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func roleSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["namespaceId"] = d.Get("namespace").(string)
	body["name"] = d.Get("name").(string)
	body["description"] = d.Get("description").(string)

	permissions := make(map[string]interface{}, 0)
	statePermissions := d.Get("permissions").(*schema.Set)
	for _, value := range statePermissions.List() {
		item := value.(map[string]interface{})
		permissions[item["type"].(string)] = item["permissions"]
	}
	body["permissions"] = permissions
	body["isDefault"] = d.Get("is_default").(bool)

	return body, nil
}

func roleApiToSchema(r map[string]interface{}, d *schema.ResourceData, c *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("name", r["name"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["namespaceId"]; ok {
		if err := d.Set("namespace", r["namespaceId"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["description"]; ok {
		if err := d.Set("description", r["description"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["permissions"]; ok {
		apiPermissions := r["permissions"].(map[string]interface{})
		var permissions []map[string]interface{}
		for typeApi, value := range apiPermissions {
			permissions = append(permissions, map[string]interface{}{
				"type":        typeApi,
				"permissions": value,
			})
		}

		if err := d.Set("permissions", permissions); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["isDefault"]; ok {
		if err := d.Set("is_default", r["isDefault"].(bool)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("is_default", false); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
