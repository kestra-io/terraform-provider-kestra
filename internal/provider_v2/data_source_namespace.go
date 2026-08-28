package provider_v2

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v2"

	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ datasource.DataSource              = &namespaceDataSource{}
	_ datasource.DataSourceWithConfigure = &namespaceDataSource{}
)

func NewNamespaceDataSource() datasource.DataSource {
	return &namespaceDataSource{}
}

type namespaceDataSource struct {
	providerData ProviderData
}

var allowedNamespaceObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"namespace": types.StringType,
}}

// The nested structures are computed lists of objects rather than blocks, to
// stay representable in protocol v5 which the mux server downgrades to.
type namespaceDataSourceModel struct {
	Id                       types.String  `tfsdk:"id"`
	TenantId                 types.String  `tfsdk:"tenant_id"`
	NamespaceId              types.String  `tfsdk:"namespace_id"`
	Description              types.String  `tfsdk:"description"`
	Variables                types.String  `tfsdk:"variables"`
	AllowedNamespaces        types.List    `tfsdk:"allowed_namespaces"`
	DefaultWorkerSelector    types.List    `tfsdk:"default_worker_selector"`
	StorageType              types.String  `tfsdk:"storage_type"`
	StorageConfiguration     types.Map     `tfsdk:"storage_configuration"`
	StorageIsolation         types.List    `tfsdk:"storage_isolation"`
	SecretIsolation          types.List    `tfsdk:"secret_isolation"`
	SecretType               types.String  `tfsdk:"secret_type"`
	SecretReadOnly           types.Bool    `tfsdk:"secret_read_only"`
	SecretConfiguration      types.Dynamic `tfsdk:"secret_configuration"`
	OutputsInInternalStorage types.Bool    `tfsdk:"outputs_in_internal_storage"`
}

func (d *namespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (d *namespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to access information about an existing Kestra Namespace.\n\n-> Some attributes are only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace id.",
			},
			"tenant_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant id.",
			},
			"namespace_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The namespace id.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace description.",
			},
			"variables": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace variables, as YAML.",
			},
			"storage_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace storage type.",
			},
			"storage_configuration": schema.MapAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The namespace storage configuration.",
			},
			"secret_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace secret type.",
			},
			"secret_read_only": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the namespace secret manager is read only.",
			},
			"secret_configuration": schema.DynamicAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace secret configuration.",
			},
			"outputs_in_internal_storage": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the task outputs are stored in the internal storage.",
			},
			"allowed_namespaces": schema.ListAttribute{
				Computed:            true,
				ElementType:         allowedNamespaceObjectType,
				MarkdownDescription: "The allowed namespaces.",
			},
			"default_worker_selector": schema.ListAttribute{
				Computed:            true,
				ElementType:         workerSelectorObjectType,
				MarkdownDescription: "The default routing applied to every task of the namespace that does not define its own: `tags`, `match` and `fallback`.",
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
		},
	}
}

func (d *namespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *namespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data namespaceDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/v1/%s/namespaces/%s", d.providerData.TenantId, data.NamespaceId.ValueString())
	out, _, err := sdk_client.RawRequest(ctx, d.providerData.Client, http.MethodGet, path, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read namespace failed", err.Error())
		return
	}

	data.TenantId = types.StringValue(d.providerData.TenantId)
	if id, ok := out["id"].(string); ok {
		data.Id = types.StringValue(id)
		data.NamespaceId = types.StringValue(id)
	}
	data.Description = optString(out["description"])
	data.Variables = yamlFromBody(out["variables"])
	data.StorageType = optString(out["storageType"])
	data.SecretType = optString(out["secretType"])
	data.SecretReadOnly = optBool(out["secretReadOnly"])
	data.OutputsInInternalStorage = optBool(out["outputsInInternalStorage"])
	data.StorageConfiguration = stringMapFromBody(out["storageConfiguration"])
	data.SecretConfiguration = dynamicFromBody(out["secretConfiguration"])
	data.AllowedNamespaces = allowedNamespacesToList(out)
	data.DefaultWorkerSelector = workerSelectorToList(out)
	data.StorageIsolation = isolationToList(out, "storageIsolation")
	data.SecretIsolation = isolationToList(out, "secretIsolation")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func allowedNamespacesToList(body map[string]interface{}) types.List {
	raw, ok := body["allowedNamespaces"].([]interface{})
	if !ok {
		return types.ListNull(allowedNamespaceObjectType)
	}
	elems := make([]attr.Value, 0, len(raw))
	for _, item := range raw {
		mp, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ns, ok := mp["namespace"].(string)
		if !ok {
			continue
		}
		obj, diags := types.ObjectValue(allowedNamespaceObjectType.AttrTypes, map[string]attr.Value{
			"namespace": types.StringValue(ns),
		})
		if diags.HasError() {
			continue
		}
		elems = append(elems, obj)
	}
	out, diags := types.ListValue(allowedNamespaceObjectType, elems)
	if diags.HasError() {
		return types.ListNull(allowedNamespaceObjectType)
	}
	return out
}

// yamlFromBody renders a decoded map back to the YAML string the schema exposes.
func yamlFromBody(raw interface{}) types.String {
	in, ok := raw.(map[string]interface{})
	if !ok || len(in) == 0 {
		return types.StringNull()
	}
	b, err := yaml.Marshal(in)
	if err != nil {
		return types.StringNull()
	}
	return types.StringValue(string(b))
}
