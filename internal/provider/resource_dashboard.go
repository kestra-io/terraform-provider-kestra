package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceDashboard() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Dashboard resource.",

		CreateContext: resourceDashboardCreate,
		ReadContext:   resourceDashboardRead,
		UpdateContext: resourceDashboardUpdate,
		DeleteContext: resourceDashboardDelete,
		Schema: map[string]*schema.Schema{
			"source_code": {
				Description: "The source code text.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"id": {
				Description: "The unique identifier.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	sourceCode := d.Get("source_code").(string)

	req, reqErr := c.yamlRequest("POST", fmt.Sprintf("%s/dashboards", apiRoot(c.TenantId)), &sourceCode)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(req.(map[string]interface{})["id"].(string))

	return diags
}

func resourceDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Id()
	url := fmt.Sprintf("%s/dashboards/%s", apiRoot(c.TenantId), id)

	req, reqErr := c.yamlRequest("GET", url, nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	response := req.(map[string]interface{})
	if err := d.Set("source_code", response["sourceCode"].(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("id", response["id"].(string)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("source_code") {
		id := d.Id()
		sourceCode := d.Get("source_code").(string)
		url := fmt.Sprintf("%s/dashboards/%s", apiRoot(c.TenantId), id)

		_, reqErr := c.yamlRequest("PUT", url, &sourceCode)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		return diags
	}
	return resourceDashboardRead(ctx, d, meta)
}

func resourceDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Id()
	url := fmt.Sprintf("%s/dashboards/%s", apiRoot(c.TenantId), id)

	_, reqErr := c.request("DELETE", url, nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")
	return diags
}
