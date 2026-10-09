package provider_v2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kestra-io/client-sdk/go-sdk/v2/kestra_api_client"
)

var (
	_ resource.Resource                   = &flowResource{}
	_ resource.ResourceWithConfigure      = &flowResource{}
	_ resource.ResourceWithIdentity       = &flowResource{}
	_ resource.ResourceWithImportState    = &flowResource{}
	_ resource.ResourceWithModifyPlan     = &flowResource{}
	_ resource.ResourceWithUpgradeState   = &flowResource{}
	_ resource.ResourceWithValidateConfig = &flowResource{}
)

func NewFlowResource() resource.Resource {
	return &flowResource{}
}

type flowResource struct {
	providerData ProviderData
}

type flowIdentityModel struct {
	Namespace types.String `tfsdk:"namespace"`
	FlowId    types.String `tfsdk:"flow_id"`
}

type flowModel struct {
	Id          types.String `tfsdk:"id"`
	TenantId    types.String `tfsdk:"tenant_id"`
	Namespace   types.String `tfsdk:"namespace"`
	FlowId      types.String `tfsdk:"flow_id"`
	Revision    types.Int64  `tfsdk:"revision"`
	Content     types.String `tfsdk:"content"`
	Disabled    types.Bool   `tfsdk:"disabled"`
	Description types.String `tfsdk:"description"`
	Labels      types.Map    `tfsdk:"labels"`
}

func (m flowModel) identity() flowIdentityModel {
	return flowIdentityModel{Namespace: m.Namespace, FlowId: m.FlowId}
}

func (r *flowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flow"
}

func (r *flowResource) IdentitySchema(_ context.Context, _ resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	// version 0, like kestra_namespace: state written by releases without identity asks for
	// an upgrade from version 0, which no upgrader can answer since the prior identity is empty
	resp.IdentitySchema = identityschema.Schema{
		Version: 0,
		Attributes: map[string]identityschema.Attribute{
			"namespace": identityschema.StringAttribute{
				RequiredForImport: true,
			},
			"flow_id": identityschema.StringAttribute{
				RequiredForImport: true,
			},
		},
	}
}

func (r *flowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Flow.\n\n" +
			"The flow `description`, `disabled` and `labels` can be written in `content` or set as attributes, but not both. " +
			"An attribute is merged into the source sent to Kestra and left out of `content` when read back.",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The flow resource id, `namespace/flow_id`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tenant_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The tenant id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"namespace": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The flow namespace. Defaults to the `namespace` in `content`, which it must match. " +
					"Set it when `content` can be unknown during planning, or an update of the flow plans a replacement.",
			},
			"flow_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: "The flow id. Defaults to the `id` in `content`, which it must match. " +
					"Set it when `content` can be unknown during planning, or an update of the flow plans a replacement.",
			},
			"revision": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The flow revision.",
			},
			"content": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The flow full content in yaml string.",
				PlanModifiers:       []planmodifier.String{YamlEqualPlanModifier()},
			},
			"disabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether the flow is disabled. Leave `disabled` out of `content` when set.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The flow description. Leave `description` out of `content` when set.",
			},
			"labels": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The flow labels. Leave `labels` out of `content` when set.",
			},
		},
	}
}

func (r *flowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ProviderData type, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerData = *providerData
}

func (r *flowResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config flowModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Content.IsNull() || config.Content.IsUnknown() {
		return
	}

	document, values, err := parseFlowSource(config.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("content"), "Invalid flow content", err.Error())
		return
	}

	// rejected even when both values match: as agreed in #120, neither one takes precedence
	for _, metadata := range []struct {
		key   string
		value attr.Value
	}{
		{"description", config.Description},
		{"disabled", config.Disabled},
		{"labels", config.Labels},
	} {
		if _, inContent := values[metadata.key]; inContent && !metadata.value.IsNull() {
			resp.Diagnostics.AddAttributeError(
				path.Root(metadata.key),
				"Conflicting flow metadata",
				fmt.Sprintf("`%s` is set both as an attribute and in content. Remove it from one of them.", metadata.key),
			)
		}
	}

	for _, identity := range []struct {
		key       string
		attribute string
		value     types.String
	}{
		{"namespace", "namespace", config.Namespace},
		{"id", "flow_id", config.FlowId},
	} {
		inContent, ok := flowScalar(document, identity.key)
		if !ok {
			resp.Diagnostics.AddAttributeError(path.Root("content"), "Invalid flow content", fmt.Sprintf("The flow must define `%s`.", identity.key))
			continue
		}
		if !identity.value.IsNull() && !identity.value.IsUnknown() && identity.value.ValueString() != inContent {
			resp.Diagnostics.AddAttributeError(
				path.Root(identity.attribute),
				"Inconsistent flow identity",
				fmt.Sprintf("%s is %q but the flow %s in content is %q.", identity.attribute, identity.value.ValueString(), identity.key, inContent),
			)
		}
	}
}

