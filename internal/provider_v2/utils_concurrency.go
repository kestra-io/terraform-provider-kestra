package provider_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Kestra 2.0 evaluates a concurrency limit and a set of quotas on both tenants
// and namespaces. Both carry the same shape on either resource, so the schema
// blocks and the API conversions live here once.

type concurrency struct {
	Limit    types.Int64  `tfsdk:"limit"`
	Behavior types.String `tfsdk:"behavior"`
}

type quota struct {
	Duration types.String `tfsdk:"duration"`
	Limit    types.Int64  `tfsdk:"limit"`
	Behavior types.String `tfsdk:"behavior"`
}

// The two behaviours are not the same enum: a concurrency limit can queue the
// execution, a quota can only fail or cancel it.
var (
	concurrencyBehaviors = []string{"QUEUE", "CANCEL", "FAIL"}
	quotaBehaviors       = []string{"FAIL", "CANCEL"}
)

func concurrencyNestedBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "The concurrency limit applied to the executions of every flow inside this scope and its descendants.",
		Validators:          []validator.List{listvalidator.SizeAtMost(1)},
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"limit": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "The maximum number of concurrent executions.",
					Validators:          []validator.Int64{int64validator.AtLeast(1)},
				},
				"behavior": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "What happens to an execution once the limit is reached.",
					Validators:          []validator.String{stringvalidator.OneOf(concurrencyBehaviors...)},
				},
			},
		},
	}
}

func quotasNestedBlock() schema.ListNestedBlock {
	return schema.ListNestedBlock{
		MarkdownDescription: "Quotas evaluated before an execution starts. Without any quota, executions run normally.",
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"duration": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "The sliding window the quota is counted over, as an ISO-8601 duration (for example `PT1H`).",
				},
				"limit": schema.Int64Attribute{
					Required:            true,
					MarkdownDescription: "The maximum number of executions allowed inside the window.",
					Validators:          []validator.Int64{int64validator.AtLeast(1)},
				},
				"behavior": schema.StringAttribute{
					Required:            true,
					MarkdownDescription: "What happens to an execution once the quota is exhausted.",
					Validators:          []validator.String{stringvalidator.OneOf(quotaBehaviors...)},
				},
			},
		},
	}
}

// concurrencyToBody renders the block for a request body, or nil when no block
// is configured so the caller can leave the key out. Writes replace the whole
// entity, so an omitted key clears any limit the server was holding -- which is
// what an absent block should mean.
func concurrencyToBody(list []concurrency) interface{} {
	if len(list) == 0 {
		return nil
	}
	return map[string]interface{}{
		"limit":    list[0].Limit.ValueInt64(),
		"behavior": list[0].Behavior.ValueString(),
	}
}

func quotasToBody(list []quota) interface{} {
	if len(list) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, len(list))
	for i, q := range list {
		out[i] = map[string]interface{}{
			"duration": q.Duration.ValueString(),
			"limit":    q.Limit.ValueInt64(),
			"behavior": q.Behavior.ValueString(),
		}
	}
	return out
}

// configuredConcurrency snapshots the planned blocks so a write can restore them
// over whatever the response carried.
//
// The provider supports Kestra 2.0.0-rc1, which is in the CI matrix and which
// accepts these fields but omits them from the write response. They are pure
// configuration, so the framework requires the post-write state to match the
// plan; letting the response clear them fails the apply outright.
//
// The restore also masks server-side normalisation: an instance that clamped a
// limit or rewrote a behavior would land the planned value in state rather than
// Terraform raising "Provider produced inconsistent result after apply". Read
// treats an absent or differing key as drift, so either case resurfaces as a
// diff on the next plan.
type configuredConcurrency struct {
	concurrency []concurrency
	quotas      []quota
}

func snapshotConcurrency(c []concurrency, q []quota) configuredConcurrency {
	return configuredConcurrency{concurrency: c, quotas: q}
}

func (s configuredConcurrency) restore(c *[]concurrency, q *[]quota) {
	*c, *q = s.concurrency, s.quotas
}

// concurrencyFromBody mirrors the defaultWorkerSelector handling: the API omits
// null fields, so an absent key means genuinely unset and has to surface as
// drift rather than leaving the prior value in place.
func concurrencyFromBody(body map[string]interface{}) []concurrency {
	raw, ok := body["concurrency"].(map[string]interface{})
	if !ok {
		return nil
	}
	one := concurrency{Limit: types.Int64Null(), Behavior: types.StringNull()}
	if limit, ok := numberFromBody(raw["limit"]); ok {
		one.Limit = types.Int64Value(limit)
	}
	if behavior, ok := raw["behavior"].(string); ok && behavior != "" {
		one.Behavior = types.StringValue(behavior)
	}
	// Both are Required in the schema, so an empty object would put nulls in state
	// and fail the apply with an opaque type error. Treat it as unset.
	if one.Limit.IsNull() && one.Behavior.IsNull() {
		return nil
	}
	return []concurrency{one}
}

func quotasFromBody(body map[string]interface{}) []quota {
	raw, ok := body["quotas"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]quota, 0, len(raw))
	for _, item := range raw {
		mp, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		one := quota{Duration: types.StringNull(), Limit: types.Int64Null(), Behavior: types.StringNull()}
		if duration, ok := mp["duration"].(string); ok && duration != "" {
			one.Duration = types.StringValue(duration)
		}
		if limit, ok := numberFromBody(mp["limit"]); ok {
			one.Limit = types.Int64Value(limit)
		}
		if behavior, ok := mp["behavior"].(string); ok && behavior != "" {
			one.Behavior = types.StringValue(behavior)
		}
		out = append(out, one)
	}
	return out
}

// numberFromBody accepts the shapes a JSON number can decode into, since the
// request bodies are round-tripped through map[string]interface{} rather than a
// typed model.
func numberFromBody(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	}
	return 0, false
}

// Data sources expose these as computed lists of objects rather than blocks: the
// mux server downgrades the framework provider to protocol v5, and a computed
// nested attribute is not representable there.

var concurrencyObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"limit":    types.Int64Type,
	"behavior": types.StringType,
}}

var quotaObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"duration": types.StringType,
	"limit":    types.Int64Type,
	"behavior": types.StringType,
}}

func concurrencyToList(body map[string]interface{}) types.List {
	list := concurrencyFromBody(body)
	elems := make([]attr.Value, 0, len(list))
	for _, c := range list {
		obj, diags := types.ObjectValue(concurrencyObjectType.AttrTypes, map[string]attr.Value{
			"limit":    c.Limit,
			"behavior": c.Behavior,
		})
		if diags.HasError() {
			continue
		}
		elems = append(elems, obj)
	}
	out, diags := types.ListValue(concurrencyObjectType, elems)
	if diags.HasError() {
		return types.ListNull(concurrencyObjectType)
	}
	return out
}

func quotasToList(body map[string]interface{}) types.List {
	list := quotasFromBody(body)
	elems := make([]attr.Value, 0, len(list))
	for _, q := range list {
		obj, diags := types.ObjectValue(quotaObjectType.AttrTypes, map[string]attr.Value{
			"duration": q.Duration,
			"limit":    q.Limit,
			"behavior": q.Behavior,
		})
		if diags.HasError() {
			continue
		}
		elems = append(elems, obj)
	}
	out, diags := types.ListValue(quotaObjectType, elems)
	if diags.HasError() {
		return types.ListNull(quotaObjectType)
	}
	return out
}
