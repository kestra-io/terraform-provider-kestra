package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNamespaceFile() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Namespace File",

		ReadContext: dataSourceNamespaceFileRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"namespace": {
				Description: "The namespace of the namespace file resource.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"destination_path": {
				Description: "The path to the namespace file.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			"content": {
				Description: "Content to store in the file, expected to be a UTF-8 encoded string.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func dataSourceNamespaceFileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	namespace := d.Get("namespace").(string)
	filename := d.Get("filename").(string)
	tenantId := d.Get("tenant_id").(string)

	url := c.Url + fmt.Sprintf("%s/files/namespaces/%s?path=%s", apiRoot(tenantId), namespace, filename)

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(url), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	statusCode, body, reqErr := c.rawResponseRequest("GET", req)
	if reqErr != nil {
		if statusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(fmt.Sprintf("%s/%s", namespace, filename))
	if err := d.Set("namespace", namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filename", filename); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("content", string(body)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
