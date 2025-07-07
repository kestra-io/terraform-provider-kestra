package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"regexp"
)

func resourceServiceAccountApiToken() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Service Account Api Token." +
			EnterpriseEditionDescription,

		CreateContext: resourceServiceAccountApiTokenCreate,
		ReadContext:   resourceServiceAccountApiTokenRead,
		UpdateContext: resourceServiceAccountApiTokenUpdate,
		DeleteContext: resourceServiceAccountApiTokenDelete,
		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Description: "The ID of the Service Account owning the API Token.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The API token display name.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateDiagFunc: func(v any, p cty.Path) diag.Diagnostics {
					value := v.(string)
					var diags diag.Diagnostics
					regex := `^[a-z0-9]+(?:-[a-z0-9]+)*$`
					if !regexp.MustCompile(regex).MatchString(value) || len(value) > 63 {
						diag := diag.Diagnostic{
							Severity: diag.Error,
							Summary:  "Invalid value for token name",
							Detail: fmt.Sprintf(
								"The value %q provided for 'name' does not meet the required criteria. "+
									"It either does not match the specified regular expression %q or exceeds the maximum allowed length of 63 characters.",
								value,
								regex,
							),
						}
						diags = append(diags, diag)
					}
					return diags
				},
			},
			"description": {
				Description: "The API token description.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"max_age": {
				Description: "The time the token remains valid since creation (ISO 8601 duration format).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"extended": {
				Description: "Specify whether the expiry date is automatically moved forward by max age whenever the token is used.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"full_token": {
				Description: "The full API token.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceServiceAccountApiTokenCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := serviceAccountApiTokenFromSchema(d)
	if err != nil {
		return diag.FromErr(err)
	}

	serviceAccountId := d.Get("service_account_id").(string)
	tenantId := c.TenantId

	r, reqErr := c.request("POST", fmt.Sprintf("%s/service-accounts/%s/api-tokens", apiRoot(tenantId), serviceAccountId), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := serviceAccountApiTokenToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceServiceAccountApiTokenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This operation is no-op, since there's nothing to read
	return nil
}

func resourceServiceAccountApiTokenUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This operation is no-op, since there's nothing to update
	return nil
}

func resourceServiceAccountApiTokenDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tokenId := d.Id()
	serviceAccountId := d.Get("service_account_id").(string)
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/service-accounts/%s/api-tokens/%s", apiRoot(tenantId), serviceAccountId, tokenId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
