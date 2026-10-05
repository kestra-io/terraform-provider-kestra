package provider_v2

import (
	"reflect"
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

// Appending keeps the configured text byte for byte, so the flow shown in the Kestra UI still
// carries the comments and layout written in Terraform.
func TestMergeFlowMetadataAppendsToTheSource(t *testing.T) {
	disabled, description := true, "Says hello: twice"
	merged, err := mergeFlowMetadata(testFlowSource, flowMetadata{
		Description: &description,
		Disabled:    &disabled,
		Labels:      &map[string]string{"team": "data"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(merged, testFlowSource) {
		t.Fatalf("the configured source was rewritten:\n%s", merged)
	}

	_, values, err := parseFlowSource(merged)
	if err != nil {
		t.Fatal(err)
	}
	if values["disabled"] != true || values["description"] != "Says hello: twice" {
		t.Errorf("metadata not merged: %v", values)
	}
	if labels, _ := flowLabels(values["labels"]); !reflect.DeepEqual(labels, map[string]string{"team": "data"}) {
		t.Errorf("labels = %v", labels)
	}
}

// A flow-style root mapping cannot take appended lines, so the document is re-encoded. An
// explicit false is a configured value and is written too.
func TestMergeFlowMetadataReencodesAFlowStyleRoot(t *testing.T) {
	disabled := false
	merged, err := mergeFlowMetadata(`{id: hello, namespace: company.team, tasks: []}`, flowMetadata{Disabled: &disabled})
	if err != nil {
		t.Fatal(err)
	}
	_, values, err := parseFlowSource(merged)
	if err != nil {
		t.Fatalf("invalid merged source %q: %s", merged, err)
	}
	if values["id"] != "hello" || values["disabled"] != false {
		t.Errorf("unexpected merged flow: %v", values)
	}
}

// A key that the content already defines would end up twice in the source, so the merge
// refuses it.
func TestMergeFlowMetadataRejectsAKeyAlreadyInContent(t *testing.T) {
	description := "from the attribute"
	_, err := mergeFlowMetadata(testFlowSource+"description: from content\n", flowMetadata{Description: &description})
	if err == nil || !strings.Contains(err.Error(), "`description` is set both as an attribute and in content") {
		t.Errorf("expected a conflict error, got %v", err)
	}
}

// `labels = {}` is a configured value, not an unset attribute: it is written, so the flow has
// no labels, and it conflicts with labels in content like any other value.
func TestMergeFlowMetadataKeepsConfiguredEmptyLabels(t *testing.T) {
	empty := map[string]string{}

	merged, err := mergeFlowMetadata(testFlowSource, flowMetadata{Labels: &empty})
	if err != nil {
		t.Fatal(err)
	}
	_, values, err := parseFlowSource(merged)
	if err != nil {
		t.Fatal(err)
	}
	if labels, present := values["labels"]; !present || !reflect.DeepEqual(labels, map[string]interface{}{}) {
		t.Errorf("labels = %v (present %v), want an empty map", labels, present)
	}

	_, err = mergeFlowMetadata(testFlowSource+"labels:\n  team: data\n", flowMetadata{Labels: &empty})
	if err == nil || !strings.Contains(err.Error(), "`labels` is set both as an attribute and in content") {
		t.Errorf("expected a conflict error, got %v", err)
	}
}

func TestMergeFlowMetadataWithoutMetadataKeepsTheSource(t *testing.T) {
	merged, err := mergeFlowMetadata(testFlowSource, flowMetadata{})
	if err != nil {
		t.Fatal(err)
	}
	if merged != testFlowSource {
		t.Errorf("source changed without metadata:\n%s", merged)
	}
}

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

func TestRemoveFlowKeys(t *testing.T) {
	document, _, err := parseFlowSource(testFlowSource + "disabled: true\nlabels:\n  team: data\n")
	if err != nil {
		t.Fatal(err)
	}
	removeFlowKeys(document, []string{"disabled", "labels"})
	source, err := encodeFlowSource(document)
	if err != nil {
		t.Fatal(err)
	}
	if !flowSourcesEqual(source, testFlowSource) {
		t.Errorf("unexpected source after removal:\n%s", source)
	}
}

func TestFlowLabelsAcceptsBothForms(t *testing.T) {
	want := map[string]string{"team": "data", "tier": "1"}
	for _, source := range []string{
		"labels: {team: data, tier: 1}",
		"labels: [{key: team, value: data}, {key: tier, value: 1}]",
	} {
		_, values, err := parseFlowSource(source)
		if err != nil {
			t.Fatal(err)
		}
		labels, err := flowLabels(values["labels"])
		if err != nil {
			t.Errorf("%s: %s", source, err)
			continue
		}
		if !reflect.DeepEqual(labels, want) {
			t.Errorf("%s: labels = %v, want %v", source, labels, want)
		}
	}

	if _, err := flowLabels("team"); err == nil {
		t.Error("a scalar is not a label set")
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
