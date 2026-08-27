package provider_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &policyDataSource{}
var _ datasource.DataSourceWithValidateConfig = &policyDataSource{}

func NewPolicyDataSource() datasource.DataSource {
	return &policyDataSource{}
}

type policyDataSource struct {
	providerData ProviderData
}

func (d *policyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *policyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kestra governance Policy (EE) at the `INSTANCE`, `TENANT` or `NAMESPACE` scope.",

		Attributes: map[string]schema.Attribute{
			"scope": schema.StringAttribute{
				MarkdownDescription: "The policy scope: `INSTANCE`, `TENANT` or `NAMESPACE`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(policyScopeInstance, policyScopeTenant, policyScopeNamespace),
				},
			},
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "The policy id.",
				Required:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The tenant id, for `TENANT` and `NAMESPACE` scopes. Defaults to the provider tenant when omitted. Must not be set for the `INSTANCE` scope.",
				Optional:            true,
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace the policy is attached to. Required for the `NAMESPACE` scope.",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The policy YAML source.",
				Computed:            true,
			},
		},
	}
}

func (d *policyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data source Configure Type",
			fmt.Sprintf("Expected ProviderData type, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerData = *providerData
}

// ValidateConfig mirrors the resource's scope-binding rules so a namespace or tenant that
// would silently be ignored when building the request URL is rejected at config time.
func (d *policyDataSource) ValidateConfig(ctx context.Context, req datasource.ValidateConfigRequest, resp *datasource.ValidateConfigResponse) {
	var data policyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Scope.IsNull() || data.Scope.IsUnknown() {
		return
	}

	switch data.Scope.ValueString() {
	case policyScopeInstance:
		if !data.TenantId.IsNull() && !data.TenantId.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("tenant_id"), "Invalid scope binding", "tenant_id must not be set for an INSTANCE scope policy.")
		}
		if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("namespace"), "Invalid scope binding", "namespace must not be set for an INSTANCE scope policy.")
		}
	case policyScopeTenant:
		if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() {
			resp.Diagnostics.AddAttributeError(path.Root("namespace"), "Invalid scope binding", "namespace must not be set for a TENANT scope policy.")
		}
	case policyScopeNamespace:
		if data.Namespace.IsNull() {
			resp.Diagnostics.AddAttributeError(path.Root("namespace"), "Invalid scope binding", "namespace is required for a NAMESPACE scope policy.")
		}
	}
}

func (d *policyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data policyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, _, err := readPolicy(ctx, d.providerData.Client, data.Scope.ValueString(), resolvePolicyTenantId(d.providerData, data), data.Namespace.ValueString(), data.PolicyId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read policy data source, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a policy data source, res: %+v", read))

	// force the content to be rendered from the API payload
	data.Content = types.StringNull()
	populatePolicyModel(ctx, &data, read, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if data.Scope.ValueString() != policyScopeInstance {
		data.TenantId = types.StringValue(resolvePolicyTenantId(d.providerData, data))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
