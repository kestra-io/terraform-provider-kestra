package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

func namespaceSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	body["id"] = d.Get("namespace_id").(string)
	body["description"] = d.Get("description").(string)

	variables := make(map[string]interface{}, 0)
	err := yaml.Unmarshal([]byte(d.Get("variables").(string)), &variables)
	if err != nil {
		return nil, err
	}
	body["variables"] = variables

	var pluginDefaults interface{}
	err = yaml.Unmarshal([]byte(d.Get("plugin_defaults").(string)), &pluginDefaults)
	if err != nil {
		return nil, err
	}
	body["pluginDefaults"] = pluginDefaults

	return body, nil
}

func namespaceApiToSchema(r map[string]interface{}, d *schema.ResourceData, c *Client) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))
	if *c.TenantId != "" {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("namespace_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["description"]; ok {
		if err := d.Set("description", r["description"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["variables"]; ok {
		toYaml, err := toYaml(r["variables"].(map[string]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		if pointerToString(toYaml) != "{}\n" {
			if err := d.Set("variables", toYaml); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if _, ok := r["pluginDefaults"]; ok {
		toYaml, err := toYaml(r["pluginDefaults"].(interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("plugin_defaults", toYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func namespaceSecretSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	secret := make(map[string]interface{}, 0)
	secret["key"] = d.Get("secret_key").(string)
	secret["value"] = d.Get("secret_value").(string)
	secret["description"] = d.Get("secret_description").(string)

	tagsByKey := d.Get("secret_tags").(map[string]interface{})
	tags := make([]interface{}, 0, len(tagsByKey))
	for key, value := range tagsByKey {
		tag := make(map[string]interface{}, 0)
		tag["key"] = key
		tag["value"] = value
		tags = append(tags, tag)
	}
	secret["tags"] = tags

	return secret, nil
}
