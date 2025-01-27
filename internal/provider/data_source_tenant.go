package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Tenant." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceTenantRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The tenant name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTenantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Get("tenant_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("%s/tenants/%s", apiRoot(nil), tenantId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
