package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceTenant() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Tenant.",

		CreateContext: resourceTenantCreate,
		ReadContext:   resourceTenantRead,
		UpdateContext: resourceTenantUpdate,
		DeleteContext: resourceTenantDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The tenant name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTenantCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := tenantSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", fmt.Sprintf("%s/tenants", apiRoot("")), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTenantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("%s/tenants/%s", apiRoot(""), tenantId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTenantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("name") {
		body, err := tenantSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		tenantId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/tenants/%s", apiRoot(""), tenantId), body)
		if err != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := tenantApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceTenantRead(ctx, d, meta)
	}
}

func resourceTenantDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/tenants/%s", apiRoot(""), tenantId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
