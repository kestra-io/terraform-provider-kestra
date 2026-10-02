package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ResourceFlow struct{}

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description:   "Manages a Kestra Flow.",
		CreateContext: resourceFlowCreate,
		ReadContext:   resourceFlowRead,
		UpdateContext: resourceFlowUpdate,
		DeleteContext: resourceFlowDelete,
		Identity: &schema.ResourceIdentity{
			SchemaFunc: func() map[string]*schema.Schema {
				return map[string]*schema.Schema{
					"namespace": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
					"flow_id": {
						Type:              schema.TypeString,
						RequiredForImport: true,
					},
				}
			},
		},
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
				Optional:    true,
				ForceNew:    true,
			},
			"flow_id": {
				Description: "The flow id.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"revision": {
				Description: "The flow revision.",
				Type:        schema.TypeInt,
				Computed:    true,
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
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				if d.Id() != "" {
					return []*schema.ResourceData{d}, nil
				}

				identity, err := d.Identity()
				if err != nil {
					return nil, fmt.Errorf("error getting flow identity: %w", err)
				}

				namespace, ok := identity.Get("namespace").(string)
				if !ok || namespace == "" {
					return nil, fmt.Errorf("flow identity attribute %q must be a non-empty string", "namespace")
				}

				flowID, ok := identity.Get("flow_id").(string)
				if !ok || flowID == "" {
					return nil, fmt.Errorf("flow identity attribute %q must be a non-empty string", "flow_id")
				}

				if err := d.Set("namespace", namespace); err != nil {
					return nil, fmt.Errorf("error setting flow namespace: %w", err)
				}
				if err := d.Set("flow_id", flowID); err != nil {
					return nil, fmt.Errorf("error setting flow id: %w", err)
				}

				d.SetId(fmt.Sprintf("%s/%s", namespace, flowID))
				return []*schema.ResourceData{d}, nil
			},
		},
	}
}

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	var yamlSourceCode = *c.KeepOriginalSource

	tenantId := c.TenantId

	diags = validateFlow(c, d.Get("content").(string), diags)

	if yamlSourceCode == true {
		// if namespace or flow_id not provided, extract them from content
		if d.Get("namespace").(string) == "" || d.Get("flow_id").(string) == "" {
			ns, id, err := extractNamespaceAndIdFromContent(d.Get("content").(string))
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}
			_ = d.Set("namespace", ns)
			_ = d.Set("flow_id", id)
		}

		r, reqErr := c.yamlRequest("POST", fmt.Sprintf("%s/flows", apiRoot(tenantId)), stringToPointer(d.Get("content").(string)))
		if reqErr != nil {
			return append(diags, diag.FromErr(reqErr.Err)...)
		}

		errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)

		diags = append(diags, setFlowIdentity(d)...)
		if errs != nil {
			return append(diags, errs...)
		}

		return diags
	} else {
		body, err := flowSchemaToApi(d)
		if err != nil {
			return append(diags, diag.FromErr(err)...)
		}

		r, reqErr := c.request("POST", fmt.Sprintf("%s/flows", apiRoot(tenantId)), body)
		if reqErr != nil {
			return append(diags, diag.FromErr(reqErr.Err)...)
		}

		errs := flowApiToSchema(r.(map[string]interface{}), d, c)

		diags = append(diags, setFlowIdentity(d)...)
		if errs != nil {
			return append(diags, errs...)
		}

		// Add a warning for JSON creation deprecation
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Deprecation warning",
			Detail:   "Creating flow not using the YAML source code is deprecated and will be soon removed.",
		})

		return diags
	}
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId, flowId := flowConvertId(d.Id())
	var yamlSourceCode = *c.KeepOriginalSource

	tenantId := c.TenantId

	diags = validateFlow(c, d.Get("content").(string), diags)

	if yamlSourceCode == true {
		r, reqErr := c.yamlRequest("GET", fmt.Sprintf("%s/flows/%s/%s?source=true", apiRoot(tenantId), namespaceId, flowId), nil)
		if reqErr != nil {
			if reqErr.StatusCode == http.StatusNotFound {
				d.SetId("")
				return diags
			}

			return append(diags, diag.FromErr(reqErr.Err)...)
		}

		errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)

		diags = append(diags, setFlowIdentity(d)...)
		if errs != nil {
			return append(diags, errs...)
		}

		return diags
	} else {
		r, reqErr := c.request("GET", fmt.Sprintf("%s/flows/%s/%s", apiRoot(tenantId), namespaceId, flowId), nil)
		if reqErr != nil {
			if reqErr.StatusCode == http.StatusNotFound {
				d.SetId("")
				return diags
			}

			return append(diags, diag.FromErr(reqErr.Err)...)
		}

		errs := flowApiToSchema(r.(map[string]interface{}), d, c)

		diags = append(diags, setFlowIdentity(d)...)
		if errs != nil {
			return append(diags, errs...)
		}

		return diags
	}
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("content") {
		var yamlSourceCode = *c.KeepOriginalSource
		tenantId := c.TenantId

		if yamlSourceCode == true {
			// ensure namespace/flow_id exist by extracting from content if missing
			if d.Get("namespace").(string) == "" || d.Get("flow_id").(string) == "" {
				ns, id, err := extractNamespaceAndIdFromContent(d.Get("content").(string))
				if err != nil {
					return append(diags, diag.FromErr(err)...)
				}
				_ = d.Set("namespace", ns)
				_ = d.Set("flow_id", id)
			}

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
				return append(diags, diag.FromErr(reqErr.Err)...)
			}

			errs := flowSourceApiToSchema(r.(map[string]interface{}), d, c)

			diags = append(diags, setFlowIdentity(d)...)
			if errs != nil {
				return append(diags, errs...)
			}

			return diags
		} else {
			body, err := flowSchemaToApi(d)
			if err != nil {
				return append(diags, diag.FromErr(err)...)
			}

			namespaceId, flowId := flowConvertId(d.Id())

			r, reqErr := c.request("PUT", fmt.Sprintf("%s/flows/%s/%s", apiRoot(tenantId), namespaceId, flowId), body)
			if reqErr != nil {
				return append(diags, diag.FromErr(reqErr.Err)...)
			}

			errs := flowApiToSchema(r.(map[string]interface{}), d, c)

			diags = append(diags, setFlowIdentity(d)...)
			if errs != nil {
				return append(diags, errs...)
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
		return append(diags, diag.FromErr(reqErr.Err)...)
	}

	d.SetId("")

	return diags
}

