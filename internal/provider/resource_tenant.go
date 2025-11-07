package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceTenant() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Tenant." +
			EnterpriseEditionDescription,

		CreateContext: resourceTenantCreate,
		ReadContext:   resourceTenantRead,
		UpdateContext: resourceTenantUpdate,
		DeleteContext: resourceTenantDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The tenant name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"worker_group": {
				Description: "The worker group.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Description: "The worker group key.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"fallback": {
							Description:  "The fallback strategy.",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"FAIL", "WAIT", "CANCEL"}, false),
						},
					},
				},
			},
			"storage_type": {
				Description: "The storage type.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"storage_configuration": {
				Description: "The storage configuration.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_isolation": {
				Description: "Storage isolation configuration.",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Enable storage isolation.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"denied_services": {
							Description: "List of denied services for isolation.",
							Type:        schema.TypeList,
							Optional:    true,
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
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Description: "Enable secret isolation.",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"denied_services": {
							Description: "List of denied services for secret isolation.",
							Type:        schema.TypeList,
							Optional:    true,
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
				Optional:    true,
			},
			"secret_read_only": {
				Description: "Whether secrets are read-only in this tenant.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"secret_configuration": {
				Description: "The secret configuration.",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"require_existing_namespace": {
				Description: "Whether tenant requires an existing namespace.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"outputs_in_internal_storage": {
				Description: "Whether outputs are stored in internal storage.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceTenantCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := tenantSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", fmt.Sprintf("%s/tenants", apiRoot(nil)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTenantRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Id()

	r, reqErr := c.request("GET", fmt.Sprintf("%s/tenants/%s", apiRoot(nil), tenantId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := tenantApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceTenantUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("name", "storage_configuration", "storage_isolation", "secret_isolation", "secret_type", "secret_read_only", "secret_configuration", "require_existing_namespace", "outputs_in_internal_storage") {
		body, err := tenantSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		tenantId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/tenants/%s", apiRoot(nil), tenantId), body)
		if err != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := tenantApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceTenantRead(ctx, d, meta)
	}
}

func resourceTenantDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := d.Id()

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/tenants/%s", apiRoot(nil), tenantId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
