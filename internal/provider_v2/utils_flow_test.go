package provider_v2

import (
	"strings"
	"testing"
)

const testFlowSource = `id: hello
namespace: company.team
# kept in the source stored by Kestra
tasks:
  - id: log
    type: io.kestra.plugin.core.log.Log
    message: "{{ flow.id }}" # inline comment
`

func TestFlowScalarReadsIdentityAsText(t *testing.T) {
	document, _, err := parseFlowSource("id: 2024\nnamespace: company.team\nlabels: {}\n")
	if err != nil {
		t.Fatal(err)
	}
	if id, ok := flowScalar(document, "id"); !ok || id != "2024" {
		t.Errorf("id = %q, %v; want 2024", id, ok)
	}
	if _, ok := flowScalar(document, "labels"); ok {
		t.Error("a mapping is not a scalar")
	}
	if _, ok := flowScalar(document, "description"); ok {
		t.Error("an absent key has no value")
	}
}

func TestFlowSourcesEqual(t *testing.T) {
	cases := []struct {
		name  string
		other string
		equal bool
	}{
		{"formatting and comments", "id: hello\nnamespace: company.team\ntasks: [{id: log, type: io.kestra.plugin.core.log.Log, message: '{{ flow.id }}'}]\n", true},
		{"changed task", strings.Replace(testFlowSource, "{{ flow.id }}", "{{ flow.namespace }}", 1), false},
		{"added metadata", testFlowSource + "disabled: false\n", false},
	}
	for _, c := range cases {
		if got := flowSourcesEqual(testFlowSource, c.other); got != c.equal {
			t.Errorf("%s: equal = %v, want %v", c.name, got, c.equal)
		}
	}
}
