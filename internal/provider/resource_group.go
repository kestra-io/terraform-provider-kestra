package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Group.",

		CreateContext: resourceGroupCreate,
		ReadContext:   resourceGroupRead,
		UpdateContext: resourceGroupUpdate,
		DeleteContext: resourceGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The group name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The group description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"global_roles": {
				Description: "The group global roles ids.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"namespace_roles": {
				Description: "The namespace roles for this group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The namespace.",
						},

						"roles": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The roles id this namespace.",
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := groupSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", "/api/v1/groups", body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := groupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	groupId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/groups/%s", groupId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := groupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("name", "description", "global_roles", "namespace_roles") {
		body, err := groupSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		groupId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("/api/v1/groups/%s", groupId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := groupApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceGroupRead(ctx, d, meta)
	}
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	groupId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("/api/v1/groups/%s", groupId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
