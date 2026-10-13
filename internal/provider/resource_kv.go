package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

func resourceKv() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a Kestra Key-Value pair.",

		CreateContext: resourceKvSet,
		UpdateContext: resourceKvSet,
		ReadContext:   resourceKvRead,
		DeleteContext: resourceKvDelete,
		Schema: map[string]*schema.Schema{
			"tenant_id": {
				Description: "The tenant id.",
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
			},
			"namespace": {
				Description: "The namespace of the Key-Value pair.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"key": {
				Description: "The key of the pair.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Description: "The type of the value. If not provided, we will try to deduce the type based on the value. Useful in case you provide numbers, booleans, dates or json that you want to be stored as string." +
					" Accepted values are: STRING, NUMBER, BOOLEAN, DATETIME, DATE, DURATION, JSON.",
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return newValue == "" || oldValue == newValue
				},
				DiffSuppressOnRefresh: true,
			},
			"value": {
				Description: "The fetched value.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description:  "An optional description for the Key-Value pair. Printable ASCII only, single line, up to 4096 bytes. Only applied by Kestra versions that support KV descriptions (2.1.0 or later).",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateKvDescription,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// The server decodes the header as ISO-8859-1, so anything but printable ASCII is mangled.
func validateKvDescription(v interface{}, k string) (ws []string, errs []error) {
	s := v.(string)
	for _, r := range s {
		if r < 0x20 || r > 0x7E {
			errs = append(errs, fmt.Errorf("%q must only contain printable ASCII characters on a single line, it is sent as an HTTP header", k))
			break
		}
	}
	if len(s) > 4096 {
		errs = append(errs, fmt.Errorf("%q must be at most 4096 bytes, got %d", k, len(s)))
	}
	return
}

// kvDescription reads the description from the list endpoint, the single key endpoint does not return it.
func kvDescription(ctx context.Context, c *Client, namespace string, key string) (string, error) {
	for page := 1; ; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("size", "100")
		q.Set("filters[namespace][EQUALS]", namespace)
		req, err := http.NewRequestWithContext(ctx, "GET", c.Url+apiRoot(c.TenantId)+"/kv?"+q.Encode(), nil)
		if err != nil {
			return "", err
		}
		_, body, reqErr := c.rawResponseRequest("GET", req)
		if reqErr != nil {
			return "", reqErr.Err
		}
		var list struct {
			Results []struct {
				Key         string `json:"key"`
				Description string `json:"description"`
			} `json:"results"`
			Total int `json:"total"`
		}
		if err := json.Unmarshal(body, &list); err != nil {
			return "", err
		}
		for _, r := range list.Results {
			if r.Key == key {
				return r.Description, nil
			}
		}
		if page*100 >= list.Total {
			return "", nil
		}
	}
}

func resourceKvSet(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := c.TenantId
	namespace := d.Get("namespace").(string)
	key := d.Get("key").(string)
	typeValue, typeWasProvided := d.GetOk("type")
	value := d.Get("value").(string)

	formattedValue := value
	if typeWasProvided {
		if typeValue == "STRING" && !strings.HasPrefix(value, "\"") {
			formattedValue = fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
		}
	} else {
		stringPattern := "^(?![\\d{\\[\"]+)(?!false)(?!true)(?!P(?=[^T]|T.)(?:\\d*D)?(?:T(?=.)(?:\\d*H)?(?:\\d*M)?(?:\\d*S)?)?).*$"
		matched, _ := regexp.MatchString(stringPattern, value)
		if matched {
			formattedValue = fmt.Sprintf("\"%s\"", value)
		}
	}

	url := c.Url + fmt.Sprintf("%s/namespaces/%s/kv/%s", apiRoot(tenantId), namespace, key)

	httpMethod := "PUT"
	req, err := http.NewRequestWithContext(ctx, httpMethod, fmt.Sprintf("%s", url), strings.NewReader(formattedValue))
	if err != nil {
		return diag.FromErr(err)
	}
	// Omitting the header clears the description server-side, so only send it when set.
	if description := d.Get("description").(string); description != "" {
		req.Header.Set("description", description)
	}

	_, _, reqErr := c.rawResponseRequest(httpMethod, req)
	if reqErr != nil {
		return diag.FromErr(reqErr.Err)
	}

	if tenantId != nil {
		if err := d.Set("tenant_id", c.TenantId); err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%s/%s", namespace, key))

	return diags
}

func resourceKvRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := c.TenantId
	namespace, key := idToNamespaceAndKey(d.Id())

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
	description, err := kvDescription(ctx, c, namespace, key)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", description); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceKvDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client)
	var diags diag.Diagnostics

	tenantId := c.TenantId
	namespace, key := idToNamespaceAndKey(d.Id())

	url := c.Url + fmt.Sprintf("%s/namespaces/%s/kv/%s", apiRoot(tenantId), namespace, key)

	httpMethod := "DELETE"
	req, err := http.NewRequestWithContext(ctx, httpMethod, fmt.Sprintf("%s", url), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	statusCode, _, reqErr := c.rawResponseRequest(httpMethod, req)
	if reqErr != nil {
		if statusCode != http.StatusNotFound {
			return diag.FromErr(reqErr.Err)
		}
	}

	d.SetId("")

	return diags
}

func idToNamespaceAndKey(id string) (string, string) {
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		return "", ""
	}

	return parts[0], parts[1]
}
