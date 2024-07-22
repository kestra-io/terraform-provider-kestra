package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output.
	schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
		desc := s.Description
		if s.Default != nil {
			desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
		}
		return strings.TrimSpace(desc)
	}
}

func New(version string, tenant *string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"url": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The endpoint url",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_URL", ""),
				},
				"tenant_id": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The tenant id (EE)",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_TENANT_ID", ""),
				},
				"username": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The BasicAuth username",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_USERNAME", ""),
				},
				"password": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The BasicAuth password",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_PASSWORD", ""),
				},
				"jwt": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The JWT token (EE)",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_JWT", ""),
				},
				"api_token": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The API token (EE)",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_API_TOKEN", ""),
				},
				"extra_headers": &schema.Schema{
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Extra headers to add to every request",
				},
				"keep_original_source": &schema.Schema{
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Keep original source code, keeping comment and indentation. Setting to false is now deprecated and will be removed in the future.",
					DefaultFunc: schema.EnvDefaultFunc("KESTRA_KEEP_ORIGINAL_SOURCE", true),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"kestra_binding":         dataSourceBinding(),
				"kestra_flow":            dataSourceFlow(),
				"kestra_group":           dataSourceGroup(),
				"kestra_namespace":       dataSourceNamespace(),
				"kestra_role":            dataSourceRole(),
				"kestra_template":        dataSourceTemplate(),
				"kestra_user":            dataSourceUser(),
				"kestra_tenant":          dataSourceTenant(),
				"kestra_namespace_file":  dataSourceNamespaceFile(),
				"kestra_service_account": dataSourceServiceAccount(),
				"kestra_user_api_tokens": dataSourceUserApiTokens(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"kestra_binding":          resourceBinding(),
				"kestra_flow":             resourceFlow(),
				"kestra_group":            resourceGroup(),
				"kestra_namespace":        resourceNamespace(),
				"kestra_namespace_secret": resourceNamespaceSecret(),
				"kestra_role":             resourceRole(),
				"kestra_template":         resourceTemplate(),
				"kestra_user":             resourceUser(),
				"kestra_user_password":    resourceUserPassword(),
				"kestra_tenant":           resourceTenant(),
				"kestra_namespace_file":   resourceNamespaceFile(),
				"kestra_service_account":  resourceServiceAccount(),
				"kestra_user_api_token":   resourceUserApiToken(),
			},
		}

		p.ConfigureContextFunc = func(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			url := strings.TrimRight(d.Get("url").(string), "/")
			username := d.Get("username").(string)
			password := d.Get("password").(string)
			jwt := d.Get("jwt").(string)
			apiToken := d.Get("api_token").(string)
			extraHeaders := d.Get("extra_headers")
			keepOriginalSource := d.Get("keep_original_source").(bool)

			tenantId := ""
			if tenant != nil {
				tenantId = *tenant
			} else if d.Get("tenant_id") != nil {
				tenantId = d.Get("tenant_id").(string)
			}

			var diags diag.Diagnostics

			c, err := NewClient(url, &username, &password, &jwt, &apiToken, &extraHeaders, &tenantId, &keepOriginalSource)
			if err != nil {
				return nil, diag.FromErr(err)
			}

			return c, diags
		}

		return p
	}
}

func stateFn(i interface{}) string {
	var asInterface interface{}
	_ = yaml.Unmarshal([]byte(i.(string)), &asInterface)

	newYaml, _ := yaml.Marshal(asInterface)

	return cleanUpYaml(newYaml)
}

func toYaml(source interface{}) (*string, error) {
	out, err := yaml.Marshal(source)
	if err != nil {
		return nil, err
	}

	yamlString := string(out)

	return &yamlString, nil
}

func isYamlEqualsFlow(k, old, new string, d *schema.ResourceData) bool {
	if _, ok := d.Get("keep_original_source").(bool); ok {
		if d.Get("keep_original_source").(bool) == true {
			// seems that new is the json one and not the yaml one, so using the state directly
			return old == d.Get("content").(string)
		}
	}

	oldInterface := make(map[string]interface{}, 0)
	_ = yaml.Unmarshal([]byte(old), &oldInterface)

	newInterface := make(map[string]interface{}, 0)
	_ = yaml.Unmarshal([]byte(new), &newInterface)

	delete(oldInterface, "deleted")
	delete(oldInterface, "id")
	delete(oldInterface, "namespace")
	delete(oldInterface, "revision")

	delete(newInterface, "deleted")
	delete(newInterface, "id")
	delete(newInterface, "namespace")
	delete(newInterface, "revision")

	return yamlCompare(oldInterface, newInterface)
}

//goland:noinspection GoUnhandledErrorResult
func isYamlEquals(k, old, new string, d *schema.ResourceData) bool {
	var oldInterface interface{}
	yaml.Unmarshal([]byte(old), &oldInterface)

	var newInterface interface{}
	yaml.Unmarshal([]byte(new), &newInterface)

	return yamlCompare(oldInterface, newInterface)
}

func yamlCompare(oldInterface, newInterface interface{}) bool {
	result := reflect.DeepEqual(oldInterface, newInterface)

	return result
}

func cleanUpYaml(ymlBytes []byte) string {
	ymlStr := string(ymlBytes)
	return strings.ReplaceAll(ymlStr, "\r\n", "\n")
}

func stringToPointer(s string) *string {
	return &s
}

func pointerToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func apiRoot(tenantId *string) string {
	if tenantId == nil || *tenantId == "" {
		return "/api/v1"
	}

	return fmt.Sprintf("/api/v1/%s", *tenantId)
}
