package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
				Description:  "The role id. Exactly one of `role_id` or `name` must be provided.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"role_id", "name"},
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description:  "The role name. Can be set instead of `role_id` to look up the role.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"role_id", "name"},
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

	if roleId == "" {
		id, err := findRoleIdByName(c, d.Get("name").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		roleId = id

		if err := d.Set("role_id", roleId); err != nil {
			return diag.FromErr(err)
		}
	}

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

func findRoleIdByName(c *Client, name string) (string, error) {
	query := url.Values{
		"q":    {name},
		"size": {"100"},
	}

	r, reqErr := c.request("GET", fmt.Sprintf("%s/roles/search?%s", apiRoot(c.TenantId), query.Encode()), nil)
	if reqErr != nil {
		return "", reqErr.Err
	}

	body, _ := r.(map[string]interface{})
	results, ok := body["results"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response searching roles by name '%s'", name)
	}

	var matches []string
	for _, item := range results {
		role, _ := item.(map[string]interface{})
		roleName, _ := role["name"].(string)
		id, _ := role["id"].(string)

		if id != "" && strings.EqualFold(roleName, name) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no role found with name '%s'", name)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple roles found with name '%s', use `role_id` instead", name)
	}
}
