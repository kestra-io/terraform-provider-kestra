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
	namespace, provided := d.GetOk("namespace")
	if provided {
		body["namespaceId"] = namespace.(string)
	}

	return body, nil
}

func bindingApiToSchema(binding map[string]interface{}, d *schema.ResourceData, c *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(binding["id"].(string))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("type", binding["type"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if binding["type"].(string) == "GROUP" {
		if groupValue, ok := binding["group"]; ok {
			if groupMap, ok := groupValue.(map[string]interface{}); ok {
				if err := d.Set("external_id", groupMap["id"].(string)); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	} else if binding["type"].(string) == "USER" {
		if userValue, ok := binding["user"]; ok {
			if userMap, ok := userValue.(map[string]interface{}); ok {
				if err := d.Set("external_id", userMap["id"].(string)); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	if roleValue, ok := binding["role"]; ok {
		if roleMap, ok := roleValue.(map[string]interface{}); ok {
			if err := d.Set("role_id", roleMap["id"].(string)); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, ok := binding["namespaceId"]; ok {
		if err := d.Set("namespace", binding["namespaceId"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
