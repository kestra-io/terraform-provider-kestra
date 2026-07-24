package provider_v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ datasource.DataSource              = &workerGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &workerGroupDataSource{}
)

func NewWorkerGroupDataSource() datasource.DataSource {
	return &workerGroupDataSource{}
}

type workerGroupDataSource struct {
	providerData ProviderData
}

func (d *workerGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_group"
}

func (d *workerGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing Kestra Worker Group.\n\n-> This data source is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The worker group id.",
			},
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The worker group identifier.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The worker group display name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The worker group description.",
			},
			// A plain list-of-objects attribute (not a nested attribute) so the
			// schema stays representable in protocol v5, which the mux server
			// downgrades the framework provider to.
			"subscriptions": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "The Worker Queue subscriptions of the worker group: `worker_queue_id` (`default` for the global default queue), `reserved_percent` (`-1` means no reservation), and `mode` (`STRICT` or `ELASTIC`).",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"worker_queue_id":  types.StringType,
						"reserved_percent": types.Int64Type,
						"mode":             types.StringType,
					},
				},
			},
		},
	}
}

func (d *workerGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *workerGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data workerGroupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Force the subscriptions list to be populated even when empty; unlike the
	// resource there is no configured value to preserve.
	data.Subscriptions = []subscriptionModel{}

	out, _, err := sdk_client.RawRequest(ctx, d.providerData.Client, http.MethodGet, workerGroupPath(data.GroupId.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Read worker group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerGroupModel(out, &data)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
