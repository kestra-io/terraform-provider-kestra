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

	body["name"] = d.Get("name").(string)
	body["description"] = d.Get("description").(string)

	namespaceRoles := make(map[string]interface{}, 0)
	stateNamespaceRoles := d.Get("namespace_roles").(*schema.Set)
	for _, value := range stateNamespaceRoles.List() {
		stateNamespaceRole := value.(map[string]interface{})
		namespaceRoles[stateNamespaceRole["namespace"].(string)] = stateNamespaceRole["roles"]
	}
	body["namespaceRoles"] = namespaceRoles

	body["globalRoles"] = d.Get("global_roles").([]interface{})

	return body, nil
}

func groupApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("name", r["name"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["description"]; ok {
		if err := d.Set("description", r["description"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["namespaceRoles"]; ok {
		apiNamespaceRoles := r["namespaceRoles"].(map[string]interface{})
		var stateNamespaceRoles []map[string]interface{}
		for namespace, value := range apiNamespaceRoles {
			stateNamespaceRoles = append(stateNamespaceRoles, map[string]interface{}{
				"namespace": namespace,
				"roles":     value,
			})
		}

		if err := d.Set("namespace_roles", stateNamespaceRoles); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["globalRoles"]; ok {
		if err := d.Set("global_roles", r["globalRoles"].([]interface{})); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}
