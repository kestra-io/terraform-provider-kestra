package provider_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ resource.Resource                 = &tenantResource{}
	_ resource.ResourceWithConfigure    = &tenantResource{}
	_ resource.ResourceWithImportState  = &tenantResource{}
	_ resource.ResourceWithUpgradeState = &tenantResource{}
)

func NewTenantResource() resource.Resource {
	return &tenantResource{}
}

type tenantResource struct {
	providerData ProviderData
}

type tenantModel struct {
	Id                       types.String     `tfsdk:"id"`
	TenantId                 types.String     `tfsdk:"tenant_id"`
	Name                     types.String     `tfsdk:"name"`
	DefaultWorkerSelector    []workerSelector `tfsdk:"default_worker_selector"`
	StorageType              types.String     `tfsdk:"storage_type"`
	StorageConfiguration     types.Map        `tfsdk:"storage_configuration"`
	StorageIsolation         []isolation      `tfsdk:"storage_isolation"`
	SecretIsolation          []isolation      `tfsdk:"secret_isolation"`
	SecretType               types.String     `tfsdk:"secret_type"`
	SecretReadOnly           types.Bool       `tfsdk:"secret_read_only"`
	SecretConfiguration      types.Map        `tfsdk:"secret_configuration"`
	RequireExistingNamespace types.Bool       `tfsdk:"require_existing_namespace"`
	OutputsInInternalStorage types.Bool       `tfsdk:"outputs_in_internal_storage"`
	Concurrency              []concurrency    `tfsdk:"concurrency"`
	Quotas                   []quota          `tfsdk:"quotas"`
}

func (r *tenantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant"
}

func (r *tenantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Tenant.\n\n-> This resource is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Version:             1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The tenant id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The tenant name.",
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
				MarkdownDescription: "Whether secrets are read-only in this tenant.",
			},
			"secret_configuration": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The secret configuration.",
			},
			"require_existing_namespace": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether tenant requires an existing namespace.",
			},
			"outputs_in_internal_storage": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether outputs are stored in internal storage.",
			},
		},
		Blocks: map[string]schema.Block{
			"default_worker_selector": schema.ListNestedBlock{
				MarkdownDescription: "The default routing applied to every task of the tenant that does not define its own. Tasks are routed to a `kestra_worker_queue` whose tag set matches.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"tags": schema.SetAttribute{
							Required:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "The tags used to route to a matching Worker Queue (each tag is an RFC 1123 label). The API rejects `match` and `fallback` without a non-empty tag set.",
							Validators:          []validator.Set{setvalidator.SizeAtLeast(1), setvalidator.SizeAtMost(20)},
						},
						"match": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "How the tags are matched against a Worker Queue tag set: `ALL` (default, the queue tags must be a superset) or `ANY` (they must intersect).",
							Validators:          []validator.String{stringvalidator.OneOf("ALL", "ANY")},
						},
						"fallback": schema.StringAttribute{
							Optional:            true,
							MarkdownDescription: "The strategy when no worker is available: `FAIL` (default), `WAIT`, `CANCEL` or `IGNORE`.",
							Validators:          []validator.String{stringvalidator.OneOf("FAIL", "WAIT", "CANCEL", "IGNORE")},
						},
					},
				},
			},
			"concurrency": concurrencyNestedBlock(),
			"quotas":      quotasNestedBlock(),
			"storage_isolation": schema.ListNestedBlock{
				MarkdownDescription: "Storage isolation configuration.",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject:        isolationNestedObject(),
			},
			"secret_isolation": schema.ListNestedBlock{
				MarkdownDescription: "Secret isolation configuration (same shape as storage_isolation).",
				Validators:          []validator.List{listvalidator.SizeAtMost(1)},
				NestedObject:        isolationNestedObject(),
			},
		},
	}
}

func (r *tenantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	pd, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.providerData = *pd
}

// Tenants are instance scoped, so unlike namespaces their path carries no tenant
// segment.
func (r *tenantResource) tenantPath(id string) string {
	if id == "" {
		return "/api/v1/tenants"
	}
	return "/api/v1/tenants/" + id
}

func (r *tenantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := tenantModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPost, r.tenantPath(""), body)
	if err != nil {
		resp.Diagnostics.AddError("Create tenant failed", err.Error())
		return
	}
	plannedConcurrency, plannedQuotas := plan.Concurrency, plan.Quotas
	resp.Diagnostics.Append(bodyToTenantModel(ctx, out, &plan)...)
	// A server that predates these fields (2.0.0-rc1) accepts the write but omits
	// them from the response. They are pure configuration, so the framework
	// requires the post-write state to match the plan; letting the response clear
	// them fails the apply outright. Read still treats an absent key as drift, so
	// a server that ignores them shows up as a diff on the next plan.
	plan.Concurrency, plan.Quotas = plannedConcurrency, plannedQuotas
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tenantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodGet, r.tenantPath(state.TenantId.ValueString()), nil)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read tenant failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToTenantModel(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tenantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tenantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := tenantModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPut, r.tenantPath(plan.TenantId.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Update tenant failed", err.Error())
		return
	}
	plannedConcurrency, plannedQuotas := plan.Concurrency, plan.Quotas
	resp.Diagnostics.Append(bodyToTenantModel(ctx, out, &plan)...)
	// A server that predates these fields (2.0.0-rc1) accepts the write but omits
	// them from the response. They are pure configuration, so the framework
	// requires the post-write state to match the plan; letting the response clear
	// them fails the apply outright. Read still treats an absent key as drift, so
	// a server that ignores them shows up as a diff on the next plan.
	plan.Concurrency, plan.Quotas = plannedConcurrency, plannedQuotas
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tenantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tenantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodDelete, r.tenantPath(state.TenantId.ValueString()), nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Delete tenant failed", err.Error())
	}
}

