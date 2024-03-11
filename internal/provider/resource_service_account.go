package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceServiceAccount() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Service Account.",

		CreateContext: resourceServiceAccountCreate,
		ReadContext:   resourceServiceAccountRead,
		UpdateContext: resourceServiceAccountUpdate,
		DeleteContext: resourceServiceAccountDelete,
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "The service account name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The service account description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"group": {
				Description: "The service account group.",
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"group_id": {
							Description: "The group id.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"tenant_id": {
							Description: "The tenant id for this group.",
							Type:        schema.TypeString,
							Optional:    true,
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

func resourceServiceAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := serviceAccountSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", fmt.Sprintf("%s/users/service-accounts", apiRoot(nil)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := serviceAccountApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceServiceAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("%s/users/%s", apiRoot(nil), id), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := serviceAccountApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceServiceAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("username", "description", "groups") {
		body, err := serviceAccountSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		id := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/users/service-accounts/%s", apiRoot(nil), id), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := serviceAccountApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceServiceAccountRead(ctx, d, meta)
	}
}

func resourceServiceAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	id := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/users/%s", apiRoot(nil), id), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
