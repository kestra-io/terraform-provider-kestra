package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceBinding() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra binding",

		ReadContext: dataSourceBindingRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"binding_id": {
				Description: "The binding id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The binding type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"external_id": {
				Description: "The binding external id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The role id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	bindingId := d.Get("binding_id").(string)
	tenantId := d.Get("tenant_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("%s/bindings/%s", apiRoot(tenantId), bindingId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := bindingApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
