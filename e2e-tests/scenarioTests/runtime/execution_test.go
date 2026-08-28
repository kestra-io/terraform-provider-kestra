//go:build e2e

package runtime

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
	Outputs     map[string]any `json:"outputs"`
	TaskRunList []struct {
		TaskID string `json:"taskId"`
		State  struct {
			Current string `json:"current"`
		} `json:"state"`
	} `json:"taskRunList"`
}

// The assertion the whole suite exists for: the app user Terraform created can run the
// flow Terraform created, it succeeds, and it returns what the fixtures say it should.
//
// Because the flow reads a secret, a KV entry and a namespace file, a SUCCESS here also
// says kestra_namespace_secret, kestra_kv and kestra_namespace_file wrote values the
// engine can actually resolve — not merely values the API accepted.
func TestAppUserRunsFlowSuccessfully(t *testing.T) {
	c := newClient(t, "app user", "KESTRA_E2E_USER_TOKEN")

	namespace := env(t, "KESTRA_E2E_NAMESPACE")
	flowID := env(t, "KESTRA_E2E_FLOW_ID")
	greeting := "hello-from-e2e"

	// wait=true has the server hold the response until the execution finishes, so there
	// is no polling loop to get flaky.
	body := c.expectOK(t, "POST",
		fmt.Sprintf("/executions/%s/%s", namespace, flowID),
		url.Values{
			"wait":     []string{"true"},
			"greeting": []string{greeting},
			"labels":   []string{"suite:e2e-scenario"},
		})

	exec := decode[execution](t, body, "execution")

	if exec.State.Current != "SUCCESS" {
		t.Fatalf("execution %s finished in state %q, want SUCCESS\ntask states: %s\nbody: %s",
			exec.ID, exec.State.Current, taskStates(exec), truncate(body))
	}

	want := greeting + ":" + env(t, "KESTRA_E2E_KV_VALUE")
	got, ok := exec.Outputs["result"].(string)
	if !ok {
		t.Fatalf("execution %s has no string output \"result\"; outputs=%v", exec.ID, exec.Outputs)
	}
	if got != want {
		t.Errorf("flow output = %q, want %q — the input, the KV entry or both did not reach the flow", got, want)
	}

	t.Logf("execution %s SUCCESS, result=%q", exec.ID, got)
}

// The test suite kestra_test manages is currently only checked for round-trip. Running
// it proves Kestra can actually execute what the provider persisted.
func TestManagedTestSuitePasses(t *testing.T) {
	c := newClient(t, "platform admin", "KESTRA_E2E_ADMIN_TOKEN")

	namespace := env(t, "KESTRA_E2E_NAMESPACE")
	suiteID := env(t, "KESTRA_E2E_TESTSUITE_ID")

	body := c.expectOK(t, "POST", fmt.Sprintf("/tests/%s/%s/run", namespace, suiteID), nil)

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
	out := ""
	for _, tr := range e.TaskRunList {
		out += fmt.Sprintf("%s=%s ", tr.TaskID, tr.State.Current)
	}
	return out
}
