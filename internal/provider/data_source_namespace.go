package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNamespace() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Namespace.",

		ReadContext: dataSourceNamespaceRead,
		Schema: map[string]*schema.Schema{
			"namespace_id": {
				Description: "The namespace.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The namespace friendly description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"variables": {
				Description: "The namespace variables.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"task_defaults": {
				Description: "The namespace task defaults.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId := d.Get("namespace_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("/api/v1/namespaces/%s", namespaceId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := namespaceApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
