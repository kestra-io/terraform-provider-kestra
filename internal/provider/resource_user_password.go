package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceUserPassword() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra User Basic Auth Password.",

		CreateContext: resourceUserPasswordCreate,
		ReadContext:   resourceUserPasswordRead,
		UpdateContext: resourceUserPasswordUpdate,
		DeleteContext: resourceUserPasswordDelete,
		Schema: map[string]*schema.Schema{
			"user_id": {
				Description: "The user id.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"password": {
				Description: "The user password.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceUserPasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceUserPasswordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	body, err := userPasswordSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	userId := d.Get("user_id").(string)

	r, reqErr := c.request("PUT", fmt.Sprintf("/api/v1/users/%s/password", userId), body)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	d.SetId(r.(map[string]interface{})["id"].(string))

	return diags
}

func resourceUserPasswordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("password") {
		body, err := userPasswordSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		userId := d.Id()

		r, reqErr := c.request("PUT", fmt.Sprintf("/api/v1/users/%s/password", userId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		d.SetId(r.(map[string]interface{})["id"].(string))

		return diags
	} else {
		return diags
	}
}

func resourceUserPasswordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	d.SetId("")

	return diags
}
