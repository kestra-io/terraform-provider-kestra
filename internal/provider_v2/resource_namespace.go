package provider_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	listvalidator "github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
	"gopkg.in/yaml.v2"
)

var (
	_ resource.Resource                   = &namespaceResource{}
	_ resource.ResourceWithImportState    = &namespaceResource{}
	_ resource.ResourceWithConfigure      = &namespaceResource{}
	_ resource.ResourceWithUpgradeState   = &namespaceResource{}
)

func NewNamespaceResource() resource.Resource {
	return &namespaceResource{}
}

type namespaceResource struct {
	providerData ProviderData
}

type namespaceModel struct {
	Id                       types.String  `tfsdk:"id"`
	TenantId                 types.String  `tfsdk:"tenant_id"`
	NamespaceId              types.String  `tfsdk:"namespace_id"`
	Description              types.String  `tfsdk:"description"`
	Variables                types.String  `tfsdk:"variables"`
	PluginDefaults           types.String  `tfsdk:"plugin_defaults"`
	AllowedNamespaces        []allowedNS   `tfsdk:"allowed_namespaces"`
	WorkerGroup              []workerGroup `tfsdk:"worker_group"`
	StorageType              types.String  `tfsdk:"storage_type"`
	StorageConfiguration     types.Map     `tfsdk:"storage_configuration"`
	StorageIsolation         []isolation   `tfsdk:"storage_isolation"`
	SecretIsolation          []isolation   `tfsdk:"secret_isolation"`
	SecretType               types.String  `tfsdk:"secret_type"`
	SecretReadOnly           types.Bool    `tfsdk:"secret_read_only"`
	SecretConfiguration      types.Dynamic `tfsdk:"secret_configuration"`
	OutputsInInternalStorage types.Bool    `tfsdk:"outputs_in_internal_storage"`
}

type allowedNS struct {
	Namespace types.String `tfsdk:"namespace"`
}

type workerGroup struct {
	Key      types.String `tfsdk:"key"`
	Fallback types.String `tfsdk:"fallback"`
}

type isolation struct {
	Enabled        types.Bool `tfsdk:"enabled"`
	DeniedServices types.Set `tfsdk:"denied_services"`
}

func (r *namespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *namespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Namespace.\n\n-> This resource is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The namespace id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"namespace_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The namespace id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The namespace friendly description.",
			},
			"variables": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The namespace variables in yaml string.",
				PlanModifiers:       []planmodifier.String{YamlEqualPlanModifier()},
			},
			"plugin_defaults": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The namespace plugin defaults in yaml string.",
				PlanModifiers:       []planmodifier.String{YamlEqualPlanModifier()},
			},
			"storage_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The storage type.",
			},
			"storage_configuration": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The storage configuration.",
			},
			"secret_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The secret type.",
			},
			"secret_read_only": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether secrets are read-only in this namespace.",
			},
			"secret_configuration": schema.DynamicAttribute{
				Optional:            true,
				MarkdownDescription: "Per-backend secret configuration. The whole value is a free-form map keyed by backend type (e.g. `vault`, `aws`, `gcp`), where each value is either a string or a nested object describing that backend's config.",
			},
			"outputs_in_internal_storage": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether outputs are stored in internal storage.",
			},
		},
		Blocks: map[string]schema.Block{
			"allowed_namespaces": schema.ListNestedBlock{
				MarkdownDescription: "The allowed namespaces.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"namespace": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The namespace.",
						},
					},
				},
			},
			"worker_group": schema.ListNestedBlock{
				MarkdownDescription: "The worker group.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The worker group key.",
						},
						"fallback": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The fallback strategy.",
						},
					},
				},
			},
			"storage_isolation": schema.ListNestedBlock{
				MarkdownDescription: "Storage isolation configuration.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject:        isolationNestedObject(),
			},
			"secret_isolation": schema.ListNestedBlock{
				MarkdownDescription: "Secret isolation configuration.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject:        isolationNestedObject(),
			},
		},
	}
}

func isolationNestedObject() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether isolation is enabled.",
			},
			"denied_services": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Set of denied services.",
			},
		},
	}
}

