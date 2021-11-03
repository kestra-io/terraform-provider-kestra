package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func userSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)

	if d.Id() != "" {
		body["id"] = d.Id()
	}

	body["username"] = d.Get("username").(string)
	body["namespaceId"] = d.Get("namespace").(string)
	body["description"] = d.Get("description").(string)
	body["firstName"] = d.Get("first_name").(string)
	body["lastName"] = d.Get("last_name").(string)
	body["email"] = d.Get("email").(string)
	body["groups"] = d.Get("groups").([]interface{})

	return body, nil
}

func userApiToSchema(r map[string]interface{}, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId(r["id"].(string))

	if err := d.Set("username", r["username"].(string)); err != nil {
		return diag.FromErr(err)
	}

	if _, ok := r["namespace"]; ok {
		if err := d.Set("namespace", r["namespaceId"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["description"]; ok {
		if err := d.Set("description", r["description"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["firstName"]; ok {
		if err := d.Set("first_name", r["firstName"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["lastName"]; ok {
		if err := d.Set("last_name", r["lastName"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["email"]; ok {
		if err := d.Set("email", r["email"].(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if _, ok := r["groups"]; ok {
		if err := d.Set("groups", r["groups"].([]interface{})); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func userPasswordSchemaToApi(d *schema.ResourceData) (map[string]interface{}, error) {
	body := make(map[string]interface{}, 0)
	body["password"] = d.Get("password").(string)

	return body, nil
}
