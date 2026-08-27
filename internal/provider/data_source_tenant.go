package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceTenant() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Tenant." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceTenantRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The tenant name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"default_worker_selector": {
				Description: "The default routing applied to every task of the tenant that does not define its own.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Description: "The tags used to route to a matching Worker Queue.",
							Type:        schema.TypeList,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"match": {
							Description: "How the tags are matched against a Worker Queue tag set: `ALL` or `ANY`.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"fallback": {
							Description: "The strategy when no worker is available: `FAIL`, `WAIT`, `CANCEL` or `IGNORE`.",
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
				Description: "Whether secrets are read-only in this tenant.",
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
			"require_existing_namespace": {
				Description: "Whether the tenant requires existing namespaces.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"outputs_in_internal_storage": {
				Description: "Whether outputs are stored in internal storage.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceTenantRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Get("tenant_id").(string)

	r, reqErr := c.request("GET", fmt.Sprintf("%s/tenants/%s", apiRoot(nil), tenantId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}
