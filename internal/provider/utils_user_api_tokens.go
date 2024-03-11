package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func userApiTokensToSchema(id string, r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(id)

	if _, ok := r["results"]; ok {
		var results = r["results"].([]interface{})
		tokens := make([]map[string]interface{}, len(results))
		for i, token := range results {
			data := token.(map[string]interface{})
			newToken := map[string]interface{}{
				"token_id":     data["id"].(string),
				"name":         data["name"].(string),
				"description":  data["description"].(string),
				"token_prefix": data["prefix"].(string),
				"iat":          data["iat"].(string),
			}
			// Check if the fields exist in the data map before accessing them
			if v, ok := data["exp"]; ok {
				newToken["exp"] = v.(string)
			}

			if v, ok := data["lastUsed"]; ok {
				newToken["last_used"] = v.(string)
			}
			if v, ok := data["extended"]; ok {
				newToken["extended"] = v.(bool)
			}
			if v, ok := data["expired"]; ok {
				newToken["expired"] = v.(bool)
			}
			tokens[i] = newToken
		}

		if err := d.Set("api_tokens", tokens); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

func userApiTokenToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	if v, ok := r["id"]; ok {
		d.SetId(v.(string))
	}

	if err := d.Set("name", r["name"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("full_token", r["fullToken"].(string)); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func userApiTokenFromSchema(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["name"] = d.Get("name").(string)
	body["description"] = d.Get("description").(string)
	body["maxAge"] = d.Get("max_age").(string)
	body["extended"] = d.Get("extended").(bool)

	return body, nil
}
