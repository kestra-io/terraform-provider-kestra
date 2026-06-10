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
				Description: "The worker group identifier.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The worker group display name.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Description: "The worker group description.",
				Type:        schema.TypeString,
				Optional:    true,
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

	r, reqErr := c.request("POST", fmt.Sprintf("%s/instance/worker-groups", apiRoot(nil)), body)
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

	r, reqErr := c.request("GET", fmt.Sprintf("%s/instance/worker-groups/%s", apiRoot(nil), workerGroupId), nil)
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

	if d.HasChanges("key", "name", "description") {
		body, err := workerGroupSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		workerGroupId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/instance/worker-groups/%s", apiRoot(nil), workerGroupId), body)
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

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/instance/worker-groups/%s", apiRoot(nil), workerGroupId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
