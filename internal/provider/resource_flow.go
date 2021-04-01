package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
)

func resourceFlow() *schema.Resource {
	return &schema.Resource{
		Description: "Sample resource in the Terraform provider kestra.",

		CreateContext: resourceFlowCreate,
		ReadContext:   resourceFlowRead,
		UpdateContext: resourceFlowUpdate,
		DeleteContext: resourceFlowDelete,

		Schema: map[string]*schema.Schema{
			"namespace": {
				Description: "The flow namespace.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"flow_id": {
				Description: "A unique ID for this flow.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"content": {
				Description: "The flow full content in yaml string.",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := &http.Client{Timeout: 10 * time.Second}

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	flowId := d.Get("flow_id")
	namespace := d.Get("namespace")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/flows/%s/%s", "http://kestra:8080", namespace, flowId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	flowApi := make(map[string]interface{}, 0)
	err = json.NewDecoder(r.Body).Decode(&flowApi)
	if err != nil {
		return diag.FromErr(err)
	}

	delete(flowApi, "deleted")
	delete(flowApi, "id")
	delete(flowApi, "namespace")
	delete(flowApi, "revision")

	flow, err := yaml.Marshal(&flowApi)
	if err != nil {
		return diag.FromErr(err)
	}

	fmt.Printf("--- t dump:\n%s\n\n", string(flow))

	if err := d.Set("flowApi", flowApi); err != nil {
		return diag.FromErr(err)
	}

	// always run
	d.SetId(string(flowApi["namespace"]) + "." + string(flowApi["id"]))

	return diags

	idFromAPI := "my-id"
	d.SetId(idFromAPI)

	return diag.Errorf("not implemented")
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}
