package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra User." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user_id": {
				Description: "The user.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"username": {
				Description: "The user name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The user description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"first_name": {
				Description: "The user first name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_name": {
				Description: "The user last name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "The user email.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"groups": {
				Description: "The user global roles in yaml string.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	userId := d.Get("user_id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/users/%s", apiRoot(tenantId), userId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := userApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
