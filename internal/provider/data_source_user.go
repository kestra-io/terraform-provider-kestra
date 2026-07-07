package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to access information about an existing Kestra User." +
			EnterpriseEditionDescription,

		ReadContext: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"user_id": {
				Description:  "The user id. Exactly one of `user_id` or `email` must be provided.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"user_id", "email"},
			},
			"namespace": {
				Description: "The linked namespace.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"username": {
				Description: "The user name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "The user description.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"first_name": {
				Description: "The user first name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_name": {
				Description: "The user last name.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description:  "The user email. Can be set instead of `user_id` to look up the user.",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"user_id", "email"},
			},
			"groups": {
				Description: "The user global roles in yaml string.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	userId := d.Get("user_id").(string)

	if userId == "" {
		id, err := findUserIdByEmail(c, d.Get("email").(string))
		if err != nil {
			return diag.FromErr(err)
		}
		userId = id

		if err := d.Set("user_id", userId); err != nil {
			return diag.FromErr(err)
		}
	}

	r, reqErr := c.request("GET", fmt.Sprintf("%s/users/%s", apiRoot(nil), userId), nil)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	errs := userApiToSchema(r.(map[string]interface{}), d)
	if errs != nil {
		return errs
	}

	return diags
}

func findUserIdByEmail(c *Client, email string) (string, error) {
	query := url.Values{
		"filters[q][EQUALS]": {email},
		"size":               {"100"},
	}

	r, reqErr := c.request("GET", fmt.Sprintf("%s/users?%s", apiRoot(nil), query.Encode()), nil)
	if reqErr != nil {
		return "", reqErr.Err
	}

	body, _ := r.(map[string]interface{})
	results, ok := body["results"].([]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response searching users by email '%s'", email)
	}

	var matches []string
	for _, item := range results {
		user, _ := item.(map[string]interface{})
		username, _ := user["username"].(string)
		id, _ := user["id"].(string)

		// usernames are emails, the search endpoint only exposes username
		if id != "" && strings.EqualFold(username, email) {
			matches = append(matches, id)
		}
	}

	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no user found with email '%s'", email)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("multiple users found with email '%s'", email)
	}
}
