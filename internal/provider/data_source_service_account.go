package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Service Account." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceServiceAccountRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The service account id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The service account name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The service account description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"super_admin": {
				Description: "The service account description.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"groups": {
				Description: "The service account group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "The group id.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Get("id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/service-accounts/%s", apiRoot(tenantId), id), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := serviceAccountApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
