package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
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
				DiffSuppressFunc: isYamlEqualsFlow,
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

	r, reqErr := c.request("POST", "/api/v1/templates", body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
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

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
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

		r, reqErr := c.request("PUT", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
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

	_, reqErr := c.request("DELETE", fmt.Sprintf("/api/v1/templates/%s/%s", namespaceId, templateId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
