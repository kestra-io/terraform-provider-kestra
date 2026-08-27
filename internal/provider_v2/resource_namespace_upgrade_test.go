package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// currentNamespaceState returns an empty state carrying the schema of the current
// implementation.
func currentNamespaceState(ctx context.Context, t *testing.T) tfsdk.State {
	t.Helper()

	schemaResp := resource.SchemaResponse{}
	(&namespaceResource{}).Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("unexpected schema diagnostics: %v", schemaResp.Diagnostics)
	}
	if schemaResp.Schema.Version != 2 {
		t.Fatalf("expected schema version 2, got %d", schemaResp.Schema.Version)
	}

	return tfsdk.State{Schema: schemaResp.Schema}
}

// TestNamespaceUpgradeStateV1 covers the migration of state written before Kestra 2.0
// removed `plugin_defaults` and replaced `worker_group` with `default_worker_selector`.
// Neither can be translated, so both are dropped and everything else survives.
func TestNamespaceUpgradeStateV1(t *testing.T) {
	ctx := context.Background()

	upgraders := (&namespaceResource{}).UpgradeState(ctx)
	upgrader, ok := upgraders[1]
	if !ok {
		t.Fatal("expected a state upgrader for version 1")
	}
	if upgrader.PriorSchema == nil {
		t.Fatal("expected the version 1 upgrader to define a prior schema")
	}

	priorSchema := *upgrader.PriorSchema
	priorType := priorSchema.Type().TerraformType(ctx).(tftypes.Object)

	allowedNamespaceType := priorType.AttributeTypes["allowed_namespaces"].(tftypes.List).ElementType
	workerGroupType := priorType.AttributeTypes["worker_group"].(tftypes.List).ElementType
	isolationType := priorType.AttributeTypes["storage_isolation"].(tftypes.List).ElementType

	priorValue := tftypes.NewValue(priorType, map[string]tftypes.Value{
		"id":              tftypes.NewValue(tftypes.String, "io.kestra.terraform"),
		"tenant_id":       tftypes.NewValue(tftypes.String, "main"),
		"namespace_id":    tftypes.NewValue(tftypes.String, "io.kestra.terraform"),
		"description":     tftypes.NewValue(tftypes.String, "namespace written before 2.0"),
		"variables":       tftypes.NewValue(tftypes.String, "k1: 1\n"),
		"plugin_defaults": tftypes.NewValue(tftypes.String, "- type: io.kestra.plugin.core.log.Log\n"),
		"storage_type":    tftypes.NewValue(tftypes.String, "s3"),
		"storage_configuration": tftypes.NewValue(
			tftypes.Map{ElementType: tftypes.String},
			map[string]tftypes.Value{"bucket": tftypes.NewValue(tftypes.String, "my-bucket")},
		),
		"secret_type":                 tftypes.NewValue(tftypes.String, "vault"),
		"secret_read_only":            tftypes.NewValue(tftypes.Bool, true),
		"secret_configuration":        tftypes.NewValue(tftypes.DynamicPseudoType, nil),
		"outputs_in_internal_storage": tftypes.NewValue(tftypes.Bool, false),
		"allowed_namespaces": tftypes.NewValue(
			tftypes.List{ElementType: allowedNamespaceType},
			[]tftypes.Value{tftypes.NewValue(allowedNamespaceType, map[string]tftypes.Value{
				"namespace": tftypes.NewValue(tftypes.String, "io.kestra.allowed"),
			})},
		),
		"worker_group": tftypes.NewValue(
			tftypes.List{ElementType: workerGroupType},
			[]tftypes.Value{tftypes.NewValue(workerGroupType, map[string]tftypes.Value{
				"key":      tftypes.NewValue(tftypes.String, "my-worker-group"),
				"fallback": tftypes.NewValue(tftypes.String, "WAIT"),
			})},
		),
		"storage_isolation": tftypes.NewValue(
			tftypes.List{ElementType: isolationType},
			[]tftypes.Value{tftypes.NewValue(isolationType, map[string]tftypes.Value{
				"enabled":         tftypes.NewValue(tftypes.Bool, true),
				"denied_services": tftypes.NewValue(tftypes.Set{ElementType: tftypes.String}, nil),
			})},
		),
		"secret_isolation": tftypes.NewValue(tftypes.List{ElementType: isolationType}, nil),
	})

	req := resource.UpgradeStateRequest{State: &tfsdk.State{Raw: priorValue, Schema: priorSchema}}
	resp := resource.UpgradeStateResponse{State: currentNamespaceState(ctx, t)}

	upgrader.StateUpgrader(ctx, req, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected upgrade diagnostics: %v", resp.Diagnostics)
	}

	var upgraded namespaceModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("unexpected state read diagnostics: %v", diags)
	}

	// `worker_group` carried a worker group key, which has no equivalent in a tag-based
	// selector; the refresh that follows the upgrade repopulates it from the API.
	if upgraded.DefaultWorkerSelector != nil {
		t.Errorf("expected default_worker_selector to be empty after the upgrade, got %v", upgraded.DefaultWorkerSelector)
	}

	if upgraded.NamespaceId.ValueString() != "io.kestra.terraform" {
		t.Errorf("expected namespace_id to be preserved, got %q", upgraded.NamespaceId.ValueString())
	}
	if upgraded.TenantId.ValueString() != "main" {
		t.Errorf("expected tenant_id to be preserved, got %q", upgraded.TenantId.ValueString())
	}
	if upgraded.Description.ValueString() != "namespace written before 2.0" {
		t.Errorf("expected description to be preserved, got %q", upgraded.Description.ValueString())
	}
	if upgraded.Variables.ValueString() != "k1: 1\n" {
		t.Errorf("expected variables to be preserved, got %q", upgraded.Variables.ValueString())
	}
	if upgraded.StorageType.ValueString() != "s3" {
		t.Errorf("expected storage_type to be preserved, got %q", upgraded.StorageType.ValueString())
	}
	if upgraded.SecretType.ValueString() != "vault" {
		t.Errorf("expected secret_type to be preserved, got %q", upgraded.SecretType.ValueString())
	}
	if !upgraded.SecretReadOnly.ValueBool() {
		t.Error("expected secret_read_only to be preserved")
	}
	if len(upgraded.AllowedNamespaces) != 1 || upgraded.AllowedNamespaces[0].Namespace.ValueString() != "io.kestra.allowed" {
		t.Errorf("expected allowed_namespaces to be preserved, got %v", upgraded.AllowedNamespaces)
	}
	if len(upgraded.StorageIsolation) != 1 || !upgraded.StorageIsolation[0].Enabled.ValueBool() {
		t.Errorf("expected storage_isolation to be preserved, got %v", upgraded.StorageIsolation)
	}
	if elements := upgraded.StorageConfiguration.Elements(); len(elements) != 1 {
		t.Errorf("expected storage_configuration to be preserved, got %v", elements)
	}
}
