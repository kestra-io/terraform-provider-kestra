package provider_v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kestra-io/client-sdk/go-sdk/v2/kestra_api_client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
var _ resource.Resource = &reusableInputsResource{}
var _ resource.ResourceWithImportState = &reusableInputsResource{}

func NewReusableInputsResource() resource.Resource {
	return &reusableInputsResource{}
}

type reusableInputsResource struct {
	providerData ProviderData
}

// reusableInputsModel describes the resource and data source data model.
type reusableInputsModel struct {
	ReusableInputsId types.String `tfsdk:"reusable_inputs_id"`
	TenantId         types.String `tfsdk:"tenant_id"`
	Namespace        types.String `tfsdk:"namespace"`
	Content          types.String `tfsdk:"content"`
	Revision         types.Int64  `tfsdk:"revision"`
}

func (r *reusableInputsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reusable_inputs"
}

// resolveReusableInputsTenantId returns the configured tenant_id, falling back to the provider tenant.
func resolveReusableInputsTenantId(providerData ProviderData, data reusableInputsModel) string {
	if !data.TenantId.IsNull() && !data.TenantId.IsUnknown() && data.TenantId.ValueString() != "" {
		return data.TenantId.ValueString()
	}
	return providerData.TenantId
}

func (r *reusableInputsResource) tenantId(data reusableInputsModel) string {
	return resolveReusableInputsTenantId(r.providerData, data)
}

func reusableInputsSchema() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"reusable_inputs_id": schema.StringAttribute{
			MarkdownDescription: "The reusable inputs block id — unique per (tenant, namespace).",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"tenant_id": schema.StringAttribute{
			MarkdownDescription: "The tenant id (EE). Defaults to the provider tenant when omitted; the value is captured at create time, so changing the provider tenant later does not retarget the block.",
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
			MarkdownDescription: "The namespace the block is attached to. A block defined in a parent namespace is visible to (and overridable by) child namespaces.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"content": schema.StringAttribute{
			MarkdownDescription: "The block YAML source: an optional `description` and the non-empty `inputs` list (any input type except another `REUSABLE_INPUTS`, which cannot be nested). " +
				"Diffs are compared semantically, so a change that only reindents, reorders keys or edits comments produces no plan and the source persisted by the API keeps its previous formatting; change a value to push a reformatted source.",
			Required: true,
			PlanModifiers: []planmodifier.String{
				YamlEqualPlanModifier(),
			},
		},
		"revision": schema.Int64Attribute{
			MarkdownDescription: "The block revision, bumped by the API on every save.",
			Computed:            true,
		},
	}
}

func (r *reusableInputsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Reusable Inputs block (EE): a named, namespace-scoped set of input " +
			"definitions that flows reference via a `REUSABLE_INPUTS` input. Its YAML source is persisted by the " +
			"API and round-tripped verbatim. Requires Kestra EE 2.0.1 or later.",
		Attributes: reusableInputsSchema(),
	}
}

func (r *reusableInputsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// populateReusableInputsModel refreshes content and revision from an API payload; source is
// optional on the wire, and a payload without one leaves the existing content untouched.
func populateReusableInputsModel(data *reusableInputsModel, res *kestra_api_client.ReusableInputsWithSource) {
	if res.Source != nil {
		data.Content = types.StringValue(*res.Source)
	}
	data.Revision = types.Int64Value(int64(res.Revision))
}

func isReusableInputsNotFound(err error) bool {
	var apiErr *kestra_api_client.ApiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func (r *reusableInputsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reusableInputsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	failIfExists := true
	created, err := r.providerData.KestraClient.ReusableInputs().CreateOrUpdateReusableInputs(ctx, plan.Namespace.ValueString(), plan.ReusableInputsId.ValueString(), r.tenantId(plan), plan.Content.ValueString(), &failIfExists)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create reusable inputs, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("created a reusable_inputs resource, res: %+v", created))

	// the API persists the source verbatim, so the planned content is stored as-is
	plan.TenantId = types.StringValue(r.tenantId(plan))
	plan.Revision = types.Int64Value(int64(created.Revision))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reusableInputsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reusableInputsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	read, err := r.providerData.KestraClient.ReusableInputs().ReusableInputs(ctx, state.Namespace.ValueString(), state.ReusableInputsId.ValueString(), r.tenantId(state), nil)
	if err != nil {
		if isReusableInputsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read reusable inputs, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("read a reusable_inputs resource, res: %+v", read))

	if read.Namespace == "" {
		resp.Diagnostics.AddError("Client Error", "Unable to read reusable inputs: the API returned a response without a namespace")
		return
	}
	// the GET resolves namespace inheritance, so another namespace answering means this block does not exist
	if read.Namespace != state.Namespace.ValueString() {
		resp.State.RemoveResource(ctx)
		return
	}

	populateReusableInputsModel(&state, read)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *reusableInputsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reusableInputsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.providerData.KestraClient.ReusableInputs().CreateOrUpdateReusableInputs(ctx, plan.Namespace.ValueString(), plan.ReusableInputsId.ValueString(), r.tenantId(plan), plan.Content.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update reusable inputs, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("updated a reusable_inputs resource, res: %+v", updated))

	plan.TenantId = types.StringValue(r.tenantId(plan))
	plan.Revision = types.Int64Value(int64(updated.Revision))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reusableInputsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reusableInputsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.providerData.KestraClient.ReusableInputs().DeleteReusableInputs(ctx, state.Namespace.ValueString(), state.ReusableInputsId.ValueString(), r.tenantId(state))
	if err != nil && !isReusableInputsNotFound(err) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete reusable inputs, got error: %s", err))
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted a reusable_inputs resource: %s/%s", state.Namespace.ValueString(), state.ReusableInputsId.ValueString()))
}

// reusableInputsImportId is the parsed form of an import id: tenant_id/namespace/reusable_inputs_id.
type reusableInputsImportId struct {
	TenantId         string
	Namespace        string
	ReusableInputsId string
}

// parseReusableInputsImportId parses the import id tenant_id/namespace/id. An empty segment is
// rejected because it would silently build a request against a malformed path.
func parseReusableInputsImportId(id string) (reusableInputsImportId, error) {
	parts := strings.Split(id, "/")
	invalidFormat := fmt.Errorf("Expected tenant_id/namespace/reusable_inputs_id, got: %s", id)

	if len(parts) != 3 {
		return reusableInputsImportId{}, invalidFormat
	}
	for _, part := range parts {
		if part == "" {
			return reusableInputsImportId{}, invalidFormat
		}
	}

	return reusableInputsImportId{TenantId: parts[0], Namespace: parts[1], ReusableInputsId: parts[2]}, nil
}

// ImportState accepts the id tenant_id/namespace/reusable_inputs_id.
func (r *reusableInputsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parsed, err := parseReusableInputsImportId(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import id", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tenant_id"), parsed.TenantId)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), parsed.Namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("reusable_inputs_id"), parsed.ReusableInputsId)...)
}
