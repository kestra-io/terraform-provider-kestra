package provider_v2

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	// yaml.v3 follows YAML 1.2, matching what providers send to the API: unquoted scalars
	// and keys like `on`, `yes` or `12:30` stay strings instead of resolving to YAML 1.1
	// booleans/sexagesimals, so a real change between them is never suppressed.
	yaml "gopkg.in/yaml.v3"
)

type yamlEqualModifier struct{}

func YamlEqualPlanModifier() planmodifier.String {
	return &yamlEqualModifier{}
}

func (m *yamlEqualModifier) Description(_ context.Context) string {
	return "Suppress plan diff when state and planned YAML are semantically equal."
}

func (m *yamlEqualModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m *yamlEqualModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	var state, plan interface{}
	if err := yaml.Unmarshal([]byte(req.StateValue.ValueString()), &state); err != nil {
		return
	}
	if err := yaml.Unmarshal([]byte(req.PlanValue.ValueString()), &plan); err != nil {
		return
	}
	// Both sides come out of the same yaml.v3 decoder, so a YAML `3` can never meet a JSON
	// `3.0` here and the JSON normalization used against API payloads would only cost
	// precision: it decodes every number as float64, collapsing integers above 2^53 and
	// silently dropping a real change from the plan.
	if reflect.DeepEqual(yamlToJSONCompatible(state), yamlToJSONCompatible(plan)) {
		resp.PlanValue = req.StateValue
	}
}
