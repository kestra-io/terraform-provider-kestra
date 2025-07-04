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
		Description: "Manages a Kestra Service Account." +
			EnterpriseEditionDescription,

		CreateContext: resourceServiceAccountCreate,
		ReadContext:   resourceServiceAccountRead,
		UpdateContext: resourceServiceAccountUpdate,
		DeleteContext: resourceServiceAccountDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The service account name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The service account description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"super_admin": {
				Description: "Is the service account a super admin.",
				Type:        schema.TypeBool,
				Optional:    true,
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
							Required:    true,
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

	tenantId := c.TenantId

	r, reqErr := c.request("POST", fmt.Sprintf("%s/service-accounts", apiRoot(tenantId)), body)
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
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/service-accounts/%s", apiRoot(tenantId), id), nil)
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

	if d.HasChanges("name", "description", "groups") {
		body, err := serviceAccountSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		id := d.Id()
		tenantId := c.TenantId

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/service-accounts/%s", apiRoot(tenantId), id), body)
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
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/service-accounts/%s", apiRoot(tenantId), id), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
