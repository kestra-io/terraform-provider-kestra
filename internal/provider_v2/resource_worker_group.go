package provider_v2

import (
	"context"
	"fmt"
	"net/http"

	int64validator "github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kestra-io/terraform-provider-kestra/internal/provider_v2/sdk_client"
)

var (
	_ resource.Resource                = &workerGroupResource{}
	_ resource.ResourceWithImportState = &workerGroupResource{}
	_ resource.ResourceWithConfigure   = &workerGroupResource{}
)

func NewWorkerGroupResource() resource.Resource {
	return &workerGroupResource{}
}

type workerGroupResource struct {
	providerData ProviderData
}

type workerGroupModel struct {
	Id            types.String        `tfsdk:"id"`
	GroupId       types.String        `tfsdk:"group_id"`
	Name          types.String        `tfsdk:"name"`
	Description   types.String        `tfsdk:"description"`
	Subscriptions []subscriptionModel `tfsdk:"subscriptions"`
}

type subscriptionModel struct {
	WorkerQueueId   types.String `tfsdk:"worker_queue_id"`
	ReservedPercent types.Int64  `tfsdk:"reserved_percent"`
	Mode            types.String `tfsdk:"mode"`
}

func (r *workerGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker_group"
}

// subscriptionsNestedBlockObject describes one worker group ↔ Worker Queue
// subscription edge. Modeled as a nested block (not a nested attribute)
// because the mux server downgrades the framework provider to protocol v5,
// which does not support nested attributes.
func subscriptionsNestedBlockObject() schema.NestedBlockObject {
	return schema.NestedBlockObject{
		Attributes: map[string]schema.Attribute{
			"worker_queue_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The id of the Worker Queue to subscribe to. Use the reserved `default` sentinel for the global default queue.",
			},
			"reserved_percent": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(-1),
				MarkdownDescription: "The reserved percentage of each worker's slots guaranteed to the Worker Queue: `-1` (no reservation, default) or a value in `[1, 100]`. The sum of reserved percentages across subscriptions must not exceed 100.",
				Validators: []validator.Int64{
					int64validator.Any(int64validator.OneOf(-1), int64validator.Between(1, 100)),
				},
			},
			"mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("STRICT"),
				MarkdownDescription: "The reservation interaction mode: `STRICT` (default, reserved slots are exclusive) or `ELASTIC` (idle reserved slots may be lent to other elastic subscriptions).",
				Validators:          []validator.String{stringvalidator.OneOf("STRICT", "ELASTIC")},
			},
		},
	}
}

func (r *workerGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Kestra Worker Group.\n\n-> This resource is only available on the [Enterprise Edition](https://kestra.io/enterprise)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The worker group id.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"group_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The worker group identifier (RFC 1123 label: lowercase alphanumerics and hyphens, must start and end with an alphanumeric, max 64 chars). Used in URLs and on the worker auth path; immutable.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The worker group display name. Defaults to the `group_id` when omitted.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The worker group description.",
			},
		},
		Blocks: map[string]schema.Block{
			"subscriptions": schema.ListNestedBlock{
				MarkdownDescription: "The Worker Queue subscriptions of the worker group. Subscriptions absent from the list are dropped on update; the underlying Worker Queue is preserved.",
				NestedObject:        subscriptionsNestedBlockObject(),
			},
		},
	}
}

func (r *workerGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// workerGroupPath returns the instance-level (SuperAdmin, non tenant-scoped)
// worker group endpoint path.
func workerGroupPath(id string) string {
	if id == "" {
		return "/api/v1/instance/worker-groups"
	}
	return "/api/v1/instance/worker-groups/" + id
}

func workerGroupModelToBody(m *workerGroupModel) map[string]interface{} {
	// The API requires a non-blank name; fall back to the group id.
	name := m.Name.ValueString()
	if m.Name.IsNull() || m.Name.IsUnknown() || name == "" {
		name = m.GroupId.ValueString()
	}

	subscriptions := make([]interface{}, 0, len(m.Subscriptions))
	for _, s := range m.Subscriptions {
		subscriptions = append(subscriptions, map[string]interface{}{
			"workerQueueId":   s.WorkerQueueId.ValueString(),
			"reservedPercent": s.ReservedPercent.ValueInt64(),
			"mode":            s.Mode.ValueString(),
		})
	}

	body := map[string]interface{}{
		"id":            m.GroupId.ValueString(),
		"name":          name,
		"subscriptions": subscriptions,
	}
	if !m.Description.IsNull() && !m.Description.IsUnknown() {
		body["description"] = m.Description.ValueString()
	}

	return body
}

func bodyToWorkerGroupModel(out map[string]interface{}, m *workerGroupModel) diag.Diagnostics {
	var diags diag.Diagnostics

	id, ok := out["id"].(string)
	if !ok || id == "" {
		diags.AddError("Invalid worker group API response", fmt.Sprintf("missing id in response: %v", out))
		return diags
	}
	m.Id = types.StringValue(id)
	m.GroupId = types.StringValue(id)

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

	// The API embeds the resolved queue in each subscription edge; only its id
	// is tracked on the resource. An empty list is kept as null when the prior
	// value was unset so an omitted attribute does not drift.
	rawSubscriptions, _ := out["subscriptions"].([]interface{})
	if len(rawSubscriptions) == 0 {
		if m.Subscriptions != nil {
			m.Subscriptions = []subscriptionModel{}
		}
		return diags
	}
	subscriptions := make([]subscriptionModel, 0, len(rawSubscriptions))
	for _, s := range rawSubscriptions {
		subscription, ok := s.(map[string]interface{})
		if !ok {
			continue
		}
		entry := subscriptionModel{
			WorkerQueueId:   types.StringNull(),
			ReservedPercent: types.Int64Value(-1),
			Mode:            types.StringValue("STRICT"),
		}
		if queue, ok := subscription["queue"].(map[string]interface{}); ok {
			if queueId, ok := queue["id"].(string); ok {
				entry.WorkerQueueId = types.StringValue(queueId)
			}
		}
		if reservedPercent, ok := subscription["reservedPercent"].(float64); ok {
			entry.ReservedPercent = types.Int64Value(int64(reservedPercent))
		}
		if mode, ok := subscription["mode"].(string); ok {
			entry.Mode = types.StringValue(mode)
		}
		subscriptions = append(subscriptions, entry)
	}
	m.Subscriptions = subscriptions

	return diags
}

func (r *workerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workerGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPost, workerGroupPath(""), workerGroupModelToBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Create worker group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerGroupModel(out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workerGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodGet, workerGroupPath(state.GroupId.ValueString()), nil)
	if err != nil {
		if status == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read worker group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerGroupModel(out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *workerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workerGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, _, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodPut, workerGroupPath(plan.GroupId.ValueString()), workerGroupModelToBody(&plan))
	if err != nil {
		resp.Diagnostics.AddError("Update worker group failed", err.Error())
		return
	}
	resp.Diagnostics.Append(bodyToWorkerGroupModel(out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workerGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, status, err := sdk_client.RawRequest(ctx, r.providerData.Client, http.MethodDelete, workerGroupPath(state.GroupId.ValueString()), nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Delete worker group failed", err.Error())
	}
}

func (r *workerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
