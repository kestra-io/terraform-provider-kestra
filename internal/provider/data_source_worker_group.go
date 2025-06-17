package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceWorkerGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra Worker Group." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "The worker group id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key": {
				Description: "The worker group key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The worker group description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"allowed_tenants": {
				Description: "The list of tenants allowed to use the worker group.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}
