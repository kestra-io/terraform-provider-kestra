package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceUserTenantAccess() *schema.Resource {
	return &schema.Resource{
		Description: "Grants a user access to a tenant.\n\n" +
			"Since Kestra 2.0 tenant access is a prerequisite rather than something granted " +
			"as a side effect: a tenant-scoped write such as `kestra_user_group_membership` " +
			"resolves the user through a tenant access check and rejects it otherwise. " +
			"Declare this resource for every user that a tenant-scoped resource refers to, " +
			"and make those resources depend on it.\n\n" +
			"Destroying this resource revokes the user's access to the tenant. Removing a " +
			"`kestra_user_group_membership` does not, since other memberships and tokens may " +
			"still rely on the access." +
			EnterpriseEditionDescription,

		CreateContext: resourceUserTenantAccessCreate,
		ReadContext:   resourceUserTenantAccessRead,
		DeleteContext: resourceUserTenantAccessDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"user_id": {
				Description: "The id of the user granted access to the tenant.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserTenantAccessImport,
		},
	}
}

func userTenantAccessUrl(tenantId *string, userId string) string {
	return fmt.Sprintf("%s/tenant-access/%s", apiRoot(tenantId), userId)
}

func resourceUserTenantAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)

	userId := d.Get("user_id").(string)

	// The grant is idempotent: the API answers 409 when the user already has
	// access to the tenant, which leaves us in the desired state either way.
	_, reqErr := c.request("PUT", userTenantAccessUrl(c.TenantId, userId), nil)
	if reqErr != nil && reqErr.StatusCode != http.StatusConflict {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(userId)

	return resourceUserTenantAccessRead(ctx, d, meta)
}

func resourceUserTenantAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	userId := d.Id()

	_, reqErr := c.request("GET", userTenantAccessUrl(c.TenantId, userId), nil)
	if reqErr != nil {
		if reqErr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return diag.FromErr(reqErr.Err)
	}

	if err := d.Set("user_id", userId); err != nil {
		return diag.FromErr(err)
	}
	if c.TenantId != nil && *c.TenantId != "" {
		if err := d.Set("tenant_id", *c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceUserTenantAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	_, reqErr := c.request("DELETE", userTenantAccessUrl(c.TenantId, d.Id()), nil)
	if reqErr != nil && reqErr.StatusCode != http.StatusNotFound {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")
	return diags
}

func resourceUserTenantAccessImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	userId := d.Id()
	if strings.Contains(userId, "/") {
		return nil, fmt.Errorf(`import id must be "<user_id>", got: %q`, userId)
	}
	if userId == "" {
		return nil, fmt.Errorf("import id must be a non-empty user id")
	}
	if err := d.Set("user_id", userId); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
