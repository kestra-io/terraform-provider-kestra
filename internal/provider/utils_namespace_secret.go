package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
