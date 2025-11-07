package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNamespace() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Namespace." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceNamespaceRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
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
			"plugin_defaults": {
				Description: "The namespace plugin defaults.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allowed_namespaces": {
				Description: "The allowed namespaces.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Description: "The namespace.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"worker_group": {
				Description: "The worker group.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "The worker group key.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"fallback": {
							Description: "The fallback strategy.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
			"storage_type": {
				Description: "The storage type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"storage_configuration": {
				Description: "The storage configuration.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_isolation": {
				Description: "Storage isolation configuration.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Enable storage isolation.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"denied_services": {
							Description: "List of denied services for isolation.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"secret_isolation": {
				Description: "Secret isolation configuration (same shape as storage_isolation).",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Enable secret isolation.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"denied_services": {
							Description: "List of denied services for secret isolation.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"secret_type": {
				Description: "The secret type.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret_read_only": {
				Description: "Whether secrets are read-only in this namespace.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"secret_configuration": {
				Description: "The secret configuration.",
				Type:        schema.TypeMap,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"outputs_in_internal_storage": {
				Description: "Whether outputs are stored in internal storage.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId := d.Get("namespace_id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/namespaces/%s", apiRoot(tenantId), namespaceId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := namespaceApiToSchema(r.(map[string]interface{}), d, c)
	if errs != nil {
		return errs
	}

	return diags
}
