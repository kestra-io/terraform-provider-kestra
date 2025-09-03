package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceKv() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access value for an existing Key-Value pair.",

		ReadContext: dataSourceKvRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace": {
				Description: "The namespace of the Key-Value pair.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description: "The key to fetch value for.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "The type of the value. One of STRING, NUMBER, BOOLEAN, DATETIME, DATE, DURATION, JSON.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"value": {
				Description: "The fetched value.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func dataSourceKvRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := c.TenantId
	namespace := d.Get("namespace").(string)
	key := d.Get("key").(string)

	url := c.Url + fmt.Sprintf("%s/namespaces/%s/kv/%s", apiRoot(tenantId), namespace, key)

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s", url), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	_, body, reqErr := c.rawResponseRequest("GET", req)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	if tenantId != nil {
		if err := d.Set("tenant_id", *tenantId); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%s/%s", namespace, key))

	var kvResponsePtr struct {
		Type  string `json:"type"`
		Value any    `json:"value"`
	}
	if err := json.Unmarshal(body, &kvResponsePtr); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("namespace", namespace); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", key); err != nil {
		return diag.FromErr(err)
	}
	valueType := kvResponsePtr.Type
	if err := d.Set("type", valueType); err != nil {
		return diag.FromErr(err)
	}

	value := ""
	if valueType == "JSON" {
		valueBytes, err := json.Marshal(kvResponsePtr.Value)
		if err != nil {
			return diag.FromErr(err)
		}
		value = string(valueBytes)
	} else {
		value = fmt.Sprint(kvResponsePtr.Value)
	}
	if err := d.Set("value", value); err != nil {
		return diag.FromErr(err)
	}

	return diags
}
