package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUserApiTokens() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about the API tokens of a Kestra User." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceUserApiTokensRead,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "The ID of the user owning the API Tokens.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"api_tokens": {
				Description: "The API tokens of the user.",
				Type:        schema.TypeSet,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"token_id": {
							Description: "The API token id.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"name": {
							Description: "The API token display name.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"description": {
							Description: "The API token description.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"token_prefix": {
							Description: "The API token prefix.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"iat": {
							Description: "The API token issued at time.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"exp": {
							Description: "The API token expiration time.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
						},
						"last_used": {
							Description: "The last time this API token was used.",
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
						},
						"extended": {
							Description: "Flag indicating whether this API token duration is extended.",
							Type:        schema.TypeBool,
							Computed:    true,
							Optional:    true,
						},
						"expired": {
							Description: "Flag indicating whether this API token has expired.",
							Type:        schema.TypeBool,
							Computed:    true,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceUserApiTokensRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Get("user_id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/users/%s/api-tokens", apiRoot(nil), id), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := userApiTokensToSchema(id, r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
