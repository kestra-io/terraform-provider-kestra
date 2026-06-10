package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Role." +
			EnterpriseEditionDescription,

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Version: 0,
				Type:    resourceRoleV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceRoleStateUpgradeV0,
			},
		},
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "The role name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "The role description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"is_default": {
				Description: "The role is the default one at user creation. Only one role can be default. Latest create/update to true will be keep as default.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"resources": {
				Description: "The role resource permissions.",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "The resource type (e.g., FLOW, EXECUTION, NAMESPACE).",
							Type:        schema.TypeString,
							Required:    true,
						},
						"actions": {
							Description: "The allowed actions for this resource type (e.g., VIEW, LIST, CREATE, UPDATE, DELETE).",
							Type:        schema.TypeList,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := c.TenantId

	body, err := roleSchemaToApi(d)
	if err != nil {
		return diag.FromErr(err)
	}

	r, reqErr := c.request("POST", fmt.Sprintf("%s/roles", apiRoot(tenantId)), body)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := roleApiToSchema(r.(map[string]interface{}), d, c)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	roleId := d.Id()
	tenantId := c.TenantId

	r, reqErr := c.request("GET", fmt.Sprintf("%s/roles/%s", apiRoot(tenantId), roleId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}

		return diag.FromErr(reqErr.Err)
	}

	errs := roleApiToSchema(r.(map[string]interface{}), d, c)
	if errs != nil {
		return errs
	}

	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	if d.HasChanges("namespace", "name", "description", "resources", "is_default") {
		body, err := roleSchemaToApi(d)
		if err != nil {
			return diag.FromErr(err)
		}

		roleId := d.Id()
		tenantId := c.TenantId

		r, reqErr := c.request("PUT", fmt.Sprintf("%s/roles/%s", apiRoot(tenantId), roleId), body)
		if reqErr != nil {
			return diag.FromErr(reqErr.Err)
		}

		errs := roleApiToSchema(r.(map[string]interface{}), d, c)
		if errs != nil {
			return errs
		}

		return diags
	} else {
		return resourceRoleRead(ctx, d, meta)
	}
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	roleId := d.Id()
	tenantId := c.TenantId

	_, reqErr := c.request("DELETE", fmt.Sprintf("%s/roles/%s", apiRoot(tenantId), roleId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")

	return diags
}

func resourceRoleV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permissions": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceRoleStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldPerms, ok := rawState["permissions"]
	if !ok || oldPerms == nil {
		return rawState, nil
	}

	var items []interface{}
	switch v := oldPerms.(type) {
	case *schema.Set:
		items = v.List()
	case []interface{}:
		items = v
	default:
		return rawState, nil
	}

	return upgradePermissions(rawState, items)
}

func upgradePermissions(rawState map[string]interface{}, items []interface{}) (map[string]interface{}, error) {
	oldMap := make(map[string]interface{})
	for _, item := range items {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		resType, _ := entry["type"].(string)
		// V0 state used "permissions" as the inner key
		oldMap[resType] = entry["permissions"]
	}

	if !isAlreadyMigrated(oldMap) {
		oldMap = migratePermissions(oldMap)
	}

	// Write to new "resources" key with "actions" inner key, remove old "permissions" key
	delete(rawState, "permissions")
	rawState["resources"] = resourcesToState(oldMap)
	return rawState, nil
}

func resourcesToState(m map[string]interface{}) []interface{} {
	result := make([]interface{}, 0, len(m))
	for resType, actions := range m {
		result = append(result, map[string]interface{}{
			"type":    resType,
			"actions": actions,
		})
	}
	return result
}
