package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"net/http"
)

func resourceSecurityIntegration() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Security Integration. Note that when imported URI and secret token are not provided.",

		CreateContext: resourceSecurityIntegrationCreate,
		ReadContext:   resourceSecurityIntegrationRead,
		DeleteContext: resourceSecurityIntegrationDelete,
		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "The unique identifier of the security integration.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The name of the security integration.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description:  "The type of the security integration.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"SCIM"}, false),
				ForceNew:     true,
			},
			"description": {
				Description: "The description of the security integration.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"uri": {
				Description: "The url of the security integration.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret_token": {
				Description: "The secret token of the security integration.",
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

func resourceSecurityIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body := map[string]interface{}{
		"name":        d.Get("name").(string),
		"type":        d.Get("type").(string),
		"description": d.Get("description").(string),
	}

	tenantId := c.TenantId

	r, reqErr := c.request("POST", fmt.Sprintf("%s/security-integrations", apiRoot(tenantId)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	securityIntegration := r.(map[string]interface{})
	log.Printf("[DEBUG] Result: %+v", securityIntegration)

	d.SetId(securityIntegration["id"].(string))
	err := d.Set("uid", securityIntegration["id"].(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("uri", securityIntegration["uri"].(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("secret_token", securityIntegration["apiToken"].(map[string]interface{})["fullToken"].(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecurityIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	integrationId := d.Id()
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/security-integrations/%s", apiRoot(tenantId), integrationId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	securityIntegration := r.(map[string]interface{})

	d.SetId(securityIntegration["id"].(string))
	err := d.Set("uid", securityIntegration["id"].(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("name", securityIntegration["name"].(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("type", securityIntegration["type"].(string))
	if err != nil {
		return diag.FromErr(err)
	}
	err = d.Set("description", securityIntegration["description"].(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceSecurityIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	integrationId := d.Id()
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/security-integrations/%s", apiRoot(tenantId), integrationId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}
