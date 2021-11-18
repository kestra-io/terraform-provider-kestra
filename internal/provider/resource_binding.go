package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceBinding() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Binding.",

		CreateContext: resourceBindingCreate,
		ReadContext:   resourceBindingRead,
		DeleteContext: resourceBindingDelete,
		Schema: map[string]*schema.Schema{
			"type": {
				Description: "The binding type.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"external_id": {
				Description: "The binding external id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"role_id": {
				Description: "The role id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceBindingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := bindingSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", "/api/v1/bindings", body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := bindingApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	bindingId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/bindings/%s", bindingId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := bindingApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceBindingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	bindingId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("/api/v1/bindings/%s", bindingId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
