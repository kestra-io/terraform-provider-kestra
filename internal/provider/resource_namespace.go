package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespace() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Namespace.",

		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,
		Schema: map[string]*schema.Schema{
			"namespace_id": {
				Description: "The namespace.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "The namespace friendly name.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"variables": {
				Description:      "The namespace variables in yaml string.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: isYamlEquals,
			},
			"task_defaults": {
				Description:      "The namespace task defaults in yaml string.",
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: isYamlEquals,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := namespaceSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := c.request("POST", "/api/v1/namespaces", body)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := namespaceApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId := d.Id()

	r, err := c.request("GET", fmt.Sprintf("/api/v1/namespaces/%s", namespaceId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	errs := namespaceApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("name", "variables", "task_defaults") {
		body, err := namespaceSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		namespaceId := d.Id()

		r, err := c.request("PUT", fmt.Sprintf("/api/v1/namespaces/%s", namespaceId), body)
		if err != nil {
			return diag.FromErr(err)
		}

		errs := namespaceApiToSchema(r.(map[string]interface{}), d)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceNamespaceRead(ctx, d, meta)

	}
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespaceId := d.Id()

	_, err := c.request("DELETE", fmt.Sprintf("/api/v1/namespaces/%s", namespaceId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
