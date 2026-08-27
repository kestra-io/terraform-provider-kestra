package provider_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Converters shared by the resources and data sources, which round-trip request
// bodies through map[string]interface{} rather than a typed model.

// stringMapFromBody renders a string map attribute, falling back to null when the
// API omitted the key.
func stringMapFromBody(raw interface{}) types.Map {
	in, ok := raw.(map[string]interface{})
	if !ok || len(in) == 0 {
		return types.MapNull(types.StringType)
	}
	els := map[string]attr.Value{}
	for k, v := range in {
		if s, ok := v.(string); ok {
			els[k] = types.StringValue(s)
		}
	}
	out, diags := types.MapValue(types.StringType, els)
	if diags.HasError() {
		return types.MapNull(types.StringType)
	}
	return out
}

// stringMapFromV0 reads a string map out of state written by the SDK v2
// implementation. The shape is the same as an API body, so it shares the decoder.
func stringMapFromV0(raw interface{}) types.Map {
	return stringMapFromBody(raw)
}

func dynamicFromBody(raw interface{}) types.Dynamic {
	in, ok := raw.(map[string]interface{})
	if !ok || len(in) == 0 {
		return types.DynamicNull()
	}
	dv, err := goValueToDynamic(in)
	if err != nil {
		return types.DynamicNull()
	}
	return dv
}
