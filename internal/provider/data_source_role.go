package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Role." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"role_id": {
				Description: "The role.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The role name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The role description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"is_default": {
				Description: "The role is the default one at user creation. Only one role can be default. Latest create/update to true will be keep as default.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"permissions": {
				Description: "The role permissions.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permissions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	roleId := d.Get("role_id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/roles/%s", apiRoot(tenantId), roleId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := roleApiToSchema(r.(map[string]interface{}), d, c)
	if errs != nil {
		return errs
	}

	return diags
}
