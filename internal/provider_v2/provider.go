package provider_v2

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	kestra_api_client "github.com/kestra-io/client-sdk/go-sdk"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
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
	resp.TypeName = "kestra"
	resp.Version = p.version
}

func (p *kestraProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": &schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The endpoint url",
			},
			"tenant_id": &schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The tenant id (EE)",
			},
			"username": &schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The BasicAuth username",
			},
			"password": &schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The BasicAuth password",
			},
			"timeout": &schema.Int64Attribute{
				Optional:            true,
				Sensitive:           false,
				MarkdownDescription: "The timeout (in seconds) for http requests",
			},
			"jwt": &schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The JWT token (EE)",
			},
			"api_token": &schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "The API token (EE)",
			},
			"extra_headers": &schema.MapAttribute{
				Optional:            true,
				MarkdownDescription: "Extra headers to add to every request",
				ElementType:         types.StringType,
			},
			"keep_original_source": &schema.BoolAttribute{
				Optional: true,
				//DeprecationMessage:  "this is not used in new provider version", cannot add this depreciation because bot provider must exactly match
				MarkdownDescription: "Keep original source code, keeping comment and indentation. Setting to false is now deprecated and will be removed in the future.",
			},
		},
	}
}

type ProviderData struct {
	Client   *kestra_api_client.APIClient
	TenantId string
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
	urlEnv, urlEnvPresent := os.LookupEnv("KESTRA_URL")
	if urlEnvPresent && urlEnv != "" {
		url = urlEnv
	}
	if !config.Url.IsNull() && config.Url.ValueString() != "" {
		url = config.Url.ValueString()
	}

	tenantId := "main"
	tenantIdEnv, isTenantIdEnvPresent := os.LookupEnv("KESTRA_TENANT_ID")
	if isTenantIdEnvPresent && tenantIdEnv != "" {
		tenantId = tenantIdEnv
	}
	if !config.TenantId.IsNull() && config.TenantId.ValueString() != "" {
		tenantId = config.TenantId.ValueString()
	}

	var username *string = nil
	usernameEnv, isUsernameEnvPresent := os.LookupEnv("KESTRA_USERNAME")
	if isUsernameEnvPresent {
		username = &usernameEnv
	}
	if !config.Username.IsNull() && config.Username.ValueString() != "" {
		tmp := config.Username.ValueString()
		username = &tmp
	}

	var password *string = nil
	passwordEnv, ispasswordEnvPresent := os.LookupEnv("KESTRA_PASSWORD")
	if ispasswordEnvPresent {
		password = &passwordEnv
	}
	if !config.Password.IsNull() && config.Password.ValueString() != "" {
		tmp := config.Password.ValueString()
		password = &tmp
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

	var jwt *string = nil
	jwtEnv, isjwtEnvPresent := os.LookupEnv("KESTRA_JWT")
	if isjwtEnvPresent {
		jwt = &jwtEnv
	}
	if !config.Jwt.IsNull() && config.Jwt.ValueString() != "" {
		tmp := config.Jwt.ValueString()
		jwt = &tmp
	}

	var apiToken *string = nil
	apiTokenEnv, isapiTokenEnvPresent := os.LookupEnv("KESTRA_API_TOKEN")
	if isapiTokenEnvPresent {
		apiToken = &apiTokenEnv
	}
	if !config.ApiToken.IsNull() && config.ApiToken.ValueString() != "" {
		tmp := config.ApiToken.ValueString()
		apiToken = &tmp
	}

	extraHeaders := make(map[string]string)
	if !config.ExtraHeaders.IsNull() {
		valueMap, _ := config.ExtraHeaders.ToMapValue(ctx)
		for k, v := range valueMap.Elements() {
			extraHeaders[k] = v.String()
		}
	}

	// validation
	if url == "" {
		resp.Diagnostics.AddError(
			"url is mandatory",
			"It was null or empty",
		)
	}
	if tenantId == "" {
		resp.Diagnostics.AddError(
			"tenantId is mandatory",
			"It was null or empty",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	_, iskeepOriginalSourceEnvPresent := os.LookupEnv("KESTRA_KEEP_ORIGINAL_SOURCE")
	if iskeepOriginalSourceEnvPresent || !config.KeepOriginalSource.IsNull() {
		resp.Diagnostics.AddWarning(
			"keep_original_source is not used anymore in our new provider", "",
		)
	}

	client, err := sdk_client.NewClient(ctx, url, int64(timeout), username, password, jwt, apiToken, &extraHeaders)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Kestra API Client",
			"An unexpected error occurred when creating the API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Kestra API Clien: "+err.Error(),
		)
		return
	}

	// Make client and tenantId available during DataSource and Resource type Configure methods.
	providerData := &ProviderData{
		Client:   client,
		TenantId: tenantId,
	}
	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *kestraProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTestResource,
	}
}

func (p *kestraProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTestDataSource,
	}
}
