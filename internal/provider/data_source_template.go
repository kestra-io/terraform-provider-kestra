package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTemplate() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Template",

		ReadContext: dataSourceTemplateRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"namespace": {
				Description: "The namespace.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"template_id": {
				Description: "The template id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"content": {
				Description: "The template content as yaml.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId := d.Get("namespace").(string)
	templateId := d.Get("template_id").(string)
	tenantId := d.Get("tenant_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("%s/templates/%s/%s", apiRoot(tenantId), namespaceId, templateId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := templateApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