// ModifyPlan derives namespace and flow_id from content when they are not configured, so
// renaming the flow in its source replaces it like changing the attributes does.
func (r *flowResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var config, plan flowModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// content that does not parse is reported by ValidateConfig
	if !config.Content.IsUnknown() {
		if document, _, err := parseFlowSource(config.Content.ValueString()); err == nil {
			if namespace, ok := flowScalar(document, "namespace"); ok && config.Namespace.IsNull() {
				plan.Namespace = types.StringValue(namespace)
			}
			if flowId, ok := flowScalar(document, "id"); ok && config.FlowId.IsNull() {
				plan.FlowId = types.StringValue(flowId)
			}
		}
	}

	if !req.State.Raw.IsNull() {
		var state flowModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if !plan.Namespace.Equal(state.Namespace) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("namespace"))
		}
		if !plan.FlowId.Equal(state.FlowId) {
			resp.RequiresReplace = append(resp.RequiresReplace, path.Root("flow_id"))
		}

		// a formatting-only change to content is suppressed by its plan modifier, but the
		// revision was already marked unknown; keep it so no update is planned
		if plan.Content.Equal(state.Content) && plan.Namespace.Equal(state.Namespace) && plan.FlowId.Equal(state.FlowId) &&
			plan.Disabled.Equal(state.Disabled) && plan.Description.Equal(state.Description) && plan.Labels.Equal(state.Labels) {
			plan.Revision = state.Revision
		}
	}

	resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
}

func isFlowNotFound(err error) bool {
	var apiErr *kestra_api_client.ApiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// sourceForWrite returns content with the metadata attributes merged in, once the server
// validated it: the server knows the plugins and their properties, so an invalid flow fails
// before anything is saved.
func (r *flowResource) sourceForWrite(ctx context.Context, plan flowModel) (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	metadata := flowMetadata{Description: plan.Description.ValueStringPointer(), Disabled: plan.Disabled.ValueBoolPointer()}
	if !plan.Labels.IsNull() {
		labels := map[string]string{}
		if diags = plan.Labels.ElementsAs(ctx, &labels, false); diags.HasError() {
			return "", diags
		}
		metadata.Labels = &labels
	}
	source, err := mergeFlowMetadata(plan.Content.ValueString(), metadata)
	if err != nil {
		diags.AddAttributeError(path.Root("content"), "Invalid flow content", err.Error())
		return "", diags
	}

	violations, err := r.providerData.KestraClient.Flows().ValidateFlows(ctx, r.providerData.TenantId, source)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to validate flow, got error: %s", err))
		return "", diags
	}
	for _, violation := range violations {
		if constraints := violation.GetConstraints(); constraints != "" {
			diags.AddAttributeError(path.Root("content"), "Invalid flow", constraints)
		}
		for _, warning := range violation.GetWarnings() {
			diags.AddAttributeWarning(path.Root("content"), "Flow validation warning", warning)
		}
		for _, deprecated := range violation.GetDeprecationPaths() {
			diags.AddAttributeWarning(path.Root("content"), "Deprecated flow property", fmt.Sprintf("%s is deprecated.", deprecated))
		}
		for _, info := range violation.GetInfos() {
			diags.AddAttributeWarning(path.Root("content"), "Flow validation info", info)
		}
	}
	return source, diags
}

// setFlowFromAPI copies the identity and revision of a flow returned by the API.
func setFlowFromAPI(data *flowModel, flow *kestra_api_client.FlowWithSource, tenantId string) {
	data.Id = types.StringValue(flow.GetNamespace() + "/" + flow.GetId())
	data.TenantId = types.StringValue(tenantId)
	data.Namespace = types.StringValue(flow.GetNamespace())
	data.FlowId = types.StringValue(flow.GetId())
	data.Revision = types.Int64Value(int64(flow.GetRevision()))
}

