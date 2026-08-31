//go:build e2e

package e2e

import (
	"fmt"
	"testing"
)

// A flow the app user has no business creating. Well-formed on purpose: if the request
// were malformed, a 400 would look like a denial and the assertion would prove nothing.
const unauthorizedFlow = `id: e2e_should_not_exist
namespace: %s
tasks:
  - id: noop
    type: io.kestra.plugin.core.log.Log
    message: this flow must never be created
`

// The launcher role in 10-platform grants VIEW, LIST and EXECUTE on flows and nothing
// more. These assertions are what stop a provider bug that widens a role or drops a
// binding's namespace scope from passing as green: without them, a role granting
// everything looks identical to a role granting exactly enough.
func TestAppUserCannotCreateFlows(t *testing.T) {
	c := asUser(t, "app user", "KESTRA_E2E_USER_TOKEN")
	namespace := env(t, "KESTRA_E2E_NAMESPACE")

	c.expectDenied(t, request{
		method:  "POST",
		path:    "/flows",
		raw:     fmt.Sprintf(unauthorizedFlow, namespace),
		rawType: "application/x-yaml",
	})
}

func TestAppUserCannotReadSiblingNamespace(t *testing.T) {
	c := asUser(t, "app user", "KESTRA_E2E_USER_TOKEN")

	// The binding is scoped to the allowed namespace. Kestra extends a namespace binding
	// to child namespaces, so this sibling must stay out of reach.
	c.expectDenied(t, request{
		method: "GET",
		path: fmt.Sprintf("/flows/%s/%s",
			env(t, "KESTRA_E2E_FORBIDDEN_NAMESPACE"), env(t, "KESTRA_E2E_FORBIDDEN_FLOW_ID")),
	})
}

func TestAppUserCannotRunSiblingNamespaceFlow(t *testing.T) {
	c := asUser(t, "app user", "KESTRA_E2E_USER_TOKEN")

	c.expectDenied(t, request{
		method: "POST",
		path: fmt.Sprintf("/executions/%s/%s",
			env(t, "KESTRA_E2E_FORBIDDEN_NAMESPACE"), env(t, "KESTRA_E2E_FORBIDDEN_FLOW_ID")),
		// the forbidden flow declares no inputs, but send a body of the right shape so a
		// rejection can only be about permission
		form: map[string]string{},
	})
}

// The other half of the same contract: the permissions the role *does* grant work. A
// suite that only asserted denials would pass with a role that grants nothing at all.
func TestAppUserCanReadItsOwnNamespaceFlow(t *testing.T) {
	c := asUser(t, "app user", "KESTRA_E2E_USER_TOKEN")

	c.expectOK(t, request{
		method: "GET",
		path:   fmt.Sprintf("/flows/%s/%s", env(t, "KESTRA_E2E_NAMESPACE"), env(t, "KESTRA_E2E_FLOW_ID")),
	})
}

// Closes the loop on stage 00. The platform admin's role and tenant-wide binding are
// what stage 10 depended on to apply at all, but nothing so far asserts they granted
// what was intended — a stage-10 failure would have been the only signal, and it would
// have pointed at the wrong stage.
func TestPlatformAdminCanManageItsTenant(t *testing.T) {
	c := asUser(t, "platform admin", "KESTRA_E2E_ADMIN_TOKEN")

	namespace := env(t, "KESTRA_E2E_NAMESPACE")

	// granted by the NAMESPACE and FLOW permissions in 00-bootstrap
	c.expectOK(t, request{method: "GET", path: fmt.Sprintf("/namespaces/%s", namespace)})
	c.expectOK(t, request{method: "GET", path: fmt.Sprintf("/flows/%s/%s", namespace, env(t, "KESTRA_E2E_FLOW_ID"))})

	// The role deliberately omits TENANT: the platform admin manages inside its tenant,
	// it does not create tenants. Proves the role is scoped, not blanket-admin.
	c.expectDenied(t, request{
		method:   "POST",
		path:     "/tenants",
		instance: true,
		raw:      `{"id":"e2e-should-not-exist","name":"Should Not Exist"}`,
		rawType:  "application/json",
	})
}
