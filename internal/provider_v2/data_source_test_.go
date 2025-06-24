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
var _ datasource.DataSource = &testDataSource{}

func NewTestDataSource() datasource.DataSource {
	return &testDataSource{}
}

type testDataSource struct {
	providerData ProviderData
}

type dataSourceModel struct {
	TestId    types.String `tfsdk:"test_id"`
	Namespace types.String `tfsdk:"namespace"`
	Content   types.String `tfsdk:"content"`
}

func (d *testDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_test"
}

func (d *testDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Test data source",

		Attributes: map[string]schema.Attribute{
			"test_id": schema.StringAttribute{
				MarkdownDescription: "The Test id",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The Test namespace",
				Required:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The actual Test YAML content",
				Computed:            true,
			},
		},
	}
}
func (d *testDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *testDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, httpResponse, err := d.providerData.Client.TestSuitesAPI.GetTestSuite(ctx, data.Namespace.ValueString(), data.TestId.ValueString(), d.providerData.TenantId).Execute()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Test data source, got error: %s, full httpResponse: %v", err, httpResponse))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a test data source, res: %v", read))
	data.Namespace = types.StringValue(read.Namespace)
	data.TestId = types.StringValue(read.Id)
	data.Content = types.StringValue(*read.Source)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
