package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceApp() *schema.Resource {
	return &schema.Resource{
		Description: "Manages an App resource.",

		CreateContext: resourceAppCreate,
		ReadContext:   resourceAppRead,
		UpdateContext: resourceAppUpdate,
		DeleteContext: resourceAppDelete,
		Schema: map[string]*schema.Schema{
			"source": {
				Description: "The source text.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"uid": {
				Description: "The unique identifier.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	source := d.Get("source").(string)

	req, reqErr := c.yamlRequest("POST", fmt.Sprintf("%s/apps", apiRoot(c.TenantId)), &source)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(req.(map[string]interface{})["uid"].(string))
	return diags
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Id()
	url := fmt.Sprintf("%s/apps/%s", apiRoot(c.TenantId), id)

	req, reqErr := c.yamlRequest("GET", url, nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	response := req.(map[string]interface{})
	if err := d.Set("source", response["source"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uid", response["uid"].(string)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("source") {
		uid := d.Id()
		source := d.Get("source").(string)
		url := fmt.Sprintf("%s/apps/%s", apiRoot(c.TenantId), uid)

		_, reqErr := c.yamlRequest("PUT", url, &source)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		return diags
	}
	return resourceAppRead(ctx, d, meta)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	uid := d.Id()
	url := fmt.Sprintf("%s/apps/%s", apiRoot(c.TenantId), uid)

	_, reqErr := c.request("DELETE", url, nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")
	return diags
}
