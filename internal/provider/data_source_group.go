package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Group.",

		ReadContext: dataSourceGroupRead,
		Schema: map[string]*schema.Schema{
			"group_id": {
				Description: "The group.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The group name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The group description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"global_roles": {
				Description: "The group global roles in yaml string.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"namespace_roles": {
				Description: "The group namespace roles in yaml string.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:     schema.TypeString,
							Required: true,
						},

						"roles": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	groupId := d.Get("group_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/groups/%s", groupId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := groupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
