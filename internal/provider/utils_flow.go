package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func controlContent(body map[string]interface{}, content map[string]interface{}) (map[string]interface{}, error) {
	if val, ok := content["id"]; ok {
		if val != body["id"] {
			return nil, fmt.Errorf("incoherent resource id: %s, yaml content id: %s. You should remove id from yaml content", body["id"], val)
		}
	}

	if val, ok := content["namespace"]; ok {
		if val != body["namespace"] {
			return nil, fmt.Errorf("incoherent resource namespace: %s, yaml content namespace: %s. You should remove namespace from yaml content", body["id"], val)
		}
	}

	delete(content, "id")
	delete(content, "namespace")

	return content, nil
}

func flowApiToSchema(r map[string]interface{}, d *schema.ResourceData, c *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(fmt.Sprintf("%s/%s", r["namespace"].(string), r["id"].(string)))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("namespace", r["namespace"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("flow_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("revision", r["revision"].(json.Number)); err != nil {
		return diag.FromErr(err)
	}

	delete(r, "deleted")
	delete(r, "id")
	delete(r, "namespace")

	toYaml, err := toYaml(r)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("content", toYaml); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
