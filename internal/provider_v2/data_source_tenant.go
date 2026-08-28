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
	_ datasource.DataSource              = &tenantDataSource{}
	_ datasource.DataSourceWithConfigure = &tenantDataSource{}
)

func NewTenantDataSource() datasource.DataSource {
	return &tenantDataSource{}
}

type tenantDataSource struct {
	providerData ProviderData
}

// The nested structures are computed lists of objects rather than blocks, to
// stay representable in protocol v5 which the mux server downgrades to.
type tenantDataSourceModel struct {
	Id                       types.String `tfsdk:"id"`
	TenantId                 types.String `tfsdk:"tenant_id"`
	Name                     types.String `tfsdk:"name"`
	DefaultWorkerSelector    types.List   `tfsdk:"default_worker_selector"`
	StorageType              types.String `tfsdk:"storage_type"`
	StorageConfiguration     types.Map    `tfsdk:"storage_configuration"`
	StorageIsolation         types.List   `tfsdk:"storage_isolation"`
	SecretIsolation          types.List   `tfsdk:"secret_isolation"`
	SecretType               types.String `tfsdk:"secret_type"`
	SecretReadOnly           types.Bool   `tfsdk:"secret_read_only"`
	SecretConfiguration      types.Map    `tfsdk:"secret_configuration"`
	RequireExistingNamespace types.Bool   `tfsdk:"require_existing_namespace"`
	OutputsInInternalStorage types.Bool   `tfsdk:"outputs_in_internal_storage"`
	Concurrency              types.List   `tfsdk:"concurrency"`
	Quotas                   types.List   `tfsdk:"quotas"`
}

func (d *tenantDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (d *tenantDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing Kestra Tenant.\n\n-> This data source is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant id.",
			},
			"tenant_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tenant id.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant name.",
			},
			"storage_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The storage type.",
			},
			"storage_configuration": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The storage configuration.",
			},
			"secret_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The secret type.",
			},
			"secret_read_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether secrets are read-only in this tenant.",
			},
			"secret_configuration": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The secret configuration.",
			},
			"require_existing_namespace": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether tenant requires an existing namespace.",
			},
			"outputs_in_internal_storage": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether outputs are stored in internal storage.",
			},
			"default_worker_selector": schema.ListAttribute{
				Computed:            true,
				ElementType:         workerSelectorObjectType,
				MarkdownDescription: "The default routing applied to every task of the tenant that does not define its own: `tags`, `match` and `fallback`.",
			},
			"storage_isolation": schema.ListAttribute{
				Computed:            true,
				ElementType:         isolationObjectType,
				MarkdownDescription: "Storage isolation configuration: `enabled` and `denied_services`.",
			},
			"secret_isolation": schema.ListAttribute{
				Computed:            true,
				ElementType:         isolationObjectType,
				MarkdownDescription: "Secret isolation configuration: `enabled` and `denied_services`.",
			},
			"concurrency": schema.ListAttribute{
				Computed:            true,
				ElementType:         concurrencyObjectType,
				MarkdownDescription: "The concurrency limit applied to the executions of every flow of the tenant: `limit` and `behavior`.",
			},
			"quotas": schema.ListAttribute{
				Computed:            true,
				ElementType:         quotaObjectType,
				MarkdownDescription: "The quotas evaluated before an execution starts: `duration`, `limit` and `behavior`.",
			},
		},
	}
}

func (d *tenantDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tenantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data tenantDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantId := data.TenantId.ValueString()
	out, _, err := sdk_client.RawRequest(ctx, d.providerData.Client, http.MethodGet, "/api/v1/tenants/"+tenantId, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read tenant failed", err.Error())
		return
	}

	data.Id = types.StringValue(tenantId)
	if id, ok := out["id"].(string); ok {
		data.Id = types.StringValue(id)
		data.TenantId = types.StringValue(id)
	}
	data.Name = optString(out["name"])
	data.StorageType = optString(out["storageType"])
	data.SecretType = optString(out["secretType"])
	data.SecretReadOnly = optBool(out["secretReadOnly"])
	data.RequireExistingNamespace = optBool(out["requireExistingNamespace"])
	data.OutputsInInternalStorage = optBool(out["outputsInInternalStorage"])
	data.StorageConfiguration = stringMapFromBody(out["storageConfiguration"])
	data.SecretConfiguration = stringMapFromBody(out["secretConfiguration"])
	data.DefaultWorkerSelector = workerSelectorToList(out)
	data.StorageIsolation = isolationToList(out, "storageIsolation")
	data.SecretIsolation = isolationToList(out, "secretIsolation")
	data.Concurrency = concurrencyToList(out)
	data.Quotas = quotasToList(out)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
