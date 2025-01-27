package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Worker Group." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The worker group id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key": {
				Description: "The worker group key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The worker group description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allowed_tenants": {
				Description: "The list of tenants allowed to use the worker group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceWorkerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Get("id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/cluster/workergroups/%s", apiRoot(tenantId), id), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := workerGroupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
