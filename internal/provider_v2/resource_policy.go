package provider_v2

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &policyResource{}
var _ resource.ResourceWithValidateConfig = &policyResource{}
var _ resource.ResourceWithImportState = &policyResource{}

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

// policyResource defines the resource implementation.
type policyResource struct {
	providerData ProviderData
}

// policyResourceModel describes the resource data model.
type policyResourceModel struct {
	Scope     types.String `tfsdk:"scope"`
	PolicyId  types.String `tfsdk:"policy_id"`
	TenantId  types.String `tfsdk:"tenant_id"`
	Namespace types.String `tfsdk:"namespace"`
	Content   types.String `tfsdk:"content"`
}

func (r *policyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

// resolvePolicyTenantId resolves the tenant used in the request URL: never one for the
// INSTANCE scope, otherwise the configured tenant_id falling back to the provider tenant.
func resolvePolicyTenantId(providerData ProviderData, data policyResourceModel) string {
	if data.Scope.ValueString() == policyScopeInstance {
		return ""
	}
	if !data.TenantId.IsNull() && !data.TenantId.IsUnknown() && data.TenantId.ValueString() != "" {
		return data.TenantId.ValueString()
	}
	return providerData.TenantId
}

func (r *policyResource) tenantId(data policyResourceModel) string {
	return resolvePolicyTenantId(r.providerData, data)
}

// resolveTenantId fills the computed tenant_id after a write: null for the INSTANCE scope,
// otherwise the tenant the request was sent to.
func (r *policyResource) resolveTenantId(data *policyResourceModel) {
	if data.Scope.ValueString() == policyScopeInstance {
		data.TenantId = types.StringNull()
		return
	}
	data.TenantId = types.StringValue(r.tenantId(*data))
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra governance Policy (EE) at the `INSTANCE`, `TENANT` or `NAMESPACE` scope. " +
			"A policy bundles mutate and validate rules applied to flows and plugins; its YAML source is persisted " +
			"by the API and round-tripped verbatim. " +
			"`STATIC` policies are declared in the Kestra configuration and cannot be managed through the API.",

		Attributes: map[string]schema.Attribute{
			"scope": schema.StringAttribute{
				MarkdownDescription: "The policy scope: `INSTANCE` (deployment-wide, super-admin only), `TENANT` or `NAMESPACE`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(policyScopeInstance, policyScopeTenant, policyScopeNamespace),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"policy_id": schema.StringAttribute{
				MarkdownDescription: "The policy id — a lowercase RFC 1123 label, unique per (scope, tenant, namespace). Must match the `id` of the YAML content.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "The tenant id, for `TENANT` and `NAMESPACE` scopes. Defaults to the provider tenant when omitted; the value is captured at create time, so changing the provider tenant later does not retarget existing policies. Must not be set for the `INSTANCE` scope.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace the policy is attached to. Required for the `NAMESPACE` scope, must not be set otherwise.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The policy YAML source: `id`, optional `displayName`, `description`, `enforcement` (defaults to `ACTIVE`) and `target`, and the non-empty `rules` list mixing mutate rules (`io.kestra.plugin.ee.rules.Add`, `Delete`) and validate rules (`Deny`, `Require`, `Restrict`). The scope, tenant and namespace are carried by the resource attributes, never by the content.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					YamlEqualPlanModifier(),
				},
			},
		},
	}
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (r *policyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data policyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Content.IsNull() && !data.Content.IsUnknown() {
		_, contentId, err := parsePolicyContent(data.Content.ValueString())
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("content"), "Invalid content", err.Error())
		} else if !data.PolicyId.IsNull() && !data.PolicyId.IsUnknown() && contentId != data.PolicyId.ValueString() {
			resp.Diagnostics.AddAttributeError(
				path.Root("content"),
				"Invalid content",
				fmt.Sprintf("The content `id` (%s) must match policy_id (%s).", contentId, data.PolicyId.ValueString()),
			)
		}
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

// populatePolicyModel refreshes the content from an API policy payload; used by Read
// (drift detection, import) and the data source — never by Create/Update, where the
// content must stay exactly as planned. The API persists the authored source verbatim,
// so the comparison is a plain string equality; policies persisted before the source
// round-trip existed carry no source and fall back to a rendered document, compared with
// containment semantics so server-stamped defaults never show up as drift.
func populatePolicyModel(ctx context.Context, data *policyResourceModel, res map[string]interface{}, diags *diag.Diagnostics) {
	if source, ok := res["source"].(string); ok && source != "" {
		data.Content = types.StringValue(source)
		return
	}

	// legacy policy without a source: keep the configured content when the server model
	// contains it, otherwise render the model back to YAML
	if !data.Content.IsNull() && !data.Content.IsUnknown() {
		if document, _, err := parsePolicyContent(data.Content.ValueString()); err == nil && policyModelContains(res, document) {
			return
		}
	}
	rendered, err := policyModelToYaml(res)
	if err != nil {
		diags.AddError("Client Error", err.Error())
		return
	}
	data.Content = types.StringValue(rendered)
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, _, err := createPolicy(ctx, r.providerData.Client, plan.Scope.ValueString(), r.tenantId(plan), plan.Namespace.ValueString(), plan.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create policy, got error: %s", err))
		return
	}
	if created == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create policy: the API returned an empty response")
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created a policy resource, res: %+v", created))

	// the API persists the source verbatim, so the planned content is stored as-is; only
	// the computed tenant_id is resolved here
	r.resolveTenantId(&plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, status, err := readPolicy(ctx, r.providerData.Client, state.Scope.ValueString(), r.tenantId(state), state.Namespace.ValueString(), state.PolicyId.ValueString())
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read policy, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a policy resource, res: %+v", read))

	populatePolicyModel(ctx, &state, read, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan policyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, _, err := updatePolicy(ctx, r.providerData.Client, plan.Scope.ValueString(), r.tenantId(plan), plan.Namespace.ValueString(), plan.PolicyId.ValueString(), plan.Content.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update policy, got error: %s", err))
		return
	}
	if updated == nil {
		resp.Diagnostics.AddError("Client Error", "Unable to update policy: the API returned an empty response")
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated a policy resource, res: %+v", updated))

	// the API persists the source verbatim, so the planned content is stored as-is; only
	// the computed tenant_id is resolved here
	r.resolveTenantId(&plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := deletePolicy(ctx, r.providerData.Client, state.Scope.ValueString(), r.tenantId(state), state.Namespace.ValueString(), state.PolicyId.ValueString())
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete policy, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted a policy resource: %s %s", state.Scope.ValueString(), state.PolicyId.ValueString()))
}

// ImportState accepts scope-shaped ids:
//   - INSTANCE/policy_id
//   - TENANT/tenant_id/policy_id
//   - NAMESPACE/tenant_id/namespace/policy_id
func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	scope := ""
	if len(parts) > 0 {
		scope = parts[0]
	}

	invalidFormat := func() {
		resp.Diagnostics.AddError(
			"Invalid import id",
			fmt.Sprintf("Expected INSTANCE/policy_id, TENANT/tenant_id/policy_id or NAMESPACE/tenant_id/namespace/policy_id, got: %s", req.ID),
		)
	}

	switch scope {
	case policyScopeInstance:
		if len(parts) != 2 {
			invalidFormat()
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[1])...)
	case policyScopeTenant:
		if len(parts) != 3 {
			invalidFormat()
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[2])...)
	case policyScopeNamespace:
		if len(parts) != 4 {
			invalidFormat()
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), parts[2])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), parts[3])...)
	default:
		invalidFormat()
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope"), scope)...)
}
