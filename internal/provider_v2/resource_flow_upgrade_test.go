package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// TestFlowUpgradeStateV0 covers state written by the SDK v2 implementation. The attribute
// names did not change, so everything has to survive the move.
func TestFlowUpgradeStateV0(t *testing.T) {
	ctx := context.Background()

	schemaResp := resource.SchemaResponse{}
	(&flowResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Schema.Version != 1 {
		t.Fatalf("expected schema version 1, got %d", schemaResp.Schema.Version)
	}

	upgrader, ok := (&flowResource{}).UpgradeState(ctx)[0]
	if !ok {
		t.Fatal("expected a state upgrader for version 0")
	}

	const priorJSON = `{
		"id": "company.team/hello",
		"tenant_id": "main",
		"namespace": "company.team",
		"flow_id": "hello",
		"revision": 4,
		"content": "id: hello\nnamespace: company.team\ndisabled: true\ntasks: []\n"
	}`

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: []byte(priorJSON)}}
	resp := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	upgrader.StateUpgrader(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected upgrade diagnostics: %v", resp.Diagnostics)
	}

	var upgraded flowModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("unexpected state read diagnostics: %v", diags)
	}

	if got := upgraded.Id.ValueString(); got != "company.team/hello" {
		t.Errorf("id = %q, want company.team/hello", got)
	}
	if got := upgraded.TenantId.ValueString(); got != "main" {
		t.Errorf("tenant_id = %q, want main", got)
	}
	if got := upgraded.Namespace.ValueString(); got != "company.team" {
		t.Errorf("namespace = %q, want company.team", got)
	}
	if got := upgraded.FlowId.ValueString(); got != "hello" {
		t.Errorf("flow_id = %q, want hello", got)
	}
	if got := upgraded.Revision.ValueInt64(); got != 4 {
		t.Errorf("revision = %d, want 4", got)
	}
	if got := upgraded.Content.ValueString(); got != "id: hello\nnamespace: company.team\ndisabled: true\ntasks: []\n" {
		t.Errorf("content = %q", got)
	}
}
