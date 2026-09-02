//go:build e2e

package e2e

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

type execution struct {
	ID    string `json:"id"`
	State struct {
		Current string `json:"current"`
	} `json:"state"`
	TaskRunList []struct {
		TaskID string `json:"taskId"`
		State  struct {
			Current string `json:"current"`
		} `json:"state"`
	} `json:"taskRunList"`
}

// The assertion this whole suite exists for: the app user Terraform created can run the
// flow Terraform created, it succeeds, and it returns what the fixtures say it should.
//
// Because the flow reads a secret, a KV entry and a namespace file, a SUCCESS here also
// says kestra_namespace_secret, kestra_kv and kestra_namespace_file wrote values the
// engine can actually resolve — not merely values the API accepted.
func TestAppUserRunsFlowSuccessfully(t *testing.T) {
	c := asUser(t, "app user", "KESTRA_E2E_USER_TOKEN")

	namespace := env(t, "KESTRA_E2E_NAMESPACE")
	flowID := env(t, "KESTRA_E2E_FLOW_ID")
	greeting := "hello-from-e2e"

	body := c.expectOK(t, request{
		method: "POST",
		path:   fmt.Sprintf("/executions/%s/%s", namespace, flowID),
		// wait=true has the server hold the response until the execution finishes, so
		// there is no polling loop to get flaky.
		query: url.Values{
			"wait":   []string{"true"},
			"labels": []string{"suite:e2e-scenario"},
		},
		// inputs are multipart/form-data, not query parameters
		form: map[string]string{"greeting": greeting},
	})

	exec := decode[execution](t, body, "execution")

	if exec.State.Current != "SUCCESS" {
		t.Fatalf("execution %s finished in state %q, want SUCCESS\ntask states: %s\nbody: %s",
			exec.ID, exec.State.Current, taskStates(exec), truncate(body))
	}

	// Flow-level outputs are fetched rather than read off the execution: Kestra 2.0.0-rc9
	// moved them out of the execution into their own table, served from this endpoint, so
	// the payload above no longer carries them. Reading them as the app user also needs
	// EXECUTION ACCESS_OUTPUTS, which the launcher role grants.
	outputs := decode[map[string]any](t, c.expectOK(t, request{
		method: "GET",
		path:   fmt.Sprintf("/outputs/executions/%s", exec.ID),
	}), "execution outputs")

	// Proves the input reached the flow (first half) and that the KV entry resolved
	// (second half). A greeting sent the wrong way would silently fall back to the
	// flow's default and show up here.
	want := greeting + ":" + env(t, "KESTRA_E2E_KV_VALUE")
	got, ok := outputs["result"].(string)
	if !ok {
		t.Fatalf("execution %s has no string output \"result\"; outputs=%v", exec.ID, outputs)
	}
	if got != want {
		t.Errorf("flow output = %q, want %q — the input, the KV entry or both did not reach the flow", got, want)
	}

	t.Logf("execution %s SUCCESS, result=%q", exec.ID, got)
}

// The test suite kestra_test manages is currently only checked for round-trip. Running
// it proves Kestra can actually execute what the provider persisted.
//
// Run as the super-admin on purpose: the question here is whether the resource
// round-trips into something runnable, so authorization should not be a variable.
func TestManagedTestSuitePasses(t *testing.T) {
	c := asSuperAdmin(t)

	namespace := env(t, "KESTRA_E2E_NAMESPACE")
	suiteID := env(t, "KESTRA_E2E_TESTSUITE_ID")

	body := c.expectOK(t, request{
		method: "POST",
		path:   fmt.Sprintf("/tests/%s/%s/run", namespace, suiteID),
	})

	// The response shape has moved between versions, so assert on what is stable: no
	// result in the payload may report a failed state.
	result := truncate(body)
	for _, bad := range []string{`"state":"FAILED"`, `"state":"ERROR"`, `"status":"FAILED"`, `"status":"ERROR"`} {
		if strings.Contains(result, bad) {
			t.Fatalf("test suite %s reported a failure (%s)\nbody: %s", suiteID, bad, result)
		}
	}
	t.Logf("test suite %s ran clean: %s", suiteID, result)
}

func taskStates(e execution) string {
	if len(e.TaskRunList) == 0 {
		return "(no task runs — the failure happened before any task started)"
	}
	var b strings.Builder
	for _, tr := range e.TaskRunList {
		fmt.Fprintf(&b, "%s=%s ", tr.TaskID, tr.State.Current)
	}
	return b.String()
}
