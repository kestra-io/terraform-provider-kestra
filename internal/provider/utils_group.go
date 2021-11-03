package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func groupSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["namespaceId"] = d.Get("namespace").(string)
	body["name"] = d.Get("name").(string)
	body["description"] = d.Get("description").(string)

	return body, nil
}

func groupApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

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

	return diags
}
