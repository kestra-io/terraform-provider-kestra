package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Role.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The role name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The role description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"permissions": {
				Description: "The role permissions.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "The type of permission.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"permissions": {
							Description: "The permissions for this type.",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
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

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := roleSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", "/api/v1/roles", body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := roleApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	roleId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/roles/%s", roleId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := roleApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("namespace", "name", "description", "permissions") {
		body, err := roleSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		roleId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("/api/v1/roles/%s", roleId), body)
		if err != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := roleApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceRoleRead(ctx, d, meta)
	}
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	roleId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("/api/v1/roles/%s", roleId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
