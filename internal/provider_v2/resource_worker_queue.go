package provider_v2

import (
	"context"
	"fmt"
	"net/http"

	setvalidator "github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ resource.Resource                = &workerQueueResource{}
	_ resource.ResourceWithImportState = &workerQueueResource{}
	_ resource.ResourceWithConfigure   = &workerQueueResource{}
)

func NewWorkerQueueResource() resource.Resource {
	return &workerQueueResource{}
}

type workerQueueResource struct {
	providerData ProviderData
}

type workerQueueModel struct {
	Id             types.String `tfsdk:"id"`
	QueueId        types.String `tfsdk:"queue_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Tags           types.Set    `tfsdk:"tags"`
	AllowedTenants types.Set    `tfsdk:"allowed_tenants"`
}

func (r *workerQueueResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_queue"
}

func (r *workerQueueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Worker Queue. Worker Queues route tasks to Worker Groups through their tag set; Worker Groups subscribe to them via the `kestra_worker_group` resource.\n\n-> This resource is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The Worker Queue id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The Worker Queue identifier (RFC 1123 label: lowercase alphanumerics and hyphens, must start and end with an alphanumeric, max 64 chars). Used as the routing identity; immutable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Worker Queue human-readable name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The Worker Queue description.",
			},
			"tags": schema.SetAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The canonical tag set of the Worker Queue (each tag is an RFC 1123 label). Must not be empty.",
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
			"allowed_tenants": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "The tenants allowed to use the Worker Queue. Omit for an unrestricted queue.",
				Validators:          []validator.Set{setvalidator.SizeAtLeast(1)},
			},
		},
	}
}

func (r *workerQueueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// workerQueuePath returns the instance-level (SuperAdmin, non tenant-scoped)
// Worker Queue endpoint path.
func workerQueuePath(id string) string {
	if id == "" {
		return "/api/v1/instance/worker-queues"
	}
	return "/api/v1/instance/worker-queues/" + id
}

func workerQueueModelToBody(ctx context.Context, m *workerQueueModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	var tags []string
	diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)

	body := map[string]interface{}{
		"id":   m.QueueId.ValueString(),
		"tags": tags,
	}
	if !m.Name.IsNull() && !m.Name.IsUnknown() {
		body["name"] = m.Name.ValueString()
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		body["description"] = m.Description.ValueString()
	}
	if !m.AllowedTenants.IsNull() && !m.AllowedTenants.IsUnknown() {
		var allowedTenants []string
		diags.Append(m.AllowedTenants.ElementsAs(ctx, &allowedTenants, false)...)
		body["allowedTenants"] = allowedTenants
	}

	return body, diags
}

func bodyToWorkerQueueModel(ctx context.Context, out map[string]interface{}, m *workerQueueModel) diag.Diagnostics {
	var diags diag.Diagnostics

	id, ok := out["id"].(string)
	if !ok || id == "" {
		diags.AddError("Invalid Worker Queue API response", fmt.Sprintf("missing id in response: %v", out))
		return diags
	}
	m.Id = types.StringValue(id)
	m.QueueId = types.StringValue(id)

	if name, ok := out["name"].(string); ok && name != "" {
		m.Name = types.StringValue(name)
	} else {
		m.Name = types.StringNull()
	}

	if description, ok := out["description"].(string); ok && description != "" {
		m.Description = types.StringValue(description)
	} else {
		m.Description = types.StringNull()
	}

	tags := apiStringList(out["tags"])
	tagsValue, d := types.SetValueFrom(ctx, types.StringType, tags)
	diags.Append(d...)
	m.Tags = tagsValue

	// An empty tenant scope means unrestricted; keep it null so an omitted
	// attribute does not drift.
	allowedTenants := apiStringList(out["allowedTenants"])
	if len(allowedTenants) == 0 {
		m.AllowedTenants = types.SetNull(types.StringType)
	} else {
		allowedTenantsValue, d := types.SetValueFrom(ctx, types.StringType, allowedTenants)
		diags.Append(d...)
		m.AllowedTenants = allowedTenantsValue
	}

	return diags
}

// apiStringList converts a JSON-decoded array into a string slice, ignoring
// non-string elements.
func apiStringList(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func (r *workerQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workerQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := workerQueueModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPost, workerQueuePath(""), body)
	if err != nil {
		resp.Diagnostics.AddError("Create Worker Queue failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerQueueModel(ctx, out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workerQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workerQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodGet, workerQueuePath(state.QueueId.ValueString()), nil)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Worker Queue failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerQueueModel(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workerQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workerQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := workerQueueModelToBody(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPut, workerQueuePath(plan.QueueId.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Update Worker Queue failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerQueueModel(ctx, out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workerQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workerQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodDelete, workerQueuePath(state.QueueId.ValueString()), nil)
	if err != nil && status != http.StatusNotFound {
		if status == http.StatusConflict {
			resp.Diagnostics.AddError(
				"Delete Worker Queue failed",
				fmt.Sprintf("One or more worker groups still subscribe to Worker Queue %q; remove the subscription(s) first: %s", state.QueueId.ValueString(), err.Error()),
			)
			return
		}
		resp.Diagnostics.AddError("Delete Worker Queue failed", err.Error())
	}
}

func (r *workerQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("queue_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
