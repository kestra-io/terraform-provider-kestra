package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespaceSecret() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Namespace Secret.",

		CreateContext: resourceNamespaceSecretCreate,
		ReadContext:   resourceNamespaceSecretRead,
		UpdateContext: resourceNamespaceSecretUpdate,
		DeleteContext: resourceNamespaceSecretDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Description: "The namespace.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"secret_key": {
				Description: "The namespace secrey key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"secret_value": {
				Description: "The namespace secrey value.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceNamespaceSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := namespaceSecretSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	namespaceId := d.Get("namespace").(string)
	secretKey := d.Get("secret_key").(string)

	var reqErr *RequestError
	_, reqErr = c.request("PUT", fmt.Sprintf("/api/v1/namespaces/%s/secrets", namespaceId), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(fmt.Sprintf("%s_%s", namespaceId, secretKey))

	return diags
}

func resourceNamespaceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceNamespaceSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("secret_key", "secret_value") {
		body, err := namespaceSecretSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceId, _ := namespaceConvertSecretId(d.Id())

		var reqErr *RequestError
		_, reqErr = c.request("PUT", fmt.Sprintf("/api/v1/namespaces/%s/secrets", namespaceId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		return diags
	} else {
		return resourceNamespaceSecretRead(ctx, d, meta)

	}
}

func resourceNamespaceSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, secretKey := namespaceConvertSecretId(d.Id())

	_, reqErr := c.request("DELETE", fmt.Sprintf("/api/v1/namespaces/%s/secrets/%s", namespaceId, secretKey), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
