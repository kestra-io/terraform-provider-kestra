package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
	"strings"
)

func namespaceSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	body["id"] = d.Get("namespace_id").(string)
	body["name"] = d.Get("name").(string)

	variables := make(map[string]interface{}, 0)
	err := yaml.Unmarshal([]byte(d.Get("variables").(string)), &variables)
	if err != nil {
		return nil, err
	}
	body["variables"] = variables

	var taskDefaults interface{}
	err = yaml.Unmarshal([]byte(d.Get("task_defaults").(string)), &taskDefaults)
	if err != nil {
		return nil, err
	}
	body["taskDefaults"] = taskDefaults

	return body, nil
}

func namespaceApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("namespace_id", r["id"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["name"]; ok {
		if err := d.Set("name", r["name"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["variables"]; ok {
		toYaml, err := toYaml(r["variables"].(map[string]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("variables", toYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["taskDefaults"]; ok {
		toYaml, err := toYaml(r["taskDefaults"].(interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("task_defaults", toYaml); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func namespaceConvertSecretId(id string) (string, string) {
	splits := strings.Split(id, "_")

	return splits[0], strings.Join(splits[1:], "_")
}

func namespaceSecretSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	body[d.Get("secret_key").(string)] = d.Get("secret_value").(string)

	return body, nil
}
