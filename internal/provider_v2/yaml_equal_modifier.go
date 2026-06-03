package provider_v2

import (
	"context"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"gopkg.in/yaml.v2"
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
	var a, b interface{}
	if err := yaml.Unmarshal([]byte(req.StateValue.ValueString()), &a); err != nil {
		return
	}
	if err := yaml.Unmarshal([]byte(req.PlanValue.ValueString()), &b); err != nil {
		return
	}
	if reflect.DeepEqual(a, b) {
		resp.PlanValue = req.StateValue
	}
}
