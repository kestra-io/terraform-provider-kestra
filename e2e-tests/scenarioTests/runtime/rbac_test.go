//go:build e2e

package runtime

import (
	"fmt"
	"testing"
)

// The role in 10-platform grants VIEW, LIST and EXECUTE on flows and nothing more. These
// assertions are what stop a provider bug that widens a role or ignores a binding's
// namespace scope from passing as green: without them, a role that grants everything
// looks identical to a role that grants exactly enough.
func TestAppUserCannotWriteFlows(t *testing.T) {
	c := newClient(t, "app user", "KESTRA_E2E_USER_TOKEN")

	// No CREATE on FLOW in the launcher role.
	c.expectDenied(t, "POST", "/flows", nil)
}

func TestAppUserCannotReadSiblingNamespace(t *testing.T) {
	c := newClient(t, "app user", "KESTRA_E2E_USER_TOKEN")

	forbiddenNS := env(t, "KESTRA_E2E_FORBIDDEN_NAMESPACE")
	forbiddenFlow := env(t, "KESTRA_E2E_FORBIDDEN_FLOW_ID")

	// The binding is scoped to the allowed namespace. Kestra extends a namespace binding
	// to child namespaces, so this sibling must stay out of reach.
	c.expectDenied(t, "GET", fmt.Sprintf("/flows/%s/%s", forbiddenNS, forbiddenFlow), nil)
}

func TestAppUserCannotRunSiblingNamespaceFlow(t *testing.T) {
	c := newClient(t, "app user", "KESTRA_E2E_USER_TOKEN")

	forbiddenNS := env(t, "KESTRA_E2E_FORBIDDEN_NAMESPACE")
	forbiddenFlow := env(t, "KESTRA_E2E_FORBIDDEN_FLOW_ID")

	c.expectDenied(t, "POST", fmt.Sprintf("/executions/%s/%s", forbiddenNS, forbiddenFlow), nil)
}

// The other half of the same contract: the permissions the role *does* grant work. A
// suite that only asserted denials would pass with a role that grants nothing.
func TestAppUserCanReadItsOwnNamespaceFlow(t *testing.T) {
	c := newClient(t, "app user", "KESTRA_E2E_USER_TOKEN")

	namespace := env(t, "KESTRA_E2E_NAMESPACE")
	flowID := env(t, "KESTRA_E2E_FLOW_ID")

	c.expectOK(t, "GET", fmt.Sprintf("/flows/%s/%s", namespace, flowID), nil)
}