func setFlowIdentity(d *schema.ResourceData) diag.Diagnostics {
	identity, err := d.Identity()
	if err != nil {
		return diag.FromErr(err)
	}

	if err := identity.Set("namespace", d.Get("namespace").(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := identity.Set("flow_id", d.Get("flow_id").(string)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func validateFlow(client *Client, content string, diags diag.Diagnostics) diag.Diagnostics {
	if *client.KeepOriginalSource && len(content) > 0 {
		// Call the /flows/validate endpoint
		r, reqErr := client.yamlRequest("POST", fmt.Sprintf("%s/flows/validate", apiRoot(client.TenantId)), stringToPointer(content))

		if strings.Contains(fmt.Sprintf("%s", r), "Unable to validate the flow") {
			// this can happen if the flow syntax is invalid, or if we try to validate a flow without id and namespace
			return append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unable to validate the flow",
				Detail:   fmt.Sprintf("%s", r),
			})
		}
		if reqErr != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error validating flow",
				Detail:   reqErr.Err.Error(),
			})
		}

		// Get the first item in the result array as we only sent 1 flow
		validationResult := r.([]interface{})[0].(map[string]interface{})

		if value, ok := validationResult["constraints"]; ok {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Validation Constraint",
				Detail:   value.(string),
			})
		}

		if value, ok := validationResult["warnings"]; ok {
			warnings := value.([]interface{})
			for _, warning := range warnings {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Validation Warning",
					Detail:   warning.(string),
				})
			}
		}

		if value, ok := validationResult["deprecationPaths"]; ok {
			deprecationPaths := value.([]interface{})
			for _, deprecation := range deprecationPaths {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Deprecation Warning",
					Detail:   fmt.Sprintf("%s is deprecated", deprecation.(string)),
				})
			}
		}

		if value, ok := validationResult["infos"]; ok {
			infos := value.([]interface{})
			for _, info := range infos {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Validation Info",
					Detail:   info.(string),
				})
			}
		}
	} else {
		if !*client.KeepOriginalSource {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Validation Skipped",
				Detail:   "Flow validation has been skipped as it is only compatible with KeepOriginalSource o",
			})
		}
		if len(content) > 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Validation Skipped",
				Detail:   "Flow validation has been skipped as content was empty",
			})
		}
	}

	return diags
}
