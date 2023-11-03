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
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
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

	tenantId := c.TenantId

	r, reqErr := c.request("POST", fmt.Sprintf("%s/templates", apiRoot(tenantId)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := templateApiToSchema(r.(map[string]interface{}), d, c)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, templateId := templateConvertId(d.Id())
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/templates/%s/%s", apiRoot(tenantId), namespaceId, templateId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := templateApiToSchema(r.(map[string]interface{}), d, c)
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
		tenantId := c.TenantId

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/templates/%s/%s", apiRoot(tenantId), namespaceId, templateId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := templateApiToSchema(r.(map[string]interface{}), d, c)
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
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/templates/%s/%s", apiRoot(tenantId), namespaceId, templateId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