func (r *namespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.providerData = *pd
}

func (r *namespaceResource) namespacePath(id string) string {
	return fmt.Sprintf("/api/v1/%s/namespaces", r.providerData.TenantId) + func() string {
		if id == "" {
			return ""
		}
		return "/" + id
	}()
}

func (r *namespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := namespaceModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPost, r.namespacePath(""), body)
	if err != nil {
		resp.Diagnostics.AddError("Create namespace failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToNamespaceModel(ctx, out, r.providerData.TenantId, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodGet, r.namespacePath(state.NamespaceId.ValueString()), nil)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read namespace failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToNamespaceModel(ctx, out, r.providerData.TenantId, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *namespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan namespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := namespaceModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPut, r.namespacePath(plan.NamespaceId.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Update namespace failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToNamespaceModel(ctx, out, r.providerData.TenantId, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *namespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state namespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodDelete, r.namespacePath(state.NamespaceId.ValueString()), nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Delete namespace failed", err.Error())
	}
}

func (r *namespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *namespaceResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeNamespaceStateV0,
		},
	}
}

func upgradeNamespaceStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(req.RawState.JSON, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read prior state", err.Error())
		return
	}

	m := namespaceModel{
		Id:                       optString(raw["id"]),
		TenantId:                 optString(raw["tenant_id"]),
		NamespaceId:              optString(raw["namespace_id"]),
		Description:              optString(raw["description"]),
		Variables:                optString(raw["variables"]),
		PluginDefaults:           optString(raw["plugin_defaults"]),
		StorageType:              optString(raw["storage_type"]),
		SecretType:               optString(raw["secret_type"]),
		SecretReadOnly:           optBool(raw["secret_read_only"]),
		OutputsInInternalStorage: optBool(raw["outputs_in_internal_storage"]),
		StorageConfiguration:     types.MapNull(types.StringType),
		SecretConfiguration:      types.DynamicNull(),
	}

	if an, ok := raw["allowed_namespaces"].([]interface{}); ok {
		out := make([]allowedNS, 0, len(an))
		for _, item := range an {
			if mp, ok := item.(map[string]interface{}); ok {
				if ns, ok := mp["namespace"].(string); ok {
					out = append(out, allowedNS{Namespace: types.StringValue(ns)})
				}
			}
		}
		m.AllowedNamespaces = out
	}

	if wg, ok := raw["worker_group"].([]interface{}); ok && len(wg) > 0 {
		if mp, ok := wg[0].(map[string]interface{}); ok {
			one := workerGroup{Fallback: types.StringNull()}
			if k, ok := mp["key"].(string); ok {
				one.Key = types.StringValue(k)
			}
			if fb, ok := mp["fallback"].(string); ok && fb != "" {
				one.Fallback = types.StringValue(fb)
			}
			m.WorkerGroup = []workerGroup{one}
		}
	}

	if sc, ok := raw["storage_configuration"].(map[string]interface{}); ok && len(sc) > 0 {
		els := map[string]attr.Value{}
		for k, v := range sc {
			if s, ok := v.(string); ok {
				els[k] = types.StringValue(s)
			}
		}
		if mv, diags := basetypes.NewMapValue(types.StringType, els); !diags.HasError() {
			m.StorageConfiguration = mv
		}
	}

	if si, ok := raw["storage_isolation"].([]interface{}); ok && len(si) > 0 {
		if mp, ok := si[0].(map[string]interface{}); ok {
			m.StorageIsolation = []isolation{isolationFromV0(mp)}
		}
	}
	if si, ok := raw["secret_isolation"].([]interface{}); ok && len(si) > 0 {
		if mp, ok := si[0].(map[string]interface{}); ok {
			m.SecretIsolation = []isolation{isolationFromV0(mp)}
		}
	}

	if sc, ok := raw["secret_configuration"].(map[string]interface{}); ok && len(sc) > 0 {
		if dv, err := goValueToDynamic(sc); err == nil {
			m.SecretConfiguration = dv
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func optString(v interface{}) types.String {
	if s, ok := v.(string); ok {
		return types.StringValue(s)
	}
	return types.StringNull()
}

func optBool(v interface{}) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}

func isolationFromV0(in map[string]interface{}) isolation {
	out := isolation{Enabled: types.BoolNull(), DeniedServices: types.SetNull(types.StringType)}
	if en, ok := in["enabled"].(bool); ok {
		out.Enabled = types.BoolValue(en)
	}
	if ds, ok := in["denied_services"].([]interface{}); ok && len(ds) > 0 {
		vals := make([]attr.Value, 0, len(ds))
		for _, v := range ds {
			if s, ok := v.(string); ok {
				vals = append(vals, types.StringValue(s))
			}
		}
		if sv, diags := basetypes.NewSetValue(types.StringType, vals); !diags.HasError() {
			out.DeniedServices = sv
		}
	}
	return out
}

func namespaceModelToBody(ctx context.Context, m *namespaceModel) (map[string]interface{}, diag.Diagnostics) {
	body := map[string]interface{}{
		"id": m.NamespaceId.ValueString(),
	}
	if !m.Description.IsNull() {
		body["description"] = m.Description.ValueString()
	}

	if !m.Variables.IsNull() && m.Variables.ValueString() != "" {
		var v interface{}
		if err := yaml.Unmarshal([]byte(m.Variables.ValueString()), &v); err == nil {
			body["variables"] = normalizeYAML(v)
		}
	}

	if !m.PluginDefaults.IsNull() && m.PluginDefaults.ValueString() != "" {
		var v interface{}
		if err := yaml.Unmarshal([]byte(m.PluginDefaults.ValueString()), &v); err == nil {
			body["pluginDefaults"] = normalizeYAML(v)
		}
	}

	allowed := make([]map[string]interface{}, len(m.AllowedNamespaces))
	for i, an := range m.AllowedNamespaces {
		allowed[i] = map[string]interface{}{"namespace": an.Namespace.ValueString()}
	}
	body["allowedNamespaces"] = allowed

	if len(m.WorkerGroup) > 0 {
		wg := map[string]interface{}{"key": m.WorkerGroup[0].Key.ValueString()}
		if !m.WorkerGroup[0].Fallback.IsNull() && m.WorkerGroup[0].Fallback.ValueString() != "" {
			wg["fallback"] = m.WorkerGroup[0].Fallback.ValueString()
		}
		body["workerGroup"] = wg
	}
	if !m.StorageType.IsNull() && m.StorageType.ValueString() != "" {
		body["storageType"] = m.StorageType.ValueString()
	}
	if !m.StorageConfiguration.IsNull() {
		sc := map[string]string{}
		for k, v := range m.StorageConfiguration.Elements() {
			if s, ok := v.(types.String); ok {
				sc[k] = s.ValueString()
			}
		}
		if len(sc) > 0 {
			body["storageConfiguration"] = sc
		}
	}
	if len(m.StorageIsolation) > 0 {
		body["storageIsolation"] = isolationToBody(ctx, m.StorageIsolation[0])
	}
	if len(m.SecretIsolation) > 0 {
		body["secretIsolation"] = isolationToBody(ctx, m.SecretIsolation[0])
	}
	if !m.SecretType.IsNull() && m.SecretType.ValueString() != "" {
		body["secretType"] = m.SecretType.ValueString()
	}
	if !m.SecretReadOnly.IsNull() {
		body["secretReadOnly"] = m.SecretReadOnly.ValueBool()
	}
	if !m.SecretConfiguration.IsNull() && !m.SecretConfiguration.IsUnknown() {
		under := m.SecretConfiguration.UnderlyingValue()
		if under != nil {
			tfv, err := under.ToTerraformValue(ctx)
			if err == nil {
				if gv, err := tfTypeValueToGo(tfv); err == nil && gv != nil {
					body["secretConfiguration"] = gv
				}
			}
		}
	}
	if !m.OutputsInInternalStorage.IsNull() {
		body["outputsInInternalStorage"] = m.OutputsInInternalStorage.ValueBool()
	}
	return body, nil
}

func isolationToBody(ctx context.Context, iso isolation) map[string]interface{} {
	out := map[string]interface{}{}
	if !iso.Enabled.IsNull() {
		out["enabled"] = iso.Enabled.ValueBool()
	}
	if !iso.DeniedServices.IsNull() {
		ds := []string{}
		for _, e := range iso.DeniedServices.Elements() {
			if s, ok := e.(types.String); ok {
				ds = append(ds, s.ValueString())
			}
		}
		if len(ds) > 0 {
			out["deniedServices"] = ds
		}
	}
	return out
}

func normalizeYAML(in interface{}) interface{} {
	switch x := in.(type) {
	case map[interface{}]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, v := range x {
			if ks, ok := k.(string); ok {
				out[ks] = normalizeYAML(v)
			}
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(x))
		for k, v := range x {
			out[k] = normalizeYAML(v)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, item := range x {
			out[i] = normalizeYAML(item)
		}
		return out
	}
	return in
}

func tfTypeValueToGo(v tftypes.Value) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	ty := v.Type()
	switch {
	case ty.Is(tftypes.String):
		var s string
		if err := v.As(&s); err != nil {
			return nil, err
		}
		return s, nil
	case ty.Is(tftypes.Number):
		bf := new(big.Float)
		if err := v.As(&bf); err != nil {
			return nil, err
		}
		if i, acc := bf.Int64(); acc == big.Exact {
			return i, nil
		}
		f, _ := bf.Float64()
		return f, nil
	case ty.Is(tftypes.Bool):
		var b bool
		if err := v.As(&b); err != nil {
			return nil, err
		}
		return b, nil
	case ty.Is(tftypes.List{}), ty.Is(tftypes.Set{}), ty.Is(tftypes.Tuple{}):
		var vs []tftypes.Value
		if err := v.As(&vs); err != nil {
			return nil, err
		}
		out := make([]interface{}, len(vs))
		for i, item := range vs {
			g, err := tfTypeValueToGo(item)
			if err != nil {
				return nil, err
			}
			out[i] = g
		}
		return out, nil
	case ty.Is(tftypes.Map{}), ty.Is(tftypes.Object{}):
		var vs map[string]tftypes.Value
		if err := v.As(&vs); err != nil {
			return nil, err
		}
		out := map[string]interface{}{}
		for k, item := range vs {
			g, err := tfTypeValueToGo(item)
			if err != nil {
				return nil, err
			}
			out[k] = g
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported tftype %s", ty.String())
}

func bodyToNamespaceModel(ctx context.Context, body map[string]interface{}, tenantId string, m *namespaceModel) diag.Diagnostics {
	if id, ok := body["id"].(string); ok {
		m.NamespaceId = types.StringValue(id)
		m.Id = types.StringValue(id)
	}
	m.TenantId = types.StringValue(tenantId)
	if d, ok := body["description"].(string); ok {
		m.Description = types.StringValue(d)
	}
	if vars, ok := body["variables"].(map[string]interface{}); ok && len(vars) > 0 {
		if m.Variables.IsNull() || m.Variables.IsUnknown() {
			if b, err := yaml.Marshal(vars); err == nil {
				m.Variables = types.StringValue(string(b))
			}
		}
	}
	if pd, ok := body["pluginDefaults"]; ok && pd != nil {
		if m.PluginDefaults.IsNull() || m.PluginDefaults.IsUnknown() {
			if b, err := yaml.Marshal(pd); err == nil {
				m.PluginDefaults = types.StringValue(string(b))
			}
		}
	}
	if an, ok := body["allowedNamespaces"].([]interface{}); ok {
		out := make([]allowedNS, 0, len(an))
		for _, item := range an {
			if mp, ok := item.(map[string]interface{}); ok {
				if ns, ok := mp["namespace"].(string); ok {
					out = append(out, allowedNS{Namespace: types.StringValue(ns)})
				}
			}
		}
		m.AllowedNamespaces = out
	}
	if wg, ok := body["workerGroup"].(map[string]interface{}); ok {
		one := workerGroup{Fallback: types.StringNull()}
		if k, ok := wg["key"].(string); ok {
			one.Key = types.StringValue(k)
		}
		if fb, ok := wg["fallback"].(string); ok && fb != "" {
			one.Fallback = types.StringValue(fb)
		}
		m.WorkerGroup = []workerGroup{one}
	}
	if st, ok := body["storageType"].(string); ok {
		m.StorageType = types.StringValue(st)
	}
	if sc, ok := body["storageConfiguration"].(map[string]interface{}); ok && len(sc) > 0 {
		els := map[string]attr.Value{}
		for k, v := range sc {
			if s, ok := v.(string); ok {
				els[k] = types.StringValue(s)
			}
		}
		mv, diags := basetypes.NewMapValue(types.StringType, els)
		if !diags.HasError() {
			m.StorageConfiguration = mv
		}
	}
	if len(m.StorageIsolation) > 0 {
		if si, ok := body["storageIsolation"].(map[string]interface{}); ok {
			m.StorageIsolation = []isolation{isolationFromBody(ctx, si)}
		}
	}
	if len(m.SecretIsolation) > 0 {
		if si, ok := body["secretIsolation"].(map[string]interface{}); ok {
			m.SecretIsolation = []isolation{isolationFromBody(ctx, si)}
		}
	}
	if st, ok := body["secretType"].(string); ok {
		m.SecretType = types.StringValue(st)
	}
	if sro, ok := body["secretReadOnly"].(bool); ok {
		m.SecretReadOnly = types.BoolValue(sro)
	}
	if sc, ok := body["secretConfiguration"].(map[string]interface{}); ok && len(sc) > 0 {
		dv, err := goValueToDynamic(sc)
		if err == nil {
			m.SecretConfiguration = dv
		}
	}
	if oi, ok := body["outputsInInternalStorage"].(bool); ok {
		m.OutputsInInternalStorage = types.BoolValue(oi)
	}
	return nil
}

func isolationFromBody(_ context.Context, in map[string]interface{}) isolation {
	out := isolation{Enabled: types.BoolNull(), DeniedServices: types.SetNull(types.StringType)}
	if en, ok := in["enabled"].(bool); ok {
		out.Enabled = types.BoolValue(en)
	}
	if ds, ok := in["deniedServices"].([]interface{}); ok && len(ds) > 0 {
		strs := make([]attr.Value, 0, len(ds))
		for _, v := range ds {
			if s, ok := v.(string); ok {
				strs = append(strs, types.StringValue(s))
			}
		}
		if sv, diags := basetypes.NewSetValue(types.StringType, strs); !diags.HasError() {
			out.DeniedServices = sv
		}
	}
	return out
}

func goValueToDynamic(v interface{}) (types.Dynamic, error) {
	switch x := v.(type) {
	case nil:
		return types.DynamicNull(), nil
	case string:
		return types.DynamicValue(types.StringValue(x)), nil
	case bool:
		return types.DynamicValue(types.BoolValue(x)), nil
	case float64:
		return types.DynamicValue(types.Float64Value(x)), nil
	case map[string]interface{}:
		attrTypes := map[string]attr.Type{}
		attrs := map[string]attr.Value{}
		for k, val := range x {
			dv, err := goValueToDynamic(val)
			if err != nil {
				return types.DynamicNull(), err
			}
			attrTypes[k] = types.DynamicType
			attrs[k] = dv
		}
		ov, diags := basetypes.NewObjectValue(attrTypes, attrs)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("object value: %s", diags)
		}
		return types.DynamicValue(ov), nil
	case []interface{}:
		els := make([]attr.Value, 0, len(x))
		for _, item := range x {
			dv, err := goValueToDynamic(item)
			if err != nil {
				return types.DynamicNull(), err
			}
			els = append(els, dv)
		}
		tv, diags := basetypes.NewTupleValue(make([]attr.Type, len(els)), els)
		if diags.HasError() {
			return types.DynamicNull(), fmt.Errorf("tuple value: %s", diags)
		}
		return types.DynamicValue(tv), nil
	}
	b, _ := json.Marshal(v)
	return types.DynamicValue(types.StringValue(string(b))), nil
}
