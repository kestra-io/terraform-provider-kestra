package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Template.",

		CreateContext: resourceTemplateCreate,
		ReadContext:   resourceTemplateRead,
		UpdateContext: resourceTemplateUpdate,
		DeleteContext: resourceTemplateDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Description: "The template namespace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"template_id": {
				Description: "The template id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"content": {
				Description:      "The template full content in yaml string.",
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

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := templateSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.request("POST", "/api/v1/templates", body)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := templateApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, templateId := templateConvertId(d.Id())

	r, err := c.request("GET", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := templateApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("content") {
		body, err := templateSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceId, templateId := templateConvertId(d.Id())

		r, err := c.request("PUT", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), body)
		if err != nil {
			return diag.FromErr(err)
		}

		errs := templateApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceTemplateRead(ctx, d, meta)
	}
}

func resourceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, templateId := templateConvertId(d.Id())

	_, err := c.request("DELETE", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
