package provider_v2

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kestraOldProvider "github.com/kestra-io/terraform-provider-kestra/internal/provider"
	"log"
	"os"
	"strconv"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &kestraProvider{}
)

type kestraProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// exampleProviderModel maps provider schema data to a Go type.
type kestraProviderModel struct {
	Url                types.String `tfsdk:"url"`
	TenantId           types.String `tfsdk:"tenant_id"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	Jwt                types.String `tfsdk:"jwt"`
	ApiToken           types.String `tfsdk:"api_token"`
	ExtraHeaders       types.Map    `tfsdk:"extra_headers"`
	KeepOriginalSource types.Bool   `tfsdk:"keep_original_source"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &kestraProvider{
			version: version,
		}
	}
}

func (p *kestraProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "hashicups"
	resp.Version = p.version
}

func (p *kestraProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": &schema.StringAttribute{
				Optional:    true,
				Description: "The endpoint url",
			},
			"tenant_id": &schema.StringAttribute{
				Optional:    true,
				Description: "The tenant id (EE)",
			},
			"username": &schema.StringAttribute{
				Optional:    true,
				Description: "The BasicAuth username",
			},
			"password": &schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The BasicAuth password",
			},
			"timeout": &schema.Int64Attribute{
				Optional:    true,
				Sensitive:   false,
				Description: "The timeout (in seconds) for http requests",
			},
			"jwt": &schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The JWT token (EE)",
			},
			"api_token": &schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The API token (EE)",
			},
			"extra_headers": &schema.MapAttribute{
				Optional:    true,
				Description: "Extra headers to add to every request",
			},
			"keep_original_source": &schema.BoolAttribute{
				Optional:    true,
				Description: "Keep original source code, keeping comment and indentation. Setting to false is now deprecated and will be removed in the future.",
			},
		},
	}
}

func (p *kestraProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config kestraProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// keeping the same conf as sdk2 provider, it will be improved when migration is done
	url := ""
	if !config.Url.IsNull() {
		url = config.Url.ValueString()
	}

	tenantId := "main"
	tenantIdEnv, isTenantIdEnvPresent := os.LookupEnv("KESTRA_TENANT_ID")
	if isTenantIdEnvPresent {
		tenantId = tenantIdEnv
	}
	if !config.TenantId.IsNull() {
		tenantId = config.TenantId.ValueString()
	}

	username := ""
	usernameEnv, isUsernameEnvPresent := os.LookupEnv("KESTRA_USERNAME")
	if isUsernameEnvPresent {
		username = usernameEnv
	}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	password := ""
	passwordEnv, ispasswordEnvPresent := os.LookupEnv("KESTRA_PASSWORD")
	if ispasswordEnvPresent {
		password = passwordEnv
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	var timeout int64 = 10
	timeoutEnv, istimeoutEnvPresent := os.LookupEnv("KESTRA_TIMEOUT")
	if istimeoutEnvPresent {
		i, err := strconv.ParseInt(timeoutEnv, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse KESTRA_TIMEOUT env var",
				"It should be a string, but was: "+timeoutEnv+", err: "+err.Error(),
			)
			return
		}
		timeout = i
	}
	if !config.Timeout.IsNull() {
		timeout = config.Timeout.ValueInt64()
	}

	jwt := ""
	jwtEnv, isjwtEnvPresent := os.LookupEnv("KESTRA_JWT")
	if isjwtEnvPresent {
		jwt = jwtEnv
	}
	if !config.Jwt.IsNull() {
		jwt = config.Jwt.ValueString()
	}

	apiToken := ""
	apiTokenEnv, isapiTokenEnvPresent := os.LookupEnv("KESTRA_API_TOKEN")
	if isapiTokenEnvPresent {
		apiToken = apiTokenEnv
	}
	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	extraHeaders := types.MapNull(config.ExtraHeaders.Type(ctx))
	if !config.ExtraHeaders.IsNull() {
		value, _ := config.ExtraHeaders.ToMapValue(ctx)
		extraHeaders = value
	}

	keepOriginalSource := true
	keepOriginalSourceEnv, iskeepOriginalSourceEnvPresent := os.LookupEnv("KESTRA_KEEP_ORIGINAL_SOURCE")
	if iskeepOriginalSourceEnvPresent {
		i, err := strconv.ParseBool(keepOriginalSourceEnv)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse KESTRA_KEEP_ORIGINAL_SOURCE env var",
				"It should be a string, but was: "+keepOriginalSourceEnv+", err: "+err.Error(),
			)
			return
		}
		keepOriginalSource = i
	}
	if !config.KeepOriginalSource.IsNull() {
		keepOriginalSource = config.KeepOriginalSource.ValueBool()
	}

	if extraHeaders.Type(ctx) != nil {
		log.Printf("rezr")
	}
	client, err := kestraOldProvider.NewClient(url, int64(timeout), &username, &password, &jwt, &apiToken, new(interface{}), &tenantId, &keepOriginalSource)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Provider API Client",
			"An unexpected error occurred when creating the Provider API client. "+
				"Kestra Client Error: "+err.Error(),
		)
		return
	}

	//client := &http.Client{}
	/*if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HashiCups API Client",
			"An unexpected error occurred when creating the HashiCups API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}*/

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *kestraProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *kestraProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}
