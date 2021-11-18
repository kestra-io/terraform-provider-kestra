package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func bindingSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["type"] = d.Get("type").(string)
	body["externalId"] = d.Get("external_id").(string)
	body["roleId"] = d.Get("role_id").(string)
	body["namespaceId"] = d.Get("namespace").(string)

	return body, nil
}

func bindingApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	binding := r["binding"].(map[string]interface{})

	d.SetId(binding["id"].(string))

	if err := d.Set("type", binding["type"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("external_id", binding["externalId"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("role_id", binding["roleId"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := binding["namespaceId"]; ok {
		if err := d.Set("namespace", binding["namespaceId"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
