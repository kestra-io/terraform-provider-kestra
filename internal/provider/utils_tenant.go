package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func tenantSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	body["id"] = d.Get("tenant_id").(string)
	body["name"] = d.Get("name").(string)

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

	return diags
}
