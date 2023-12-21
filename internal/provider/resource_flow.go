package provider

import (
	"context"
	"fmt"
	"net/http"

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
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
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
			"revision": {
				Description: "The flow revision.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"keep_original_source": {
				Description: "Use the content as source code, keeping comment and indentation.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"content": {
				Description:      "The flow full content in yaml string.",
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: isYamlEqualsFlow,
				StateFunc:        stateFn,
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
	var yamlSourceCode = true

	// GetOkExists is deprecated but the GetOk does not work correctly with boolean
	// Even if you set default to nil for the props, it will always return false if prop isnt set
	localKeepOriginalSource, isLocalSet := d.GetOkExists("keep_original_source")

	if isLocalSet == true {
		yamlSourceCode = localKeepOriginalSource.(bool)
	} else {
		if c.KeepOriginalSource != nil {
			yamlSourceCode = *c.KeepOriginalSource
		}
	}

	tenantId := c.TenantId

	if yamlSourceCode == true {
		r, reqErr := c.yamlRequest("POST", fmt.Sprintf("%s/flows", apiRoot(tenantId)), stringToPointer(d.Get("content").(string)))
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		body, err := flowSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		r, reqErr := c.request("POST", fmt.Sprintf("%s/flows", apiRoot(tenantId)), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := flowApiToSchema(r.(map[string]interface{}), d, c)
		if errs != nil {
			return errs
		}

		return diags
	}
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, flowId := flowConvertId(d.Id())
	yamlSourceCode := d.Get("keep_original_source").(bool)

	tenantId := c.TenantId

	if yamlSourceCode == true {
		r, reqErr := c.yamlRequest("GET", fmt.Sprintf("%s/flows/%s/%s?source=true", apiRoot(tenantId), namespaceId, flowId), nil)
		if reqErr != nil {
			if reqErr.StatusCode == http.StatusNotFound {
				d.SetId("")
				return diags
			}

			return diag.FromErr(reqErr.Err)
		}

		errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)
		if errs != nil {
			return errs
		}

		// Set the keep_original_source value back to ResourceData
		d.Set("keep_original_source", yamlSourceCode)

		return diags
	} else {
		r, reqErr := c.request("GET", fmt.Sprintf("%s/flows/%s/%s", apiRoot(tenantId), namespaceId, flowId), nil)
		if reqErr != nil {
			if reqErr.StatusCode == http.StatusNotFound {
				d.SetId("")
				return diags
			}

			return diag.FromErr(reqErr.Err)
		}

		errs := flowApiToSchema(r.(map[string]interface{}), d, c)
		if errs != nil {
			return errs
		}

		// Set the keep_original_source value back to ResourceData
		d.Set("keep_original_source", yamlSourceCode)

		return diags
	}
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("content") {
		yamlSourceCode := d.Get("keep_original_source").(bool)
		tenantId := c.TenantId

		if yamlSourceCode == true {
			r, reqErr := c.yamlRequest(
				"PUT",
				fmt.Sprintf(
					"%s/flows/%s/%s",
					apiRoot(tenantId),
					d.Get("namespace").(string),
					d.Get("flow_id").(string),
				),
				stringToPointer(d.Get("content").(string)),
			)
			if reqErr != nil {
				return diag.FromErr(reqErr.Err)
			}

			errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)
			if errs != nil {
				return errs
			}

			return diags
		} else {
			body, err := flowSchemaToApi(d)
			if err != nil {
				return diag.FromErr(err)
			}

			namespaceId, flowId := flowConvertId(d.Id())

			r, reqErr := c.request("PUT", fmt.Sprintf("%s/flows/%s/%s", apiRoot(tenantId), namespaceId, flowId), body)
			if reqErr != nil {
				return diag.FromErr(reqErr.Err)
			}

			errs := flowApiToSchema(r.(map[string]interface{}), d, c)
			if errs != nil {
				return errs
			}

			return diags
		}
	} else {
		return resourceFlowRead(ctx, d, meta)
	}
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, flowId := flowConvertId(d.Id())
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/flows/%s/%s", apiRoot(tenantId), namespaceId, flowId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
