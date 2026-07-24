package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// currentWorkerGroupSchema returns the schema of the current implementation.
func currentWorkerGroupSchema(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	schemaResp := resource.SchemaResponse{}
	(&workerGroupResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}
	if schemaResp.Schema.Version != 1 {
		t.Fatalf("expected schema version 1, got %d", schemaResp.Schema.Version)
	}

	return tfsdk.State{Schema: schemaResp.Schema}
}

// TestWorkerGroupUpgradeStateV0 covers the migration of state written by the SDK
// v2 implementation, which held the worker group identifier in `key`.
func TestWorkerGroupUpgradeStateV0(t *testing.T) {
	allowedTenants := tftypes.List{ElementType: tftypes.String}

	tests := []struct {
		name            string
		key             tftypes.Value
		id              tftypes.Value
		allowedTenants  tftypes.Value
		expectedGroupId string
	}{
		{
			name:            "key carried over to group_id",
			key:             tftypes.NewValue(tftypes.String, "gpu-workers"),
			id:              tftypes.NewValue(tftypes.String, "gpu-workers"),
			allowedTenants:  tftypes.NewValue(allowedTenants, nil),
			expectedGroupId: "gpu-workers",
		},
		{
			name:            "id used when key is absent",
			key:             tftypes.NewValue(tftypes.String, nil),
			id:              tftypes.NewValue(tftypes.String, "gpu-workers"),
			allowedTenants:  tftypes.NewValue(allowedTenants, nil),
			expectedGroupId: "gpu-workers",
		},
		{
			// The shape published up to provider 1.3.2: `allowed_tenants` is set
			// and has to be dropped, `name` was not part of the schema yet.
			name:            "allowed_tenants dropped",
			key:             tftypes.NewValue(tftypes.String, "gpu-workers"),
			id:              tftypes.NewValue(tftypes.String, "gpu-workers"),
			allowedTenants:  tftypes.NewValue(allowedTenants, []tftypes.Value{tftypes.NewValue(tftypes.String, "main")}),
			expectedGroupId: "gpu-workers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			upgraders := (&workerGroupResource{}).UpgradeState(ctx)
			upgrader, ok := upgraders[0]
			if !ok {
				t.Fatal("expected a state upgrader for version 0")
			}
			if upgrader.PriorSchema == nil {
				t.Fatal("expected the version 0 upgrader to define a prior schema")
			}

			priorSchema := *upgrader.PriorSchema
			priorValue := tftypes.NewValue(
				priorSchema.Type().TerraformType(ctx).(tftypes.Object),
				map[string]tftypes.Value{
					"id":              tt.id,
					"key":             tt.key,
					"name":            tftypes.NewValue(tftypes.String, "GPU Workers"),
					"description":     tftypes.NewValue(tftypes.String, "worker group created before the rename"),
					"allowed_tenants": tt.allowedTenants,
				},
			)

			req := resource.UpgradeStateRequest{State: &tfsdk.State{Raw: priorValue, Schema: priorSchema}}
			resp := resource.UpgradeStateResponse{State: currentWorkerGroupSchema(ctx, t)}

			upgrader.StateUpgrader(ctx, req, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected upgrade diagnostics: %v", resp.Diagnostics)
			}

			var upgraded workerGroupModel
			if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
				t.Fatalf("unexpected state read diagnostics: %v", diags)
			}

			if upgraded.GroupId.ValueString() != tt.expectedGroupId {
				t.Errorf("expected group_id %q, got %q", tt.expectedGroupId, upgraded.GroupId.ValueString())
			}
			if upgraded.Id.ValueString() != tt.expectedGroupId {
				t.Errorf("expected id %q, got %q", tt.expectedGroupId, upgraded.Id.ValueString())
			}
			if upgraded.Name.ValueString() != "GPU Workers" {
				t.Errorf("expected name to be preserved, got %q", upgraded.Name.ValueString())
			}
			if upgraded.Description.ValueString() != "worker group created before the rename" {
				t.Errorf("expected description to be preserved, got %q", upgraded.Description.ValueString())
			}
			if upgraded.Subscriptions == nil || len(upgraded.Subscriptions) != 0 {
				t.Errorf("expected an empty subscription list, got %v", upgraded.Subscriptions)
			}
		})
	}
}