func (r *tenantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("tenant_id"), req, resp)
}

func (r *tenantResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeTenantStateV0,
		},
	}
}

// upgradeTenantStateV0 carries over state written by the SDK v2 implementation.
// The attribute names are unchanged, so this only has to re-read the raw JSON
// into the framework model; `concurrency` and `quotas` are new and stay null.
func upgradeTenantStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(req.RawState.JSON, &raw); err != nil {
		resp.Diagnostics.AddError("Failed to read prior state", err.Error())
		return
	}

	m := tenantModel{
		Id:                       optString(raw["id"]),
		TenantId:                 optString(raw["tenant_id"]),
		Name:                     optString(raw["name"]),
		StorageType:              optString(raw["storage_type"]),
		SecretType:               optString(raw["secret_type"]),
		SecretReadOnly:           optBool(raw["secret_read_only"]),
		RequireExistingNamespace: optBool(raw["require_existing_namespace"]),
		OutputsInInternalStorage: optBool(raw["outputs_in_internal_storage"]),
		StorageConfiguration:     types.MapNull(types.StringType),
		SecretConfiguration:      types.MapNull(types.StringType),
	}

	m.SecretConfiguration = stringMapFromV0(raw["secret_configuration"])

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

	if ws, ok := raw["default_worker_selector"].([]interface{}); ok && len(ws) > 0 {
		if mp, ok := ws[0].(map[string]interface{}); ok {
			m.DefaultWorkerSelector = []workerSelector{workerSelectorFromV0(mp)}
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

func tenantModelToBody(ctx context.Context, m *tenantModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	body := map[string]interface{}{
		"id":   m.TenantId.ValueString(),
		"name": m.Name.ValueString(),
	}

	if len(m.DefaultWorkerSelector) > 0 {
		sel := m.DefaultWorkerSelector[0]
		tags := make([]string, 0)
		diags.Append(sel.Tags.ElementsAs(ctx, &tags, false)...)
		ws := map[string]interface{}{"tags": tags}
		if !sel.Match.IsNull() && sel.Match.ValueString() != "" {
			ws["match"] = sel.Match.ValueString()
		}
		if !sel.Fallback.IsNull() && sel.Fallback.ValueString() != "" {
			ws["fallback"] = sel.Fallback.ValueString()
		}
		body["defaultWorkerSelector"] = ws
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
	if !m.SecretConfiguration.IsNull() {
		sc := map[string]string{}
		for k, v := range m.SecretConfiguration.Elements() {
			if str, ok := v.(types.String); ok {
				sc[k] = str.ValueString()
			}
		}
		if len(sc) > 0 {
			body["secretConfiguration"] = sc
		}
	}
	if !m.RequireExistingNamespace.IsNull() {
		body["requireExistingNamespace"] = m.RequireExistingNamespace.ValueBool()
	}
	if !m.OutputsInInternalStorage.IsNull() {
		body["outputsInInternalStorage"] = m.OutputsInInternalStorage.ValueBool()
	}
	if c := concurrencyToBody(m.Concurrency); c != nil {
		body["concurrency"] = c
	}
	if q := quotasToBody(m.Quotas); q != nil {
		body["quotas"] = q
	}
	return body, diags
}

func bodyToTenantModel(ctx context.Context, body map[string]interface{}, m *tenantModel) diag.Diagnostics {
	if id, ok := body["id"].(string); ok {
		m.TenantId = types.StringValue(id)
		m.Id = types.StringValue(id)
	}
	if name, ok := body["name"].(string); ok {
		m.Name = types.StringValue(name)
	}
	// an absent selector clears the model: the API omits null fields, so absence
	// means it is genuinely unset and must show up as drift
	if ws, ok := body["defaultWorkerSelector"].(map[string]interface{}); ok {
		m.DefaultWorkerSelector = []workerSelector{workerSelectorFromBody(ws)}
	} else {
		m.DefaultWorkerSelector = nil
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
		if mv, diags := basetypes.NewMapValue(types.StringType, els); !diags.HasError() {
			m.StorageConfiguration = mv
		}
	}
	if len(m.StorageIsolation) > 0 {
		if si, ok := body["storageIsolation"].(map[string]interface{}); ok {
			m.StorageIsolation = []isolation{isolationFromBody(ctx, si, m.StorageIsolation[0])}
		}
	}
	if len(m.SecretIsolation) > 0 {
		if si, ok := body["secretIsolation"].(map[string]interface{}); ok {
			m.SecretIsolation = []isolation{isolationFromBody(ctx, si, m.SecretIsolation[0])}
		}
	}
	if st, ok := body["secretType"].(string); ok {
		m.SecretType = types.StringValue(st)
	}
	if sro, ok := body["secretReadOnly"].(bool); ok {
		m.SecretReadOnly = types.BoolValue(sro)
	}
	m.SecretConfiguration = stringMapFromBody(body["secretConfiguration"])
	if ren, ok := body["requireExistingNamespace"].(bool); ok {
		m.RequireExistingNamespace = types.BoolValue(ren)
	}
	if oi, ok := body["outputsInInternalStorage"].(bool); ok {
		m.OutputsInInternalStorage = types.BoolValue(oi)
	}
	m.Concurrency = concurrencyFromBody(body)
	m.Quotas = quotasFromBody(body)
	return nil
}
