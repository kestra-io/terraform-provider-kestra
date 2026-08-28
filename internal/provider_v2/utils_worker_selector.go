package provider_v2

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// The default worker selector has the same shape on tenants and namespaces, so
// both read it through these helpers.

func workerSelectorFromBody(in map[string]interface{}) workerSelector {
	out := workerSelector{Tags: types.SetNull(types.StringType), Match: types.StringNull(), Fallback: types.StringNull()}
	if raw, ok := in["tags"].([]interface{}); ok {
		vals := make([]attr.Value, 0, len(raw))
		for _, v := range raw {
			if t, ok := v.(string); ok {
				vals = append(vals, types.StringValue(t))
			}
		}
		if sv, d := basetypes.NewSetValue(types.StringType, vals); !d.HasError() {
			out.Tags = sv
		}
	}
	if match, ok := in["match"].(string); ok && match != "" {
		out.Match = types.StringValue(match)
	}
	if fb, ok := in["fallback"].(string); ok && fb != "" {
		out.Fallback = types.StringValue(fb)
	}
	return out
}

// workerSelectorFromV0 reads a selector out of state written by the SDK v2
// implementation, where the tag set is a plain JSON array.
func workerSelectorFromV0(in map[string]interface{}) workerSelector {
	return workerSelectorFromBody(in)
}

// Object types used by the data sources, which expose these nested structures as
// computed lists of objects to stay representable in protocol v5.

var workerSelectorObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"tags":     types.SetType{ElemType: types.StringType},
	"match":    types.StringType,
	"fallback": types.StringType,
}}

var isolationObjectType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"enabled":         types.BoolType,
	"denied_services": types.SetType{ElemType: types.StringType},
}}

func workerSelectorToList(body map[string]interface{}) types.List {
	raw, ok := body["defaultWorkerSelector"].(map[string]interface{})
	if !ok {
		return types.ListNull(workerSelectorObjectType)
	}
	sel := workerSelectorFromBody(raw)
	tags := sel.Tags
	if tags.IsNull() {
		tags = types.SetValueMust(types.StringType, []attr.Value{})
	}
	obj, diags := types.ObjectValue(workerSelectorObjectType.AttrTypes, map[string]attr.Value{
		"tags":     tags,
		"match":    sel.Match,
		"fallback": sel.Fallback,
	})
	if diags.HasError() {
		return types.ListNull(workerSelectorObjectType)
	}
	out, diags := types.ListValue(workerSelectorObjectType, []attr.Value{obj})
	if diags.HasError() {
		return types.ListNull(workerSelectorObjectType)
	}
	return out
}

func isolationToList(body map[string]interface{}, key string) types.List {
	raw, ok := body[key].(map[string]interface{})
	if !ok {
		return types.ListNull(isolationObjectType)
	}
	enabled := types.BoolNull()
	if v, ok := raw["enabled"].(bool); ok {
		enabled = types.BoolValue(v)
	}
	denied := types.SetValueMust(types.StringType, []attr.Value{})
	if ds, ok := raw["deniedServices"].([]interface{}); ok {
		vals := make([]string, 0, len(ds))
		for _, v := range ds {
			if s, ok := v.(string); ok {
				vals = append(vals, s)
			}
		}
		sort.Strings(vals)
		elems := make([]attr.Value, 0, len(vals))
		for _, v := range vals {
			elems = append(elems, types.StringValue(v))
		}
		denied = types.SetValueMust(types.StringType, elems)
	}
	obj, diags := types.ObjectValue(isolationObjectType.AttrTypes, map[string]attr.Value{
		"enabled":         enabled,
		"denied_services": denied,
	})
	if diags.HasError() {
		return types.ListNull(isolationObjectType)
	}
	out, diags := types.ListValue(isolationObjectType, []attr.Value{obj})
	if diags.HasError() {
		return types.ListNull(isolationObjectType)
	}
	return out
}