// populateFlowModel refreshes the model from a flow read with its source. Keys managed by an
// attribute are moved out of the source into the attribute; content keeps its configured
// text unless the remaining source differs from it.
func populateFlowModel(ctx context.Context, data *flowModel, flow *kestra_api_client.FlowWithSource, tenantId string) diag.Diagnostics {
	var diags diag.Diagnostics
	readError := func(err error) diag.Diagnostics {
		diags.AddError("Client Error", fmt.Sprintf("Unable to read flow %s/%s: %s", flow.GetNamespace(), flow.GetId(), err))
		return diags
	}
	if flow.Source == nil {
		return readError(errors.New("the API returned no source"))
	}

	source := flow.GetSource()
	document, values, err := parseFlowSource(source)
	if err != nil {
		return readError(err)
	}

	var managed []string
	if !data.Disabled.IsNull() {
		data.Disabled = types.BoolValue(flow.GetDisabled())
		managed = append(managed, "disabled")
	}
	if !data.Description.IsNull() {
		data.Description = types.StringValue(flow.GetDescription())
		managed = append(managed, "description")
	}
	if !data.Labels.IsNull() {
		labels, err := flowLabels(values["labels"])
		if err != nil {
			return readError(err)
		}
		var labelDiags diag.Diagnostics
		data.Labels, labelDiags = types.MapValueFrom(ctx, types.StringType, labels)
		diags.Append(labelDiags...)
		managed = append(managed, "labels")
	}

	if len(managed) > 0 {
		removeFlowKeys(document, managed)
		if source, err = encodeFlowSource(document); err != nil {
			return readError(err)
		}
	}
	if data.Content.IsNull() || !flowSourcesEqual(data.Content.ValueString(), source) {
		data.Content = types.StringValue(source)
	}
	setFlowFromAPI(data, flow, tenantId)
	return diags
}

func (r *flowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan flowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, diags := r.sourceForWrite(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.providerData.KestraClient.Flows().CreateFlow(ctx, r.providerData.TenantId, source)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create flow, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created a flow resource: %s/%s", created.GetNamespace(), created.GetId()))

	// content stays as planned: the API keeps the source as sent
	setFlowFromAPI(&plan, created, r.providerData.TenantId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, plan.identity())...)
}

func (r *flowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state flowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	withSource := true
	flow, err := r.providerData.KestraClient.Flows().Flow(ctx, state.Namespace.ValueString(), state.FlowId.ValueString(), r.providerData.TenantId, &withSource, nil, nil)
	if err != nil {
		if isFlowNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read flow, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a flow resource: %s/%s", flow.GetNamespace(), flow.GetId()))

	resp.Diagnostics.Append(populateFlowModel(ctx, &state, flow, r.providerData.TenantId)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, state.identity())...)
}

func (r *flowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan flowModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	source, diags := r.sourceForWrite(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.providerData.KestraClient.Flows().UpdateFlow(ctx, plan.Namespace.ValueString(), plan.FlowId.ValueString(), r.providerData.TenantId, source)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update flow, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated a flow resource: %s/%s", updated.GetNamespace(), updated.GetId()))

	setFlowFromAPI(&plan, updated, r.providerData.TenantId)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Identity.Set(ctx, plan.identity())...)
}

func (r *flowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state flowModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.KestraClient.Flows().DeleteFlow(ctx, state.Namespace.ValueString(), state.FlowId.ValueString(), r.providerData.TenantId)
	if err != nil && !isFlowNotFound(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete flow, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted a flow resource: %s/%s", state.Namespace.ValueString(), state.FlowId.ValueString()))
}

// ImportState accepts the id namespace/flow_id, or the namespace and flow_id identity.
func (r *flowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var namespace, flowId string
	if req.ID != "" {
		var found bool
		namespace, flowId, found = strings.Cut(req.ID, "/")
		if !found || namespace == "" || flowId == "" {
			resp.Diagnostics.AddError("Invalid import id", fmt.Sprintf("Expected namespace/flow_id, got: %s", req.ID))
			return
		}
	} else {
		var identity flowIdentityModel
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
		namespace, flowId = identity.Namespace.ValueString(), identity.FlowId.ValueString()
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), namespace+"/"+flowId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("flow_id"), flowId)...)
}

func (r *flowResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeFlowStateV0,
		},
	}
}

// upgradeFlowStateV0 reads state written by the SDK v2 implementation. It had no metadata
// attributes, so they start unset and content keeps managing those keys.
func upgradeFlowStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior struct {
		Id        *string `json:"id"`
		TenantId  *string `json:"tenant_id"`
		Namespace *string `json:"namespace"`
		FlowId    *string `json:"flow_id"`
		Revision  *int64  `json:"revision"`
		Content   *string `json:"content"`
	}
	if err := json.Unmarshal(req.RawState.JSON, &prior); err != nil {
		resp.Diagnostics.AddError("Failed to read prior state", err.Error())
		return
	}

	state := flowModel{
		Id:          types.StringPointerValue(prior.Id),
		TenantId:    types.StringPointerValue(prior.TenantId),
		Namespace:   types.StringPointerValue(prior.Namespace),
		FlowId:      types.StringPointerValue(prior.FlowId),
		Revision:    types.Int64PointerValue(prior.Revision),
		Content:     types.StringPointerValue(prior.Content),
		Disabled:    types.BoolNull(),
		Description: types.StringNull(),
		Labels:      types.MapNull(types.StringType),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
