package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// userMembershipLocks serialises membership writes targeting the same user.
// Kestra's API silently loses a write when two memberships for the same user
// are PUT in parallel: the second request returns 200 but the member index
// stays at its prior state. Terraform happily parallelises sibling resources,
// so without this lock a config that adds the same user to multiple groups
// flakes intermittently. The lock is keyed by user id, so different users
// remain independent and overall apply throughput is unaffected.
var (
	userMembershipLocksMu sync.Mutex
	userMembershipLocks   = map[string]*sync.Mutex{}
)

func lockUserMembership(userId string) func() {
	userMembershipLocksMu.Lock()
	lock, ok := userMembershipLocks[userId]
	if !ok {
		lock = &sync.Mutex{}
		userMembershipLocks[userId] = lock
	}
	userMembershipLocksMu.Unlock()

	lock.Lock()
	return lock.Unlock
}

func resourceUserGroupMembership() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a single membership of a user in a group. " +
			"Use this resource when different Terraform configurations need to manage " +
			"different group memberships for the same user, without overwriting each " +
			"other's assignments. Each resource owns exactly one user-group pair." +
			EnterpriseEditionDescription,

		CreateContext: resourceUserGroupMembershipCreate,
		ReadContext:   resourceUserGroupMembershipRead,
		DeleteContext: resourceUserGroupMembershipDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"user_id": {
				Description: "The id of the user.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group_id": {
				Description: "The id of the group the user should belong to.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceUserGroupMembershipImport,
		},
	}
}

func userGroupMembershipUrl(tenantId *string, groupId, userId string) string {
	return fmt.Sprintf("%s/groups/%s/members/%s", apiRoot(tenantId), groupId, userId)
}

func userGroupMembershipId(groupId, userId string) string {
	return fmt.Sprintf("%s/%s", groupId, userId)
}

func resourceUserGroupMembershipCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	groupId := d.Get("group_id").(string)
	userId := d.Get("user_id").(string)

	unlock := lockUserMembership(userId)
	defer unlock()

	_, reqErr := c.request("PUT", userGroupMembershipUrl(c.TenantId, groupId, userId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId(userGroupMembershipId(groupId, userId))
	if c.TenantId != nil && *c.TenantId != "" {
		if err := d.Set("tenant_id", *c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceUserGroupMembershipRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	userId := d.Get("user_id").(string)
	groupId := d.Get("group_id").(string)

	// Kestra's group-member listing is backed by an eventually-consistent
	// index. A freshly added member can be absent for several seconds. Retry
	// with exponential backoff before concluding the membership is gone.
	var found bool
	var err error
	backoffs := []time.Duration{0, 1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second}
	for _, wait := range backoffs {
		if wait > 0 {
			time.Sleep(wait)
		}
		found, err = userGroupMembershipExists(c, groupId, userId)
		if err != nil {
			return diag.FromErr(err)
		}
		if found {
			break
		}
	}
	if !found {
		d.SetId("")
		return diags
	}

	if c.TenantId != nil && *c.TenantId != "" {
		if err := d.Set("tenant_id", *c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}
	return diags
}

// userGroupMembershipExists checks whether userId is currently a member of
// groupId. It walks the paged /groups/{id}/members endpoint, which only
// requires read permission on the group (not on the user), so it works for
// callers that can manage memberships but not view user records directly.
func userGroupMembershipExists(c *Client, groupId, userId string) (bool, error) {
	page := 1
	pageSize := 200
	for {
		url := fmt.Sprintf("%s/groups/%s/members?page=%d&size=%d", apiRoot(c.TenantId), groupId, page, pageSize)
		r, reqErr := c.request("GET", url, nil)
		if reqErr != nil {
			if reqErr.StatusCode == http.StatusNotFound {
				return false, nil
			}
			return false, reqErr.Err
		}
		resp, ok := r.(map[string]interface{})
		if !ok {
			return false, fmt.Errorf("unexpected response shape listing members of group %s", groupId)
		}
		results, _ := resp["results"].([]interface{})
		for _, item := range results {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if id, _ := m["id"].(string); id == userId {
				return true, nil
			}
		}
		if len(results) < pageSize {
			return false, nil
		}
		page++
	}
}

func resourceUserGroupMembershipDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	groupId := d.Get("group_id").(string)
	userId := d.Get("user_id").(string)

	unlock := lockUserMembership(userId)
	defer unlock()

	_, reqErr := c.request("DELETE", userGroupMembershipUrl(c.TenantId, groupId, userId), nil)
	if reqErr != nil && reqErr.StatusCode != http.StatusNotFound {
		return diag.FromErr(reqErr.Err)
	}

	d.SetId("")
	return diags
}

func resourceUserGroupMembershipImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf(`import id must be "<group_id>/<user_id>", got: %q`, d.Id())
	}
	if err := d.Set("group_id", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("user_id", parts[1]); err != nil {
		return nil, err
	}
	d.SetId(userGroupMembershipId(parts[0], parts[1]))
	return []*schema.ResourceData{d}, nil
}
