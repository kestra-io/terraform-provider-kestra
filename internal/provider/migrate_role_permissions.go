package provider

// Resource renames: old permission name → list of new resource names.
// Based on Java PermissionMigrationHelper.RESOURCE_RENAME + Resource.fromString() aliases.
var resourceRenames = map[string][]string{
	"SETTING":          {"SYSTEM_SETTINGS", "TENANT_SETTINGS"},
	"AI_COPILOT":       {"COPILOT"},
	"TENANT_ACCESS":    {"USER"},
	"APPEXECUTION":     {"APP"},
	"GROUP_MEMBERSHIP": {"GROUP"},
	"TEST":             {"TESTSUITE"},
}

var droppedResources = map[string]bool{
	"IMPERSONATE": true,
	"UNKNOWN":     true,
	"TEMPLATE":    true,
}

// Actions that only exist in the new model, used to detect already-migrated permissions.
var newModelActions = map[string]bool{
	"VIEW": true, "LIST": true, "EXECUTE": true, "FOLLOW": true,
	"ACCESS_LOGS": true, "ACCESS_OUTPUTS": true, "ACCESS_FILES": true,
	"MANAGE_FILES": true, "USE": true, "MANAGE_MEMBERS": true,
	"MANAGE_GROUP_MEMBERSHIP": true, "EXPORT_PLUGIN_DEFAULTS": true,
	"IMPORT_PLUGIN_DEFAULTS": true, "BACKFILL": true, "UNLOCK": true,
}

func isAlreadyMigrated(permissions map[string]interface{}) bool {
	if len(permissions) == 0 {
		return true
	}
	for _, v := range permissions {
		actions, ok := v.([]interface{})
		if !ok {
			continue
		}
		for _, a := range actions {
			if s, ok := a.(string); ok && newModelActions[s] {
				return true
			}
		}
	}
	return false
}

func migratePermissions(old map[string]interface{}) map[string]interface{} {
	if len(old) == 0 {
		return old
	}

	result := make(map[string][]string)

	for oldResource, v := range old {
		if droppedResources[oldResource] {
			continue
		}

		oldActions, ok := v.([]interface{})
		if !ok {
			continue
		}

		targets, renamed := resourceRenames[oldResource]
		if !renamed {
			targets = []string{oldResource}
		}

		for _, target := range targets {
			for _, a := range oldActions {
				action, ok := a.(string)
				if !ok {
					continue
				}
				expanded := expandAction(target, action)
				result[target] = appendUnique(result[target], expanded...)
			}
		}
	}

	out := make(map[string]interface{}, len(result))
	for k, v := range result {
		iface := make([]interface{}, len(v))
		for i, s := range v {
			iface[i] = s
		}
		out[k] = iface
	}
	return out
}

func expandAction(resource, oldAction string) []string {
	switch oldAction {
	case "READ":
		return expandRead(resource)
	case "CREATE":
		return expandCreate(resource)
	case "UPDATE":
		return expandUpdate(resource)
	case "DELETE":
		return expandDelete(resource)
	default:
		return nil
	}
}

func expandRead(resource string) []string {
	switch resource {
	case "FLOW":
		return []string{"VIEW", "LIST", "EXPORT"}
	case "EXECUTION":
		return []string{"VIEW", "LIST", "ACCESS_LOGS", "ACCESS_OUTPUTS", "ACCESS_FILES", "EXPORT", "FOLLOW"}
	case "TRIGGER":
		return []string{"VIEW", "LIST", "EXPORT"}
	case "NAMESPACE":
		return []string{"VIEW", "LIST", "EXPORT_PLUGIN_DEFAULTS"}
	case "AUDITLOG":
		return []string{"VIEW", "LIST", "EXPORT"}
	case "COPILOT":
		return []string{"USE"}
	case "SYSTEM_SETTINGS", "TENANT_SETTINGS":
		return []string{"VIEW"}
	default:
		return []string{"VIEW", "LIST"}
	}
}

func expandCreate(resource string) []string {
	switch resource {
	case "FLOW":
		return []string{"CREATE", "IMPORT"}
	case "AUDITLOG", "SYSTEM_SETTINGS", "TENANT_SETTINGS", "COPILOT":
		return nil
	default:
		return []string{"CREATE"}
	}
}

func expandUpdate(resource string) []string {
	switch resource {
	case "FLOW":
		return []string{"UPDATE", "EXECUTE", "DISABLE", "ENABLE", "VALIDATE"}
	case "EXECUTION":
		return []string{"UPDATE", "RESTART", "KILL", "REPLAY", "PAUSE", "RESUME", "CHANGE_LABELS", "UNQUEUE", "FORCE_RUN"}
	case "TRIGGER":
		return []string{"UNLOCK", "RESTART", "DISABLE", "ENABLE", "BACKFILL"}
	case "NAMESPACE":
		return []string{"UPDATE", "MANAGE_FILES", "IMPORT_PLUGIN_DEFAULTS"}
	case "APP":
		return []string{"UPDATE", "EXECUTE", "ACCESS_FILES", "ACCESS_LOGS"}
	case "USER":
		return []string{"UPDATE", "MANAGE_GROUP_MEMBERSHIP"}
	case "GROUP":
		return []string{"UPDATE", "MANAGE_MEMBERS"}
	case "AUDITLOG", "COPILOT", "BINDING":
		return nil
	default:
		return []string{"UPDATE"}
	}
}

func expandDelete(resource string) []string {
	switch resource {
	case "AUDITLOG", "SYSTEM_SETTINGS", "TENANT_SETTINGS", "COPILOT":
		return nil
	default:
		return []string{"DELETE"}
	}
}

func appendUnique(slice []string, items ...string) []string {
	seen := make(map[string]bool, len(slice))
	for _, s := range slice {
		seen[s] = true
	}
	for _, item := range items {
		if !seen[item] {
			slice = append(slice, item)
			seen[item] = true
		}
	}
	return slice
}
