package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceWorkerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Worker Group." +
			EnterpriseEditionDescription,

		CreateContext: resourceWorkerGroupCreate,
		ReadContext:   resourceWorkerGroupRead,
		UpdateContext: resourceWorkerGroupUpdate,
		DeleteContext: resourceWorkerGroupDelete,
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "The worker group key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The worker group description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allowed_tenants": {
				Description: "The list of tenants allowed to use the worker group.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceWorkerGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := workerGroupSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", fmt.Sprintf("%s/instance/workergroups", apiRoot(nil)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := workerGroupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceWorkerGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	workerGroupId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("%s/instance/workergroups/%s", apiRoot(nil), workerGroupId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := workerGroupApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceWorkerGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("key", "description", "allowed_tenants") {
		body, err := workerGroupSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		workerGroupId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/instance/workergroups/%s", apiRoot(nil), workerGroupId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := workerGroupApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceWorkerGroupRead(ctx, d, meta)
	}
}

func resourceWorkerGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	workerGroupId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/instance/workergroups/%s", apiRoot(nil), workerGroupId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
