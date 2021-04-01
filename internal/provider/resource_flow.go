package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Flow.",

		CreateContext: resourceFlowCreate,
		ReadContext:   resourceFlowRead,
		UpdateContext: resourceFlowUpdate,
		DeleteContext: resourceFlowDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Description: "The flow namespace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"flow_id": {
				Description: "The flow id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Description:      "The flow full content in yaml string.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: isYamlEquals,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := flowSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.request("POST", "/api/v1/flows", body)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := flowApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, flowId := flowConvertId(d.Id())

	r, err := c.request("GET", fmt.Sprintf("/api/v1/flows/%s/%s", namespaceId, flowId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := flowApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("content") {
		body, err := flowSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceId, flowId := flowConvertId(d.Id())

		r, err := c.request("PUT", fmt.Sprintf("/api/v1/flows/%s/%s", namespaceId, flowId), body)
		if err != nil {
			return diag.FromErr(err)
		}

		errs := flowApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceFlowRead(ctx, d, meta)
	}
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, flowId := flowConvertId(d.Id())

	_, err := c.request("DELETE", fmt.Sprintf("/api/v1/flows/%s/%s", namespaceId, flowId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
