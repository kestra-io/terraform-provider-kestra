package provider

import (
	"testing"
)

func TestMigratePermissions_SimpleRename(t *testing.T) {
	old := map[string]interface{}{
		"AI_COPILOT": []interface{}{"READ"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "COPILOT", []string{"USE"})
	if _, ok := result["AI_COPILOT"]; ok {
		t.Error("AI_COPILOT should have been renamed to COPILOT")
	}
}

func TestMigratePermissions_TestToTestsuite(t *testing.T) {
	old := map[string]interface{}{
		"TEST": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "TESTSUITE", []string{"VIEW", "LIST", "CREATE", "UPDATE", "DELETE"})
	if _, ok := result["TEST"]; ok {
		t.Error("TEST should have been renamed to TESTSUITE")
	}
}

func TestMigratePermissions_FlowExpansion(t *testing.T) {
	old := map[string]interface{}{
		"FLOW": []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "FLOW", []string{
		"VIEW", "LIST", "EXPORT",
		"UPDATE", "EXECUTE", "DISABLE", "ENABLE", "VALIDATE",
	})
}

func TestMigratePermissions_ExecutionFullExpansion(t *testing.T) {
	old := map[string]interface{}{
		"EXECUTION": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "EXECUTION", []string{
		"VIEW", "LIST", "ACCESS_LOGS", "ACCESS_OUTPUTS", "ACCESS_FILES", "EXPORT", "FOLLOW",
		"CREATE",
		"UPDATE", "RESTART", "KILL", "REPLAY", "PAUSE", "RESUME", "CHANGE_LABELS", "UNQUEUE", "FORCE_RUN",
		"DELETE",
	})
}

func TestMigratePermissions_SettingSplit(t *testing.T) {
	old := map[string]interface{}{
		"SETTING": []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "SYSTEM_SETTINGS", []string{"VIEW", "UPDATE"})
	assertActions(t, result, "TENANT_SETTINGS", []string{"VIEW", "UPDATE"})
	if _, ok := result["SETTING"]; ok {
		t.Error("SETTING should have been split into SYSTEM_SETTINGS and TENANT_SETTINGS")
	}
}

func TestMigratePermissions_SettingDropsCreateDelete(t *testing.T) {
	old := map[string]interface{}{
		"SETTING": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "SYSTEM_SETTINGS", []string{"VIEW", "UPDATE"})
	assertActions(t, result, "TENANT_SETTINGS", []string{"VIEW", "UPDATE"})
}

func TestMigratePermissions_ResourceMerge(t *testing.T) {
	old := map[string]interface{}{
		"APP":          []interface{}{"READ"},
		"APPEXECUTION": []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	actions := toStringSlice(result["APP"])
	if _, ok := result["APPEXECUTION"]; ok {
		t.Error("APPEXECUTION should have been merged into APP")
	}
	assertContains(t, actions, "VIEW")
	assertContains(t, actions, "LIST")
	assertContains(t, actions, "UPDATE")
	assertContains(t, actions, "EXECUTE")
	assertContains(t, actions, "ACCESS_FILES")
	assertContains(t, actions, "ACCESS_LOGS")
}

func TestMigratePermissions_DroppedResources(t *testing.T) {
	old := map[string]interface{}{
		"IMPERSONATE": []interface{}{"READ"},
		"UNKNOWN":     []interface{}{"READ"},
		"TEMPLATE":    []interface{}{"READ", "UPDATE"},
		"FLOW":        []interface{}{"READ"},
	}
	result := migratePermissions(old)

	if _, ok := result["IMPERSONATE"]; ok {
		t.Error("IMPERSONATE should be dropped")
	}
	if _, ok := result["UNKNOWN"]; ok {
		t.Error("UNKNOWN should be dropped")
	}
	if _, ok := result["TEMPLATE"]; ok {
		t.Error("TEMPLATE should be dropped")
	}
	assertActions(t, result, "FLOW", []string{"VIEW", "LIST", "EXPORT"})
}

func TestMigratePermissions_AlreadyMigrated(t *testing.T) {
	perms := map[string]interface{}{
		"FLOW": []interface{}{"VIEW", "LIST", "CREATE"},
	}
	if !isAlreadyMigrated(perms) {
		t.Error("permissions with VIEW should be detected as already migrated")
	}
}

func TestMigratePermissions_NotYetMigrated(t *testing.T) {
	perms := map[string]interface{}{
		"FLOW": []interface{}{"READ", "UPDATE"},
	}
	if isAlreadyMigrated(perms) {
		t.Error("permissions with only READ/UPDATE should not be detected as migrated")
	}
}

func TestMigratePermissions_EmptyInput(t *testing.T) {
	result := migratePermissions(nil)
	if result != nil {
		t.Error("nil input should return nil")
	}

	result = migratePermissions(map[string]interface{}{})
	if len(result) != 0 {
		t.Error("empty input should return empty")
	}
}

func TestMigratePermissions_IsAlreadyMigratedEmpty(t *testing.T) {
	if !isAlreadyMigrated(nil) {
		t.Error("nil should be considered migrated")
	}
	if !isAlreadyMigrated(map[string]interface{}{}) {
		t.Error("empty should be considered migrated")
	}
}

func TestMigratePermissions_TriggerExpansion(t *testing.T) {
	old := map[string]interface{}{
		"TRIGGER": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "TRIGGER", []string{
		"VIEW", "LIST", "EXPORT",
		"CREATE",
		"UNLOCK", "RESTART", "DISABLE", "ENABLE", "BACKFILL",
		"DELETE",
	})
}

func TestMigratePermissions_NamespaceExpansion(t *testing.T) {
	old := map[string]interface{}{
		"NAMESPACE": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "NAMESPACE", []string{
		"VIEW", "LIST", "EXPORT_PLUGIN_DEFAULTS",
		"CREATE",
		"UPDATE", "MANAGE_FILES", "IMPORT_PLUGIN_DEFAULTS",
		"DELETE",
	})
}

func TestMigratePermissions_AuditlogDropsCreateUpdateDelete(t *testing.T) {
	old := map[string]interface{}{
		"AUDITLOG": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "AUDITLOG", []string{"VIEW", "LIST", "EXPORT"})
}

func TestMigratePermissions_UserGroupExpansion(t *testing.T) {
	old := map[string]interface{}{
		"USER":  []interface{}{"READ", "UPDATE"},
		"GROUP": []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "USER", []string{"VIEW", "LIST", "UPDATE", "MANAGE_GROUP_MEMBERSHIP"})
	assertActions(t, result, "GROUP", []string{"VIEW", "LIST", "UPDATE", "MANAGE_MEMBERS"})
}

func TestMigratePermissions_GroupMembershipMerge(t *testing.T) {
	old := map[string]interface{}{
		"GROUP":            []interface{}{"READ", "UPDATE"},
		"GROUP_MEMBERSHIP": []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	if _, ok := result["GROUP_MEMBERSHIP"]; ok {
		t.Error("GROUP_MEMBERSHIP should be merged into GROUP")
	}
	actions := toStringSlice(result["GROUP"])
	assertContains(t, actions, "VIEW")
	assertContains(t, actions, "LIST")
	assertContains(t, actions, "UPDATE")
	assertContains(t, actions, "MANAGE_MEMBERS")
}

func TestMigratePermissions_DefaultCRUD(t *testing.T) {
	old := map[string]interface{}{
		"KVSTORE": []interface{}{"READ", "CREATE", "UPDATE", "DELETE"},
	}
	result := migratePermissions(old)

	assertActions(t, result, "KVSTORE", []string{"VIEW", "LIST", "CREATE", "UPDATE", "DELETE"})
}

func TestMigratePermissions_NoDuplicates(t *testing.T) {
	old := map[string]interface{}{
		"TENANT_ACCESS": []interface{}{"READ", "UPDATE"},
		"USER":          []interface{}{"READ", "UPDATE"},
	}
	result := migratePermissions(old)

	if _, ok := result["TENANT_ACCESS"]; ok {
		t.Error("TENANT_ACCESS should be merged into USER")
	}

	actions := toStringSlice(result["USER"])
	count := make(map[string]int)
	for _, a := range actions {
		count[a]++
		if count[a] > 1 {
			t.Errorf("duplicate action %q in USER", a)
		}
	}
}

// --- helpers ---

func toStringSlice(v interface{}) []string {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func assertActions(t *testing.T, result map[string]interface{}, resource string, expected []string) {
	t.Helper()
	v, ok := result[resource]
	if !ok {
		t.Errorf("expected resource %q in result", resource)
		return
	}
	actual := toStringSlice(v)
	if len(actual) != len(expected) {
		t.Errorf("%s: expected %d actions %v, got %d actions %v", resource, len(expected), expected, len(actual), actual)
		return
	}
	expectedSet := make(map[string]bool, len(expected))
	for _, e := range expected {
		expectedSet[e] = true
	}
	for _, a := range actual {
		if !expectedSet[a] {
			t.Errorf("%s: unexpected action %q (expected %v, got %v)", resource, a, expected, actual)
		}
	}
}

func assertContains(t *testing.T, actions []string, expected string) {
	t.Helper()
	for _, a := range actions {
		if a == expected {
			return
		}
	}
	t.Errorf("expected %q in actions %v", expected, actions)
}
