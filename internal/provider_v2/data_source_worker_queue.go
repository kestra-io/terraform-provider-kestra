package provider_v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ datasource.DataSource              = &workerQueueDataSource{}
	_ datasource.DataSourceWithConfigure = &workerQueueDataSource{}
)

func NewWorkerQueueDataSource() datasource.DataSource {
	return &workerQueueDataSource{}
}

type workerQueueDataSource struct {
	providerData ProviderData
}

func (d *workerQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_queue"
}

func (d *workerQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing Kestra Worker Queue.\n\n-> This data source is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Worker Queue id.",
			},
			"queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Worker Queue identifier.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Worker Queue human-readable name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Worker Queue description.",
			},
			"tags": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The canonical tag set of the Worker Queue.",
			},
			"allowed_tenants": schema.SetAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tenants allowed to use the Worker Queue. Empty means unrestricted.",
			},
		},
	}
}

func (d *workerQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	d.providerData = *pd
}

func (d *workerQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workerQueueModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, d.providerData.Client, http.MethodGet, workerQueuePath(data.QueueId.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Worker Queue failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerQueueModel(ctx, out, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
