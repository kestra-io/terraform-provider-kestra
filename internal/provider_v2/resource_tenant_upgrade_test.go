package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// currentTenantState returns an empty state carrying the schema of the current
// implementation.
func currentTenantState(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	schemaResp := resource.SchemaResponse{}
	(&tenantResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}
	if schemaResp.Schema.Version != 1 {
		t.Fatalf("expected schema version 1, got %d", schemaResp.Schema.Version)
	}

	return tfsdk.State{Schema: schemaResp.Schema}
}

// TestTenantUpgradeStateV0 covers state written by the SDK v2 implementation. The
// attribute names did not change, so everything has to survive the move.
func TestTenantUpgradeStateV0(t *testing.T) {
	ctx := context.Background()

	upgraders := (&tenantResource{}).UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok {
		t.Fatal("expected a state upgrader for version 0")
	}

	const priorJSON = `{
		"id": "custom",
		"tenant_id": "custom",
		"name": "My custom tenant",
		"storage_type": "s3",
		"storage_configuration": {"bucket": "kestra"},
		"secret_type": "vault",
		"secret_read_only": true,
		"require_existing_namespace": true,
		"outputs_in_internal_storage": true,
		"default_worker_selector": [{"tags": ["gpu"], "match": "ANY", "fallback": "WAIT"}],
		"storage_isolation": [{"enabled": true, "denied_services": ["WORKER"]}],
		"secret_isolation": [{"enabled": false, "denied_services": []}]
	}`

	req := resource.UpgradeStateRequest{RawState: &tfprotov6.RawState{JSON: []byte(priorJSON)}}
	resp := resource.UpgradeStateResponse{State: currentTenantState(ctx, t)}

	upgrader.StateUpgrader(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected upgrade diagnostics: %v", resp.Diagnostics)
	}

	var upgraded tenantModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("unexpected state read diagnostics: %v", diags)
	}

	if got := upgraded.TenantId.ValueString(); got != "custom" {
		t.Errorf("tenant_id = %q, want custom", got)
	}
	if got := upgraded.Name.ValueString(); got != "My custom tenant" {
		t.Errorf("name = %q, want %q", got, "My custom tenant")
	}
	if got := upgraded.StorageType.ValueString(); got != "s3" {
		t.Errorf("storage_type = %q, want s3", got)
	}
	if !upgraded.SecretReadOnly.ValueBool() {
		t.Error("secret_read_only should have carried over as true")
	}
	if !upgraded.RequireExistingNamespace.ValueBool() {
		t.Error("require_existing_namespace should have carried over as true")
	}
	if !upgraded.OutputsInInternalStorage.ValueBool() {
		t.Error("outputs_in_internal_storage should have carried over as true")
	}

	if upgraded.StorageConfiguration.IsNull() {
		t.Error("storage_configuration should have carried over")
	}

	if len(upgraded.DefaultWorkerSelector) != 1 {
		t.Fatalf("expected one default_worker_selector, got %d", len(upgraded.DefaultWorkerSelector))
	}
	selector := upgraded.DefaultWorkerSelector[0]
	if got := selector.Match.ValueString(); got != "ANY" {
		t.Errorf("default_worker_selector.match = %q, want ANY", got)
	}
	if got := selector.Fallback.ValueString(); got != "WAIT" {
		t.Errorf("default_worker_selector.fallback = %q, want WAIT", got)
	}
	if selector.Tags.IsNull() {
		t.Error("default_worker_selector.tags should have carried over")
	}

	if len(upgraded.StorageIsolation) != 1 || !upgraded.StorageIsolation[0].Enabled.ValueBool() {
		t.Errorf("storage_isolation should have carried over enabled=true, got %#v", upgraded.StorageIsolation)
	}

	// The blocks the SDK v2 schema never had must not be conjured up by the upgrade.
}
