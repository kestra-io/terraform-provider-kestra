package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runYamlEqualModifier(t *testing.T, state, plan string) types.String {
	t.Helper()
	req := planmodifier.StringRequest{
		StateValue: types.StringValue(state),
		PlanValue:  types.StringValue(plan),
	}
	resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}
	YamlEqualPlanModifier().PlanModifyString(context.Background(), req, resp)
	return resp.PlanValue
}

func TestUnitYamlEqualModifierSuppressesFormattingOnlyChanges(t *testing.T) {
	state := "- type: io.kestra.plugin.ee.rules.Deny\n  on: PLUGIN\n"
	plan := "# a comment\n- on: PLUGIN\n  type: \"io.kestra.plugin.ee.rules.Deny\"\n"
	if got := runYamlEqualModifier(t, state, plan); got.ValueString() != state {
		t.Errorf("expected the state value to be kept for a formatting-only change, got: %q", got.ValueString())
	}
}

func TestUnitYamlEqualModifierKeepsRealChanges(t *testing.T) {
	tests := []struct {
		name  string
		state string
		plan  string
	}{
		{"different value", "value: a\n", "value: b\n"},
		// YAML 1.2 semantics: `on` and `true` are different strings/values, while YAML 1.1
		// would resolve both to boolean true and wrongly suppress the update
		{"yaml 1.1 boolean lookalikes", "value: on\n", "value: true\n"},
		{"yaml 1.1 sexagesimal lookalikes", "value: 12:30\n", "value: 750\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runYamlEqualModifier(t, tt.state, tt.plan); got.ValueString() != tt.plan {
				t.Errorf("expected the plan value to be kept for a real change, got: %q", got.ValueString())
			}
		})
	}
}
