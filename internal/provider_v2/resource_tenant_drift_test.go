package provider_v2

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The API omits null fields and echoes back everything that is set, verified
// against a live 2.0 instance. So a key missing from a read means the value is
// genuinely unset, and every optional field has to clear rather than leave the
// prior value in state -- otherwise a setting cleared outside Terraform never
// shows up as drift.
func TestBodyToTenantModelClearsFieldsAbsentFromTheResponse(t *testing.T) {
	populated := func() *tenantModel {
		return &tenantModel{
			Id:                       types.StringValue("t1"),
			TenantId:                 types.StringValue("t1"),
			Name:                     types.StringValue("t1"),
			StorageType:              types.StringValue("s3"),
			StorageConfiguration:     types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringValue("v")}),
			SecretType:               types.StringValue("jdbc"),
			SecretReadOnly:           types.BoolValue(true),
			SecretConfiguration:      types.MapValueMust(types.StringType, map[string]attr.Value{"k": types.StringValue("v")}),
			RequireExistingNamespace: types.BoolValue(true),
			OutputsInInternalStorage: types.BoolValue(true),
		}
	}

	m := populated()
	// a response carrying only the identity: everything else was cleared server-side
	bodyToTenantModel(context.Background(), map[string]interface{}{"id": "t1", "name": "t1"}, m)

	if !m.StorageType.IsNull() {
		t.Errorf("storage_type should clear, got %v", m.StorageType)
	}
	if !m.StorageConfiguration.IsNull() {
		t.Errorf("storage_configuration should clear, got %v", m.StorageConfiguration)
	}
	if !m.SecretType.IsNull() {
		t.Errorf("secret_type should clear, got %v", m.SecretType)
	}
	if !m.SecretReadOnly.IsNull() {
		t.Errorf("secret_read_only should clear, got %v", m.SecretReadOnly)
	}
	if !m.SecretConfiguration.IsNull() {
		t.Errorf("secret_configuration should clear, got %v", m.SecretConfiguration)
	}
	if !m.RequireExistingNamespace.IsNull() {
		t.Errorf("require_existing_namespace should clear, got %v", m.RequireExistingNamespace)
	}
	if !m.OutputsInInternalStorage.IsNull() {
		t.Errorf("outputs_in_internal_storage should clear, got %v", m.OutputsInInternalStorage)
	}
}
