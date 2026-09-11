package provider_v2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the desired interfaces.
var _ datasource.DataSource = &reusableInputsDataSource{}

func NewReusableInputsDataSource() datasource.DataSource {
	return &reusableInputsDataSource{}
}

type reusableInputsDataSource struct {
	providerData ProviderData
}

func (d *reusableInputsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reusable_inputs"
}

func (d *reusableInputsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Kestra Reusable Inputs block (EE). Requires Kestra EE 2.0.1 or later.",
		Attributes: map[string]schema.Attribute{
			"reusable_inputs_id": schema.StringAttribute{
				MarkdownDescription: "The reusable inputs block id.",
				Required:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The tenant id (EE). Defaults to the provider tenant when omitted.",
				Optional:            true,
				Computed:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace to look the block up from. A block defined in a parent namespace is visible to child namespaces and the closest definition is returned, so this keeps the namespace you requested rather than the one defining the returned block.",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The block YAML source.",
				Computed:            true,
			},
			"revision": schema.Int64Attribute{
				MarkdownDescription: "The block revision, bumped by the API on every save.",
				Computed:            true,
			},
		},
	}
}

func (d *reusableInputsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *reusableInputsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data reusableInputsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := d.providerData.KestraClient.ReusableInputs().ReusableInputs(ctx, data.Namespace.ValueString(), data.ReusableInputsId.ValueString(), resolveReusableInputsTenantId(d.providerData, data), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read reusable inputs data source, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a reusable_inputs data source, res: %+v", read))

	populateReusableInputsModel(&data, read)
	data.TenantId = types.StringValue(resolveReusableInputsTenantId(d.providerData, data))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
