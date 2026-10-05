package provider_v2

import (
	"fmt"
	"reflect"

	yaml "gopkg.in/yaml.v3"
)

// parseFlowSource decodes a flow source into its YAML document and its top-level values.
func parseFlowSource(source string) (*yaml.Node, map[string]interface{}, error) {
	var document yaml.Node
	if err := yaml.Unmarshal([]byte(source), &document); err != nil {
		return nil, nil, err
	}
	if len(document.Content) == 0 || document.Content[0].Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("the flow must be a YAML mapping")
	}
	var values map[string]interface{}
	if err := document.Decode(&values); err != nil {
		return nil, nil, err
	}
	return &document, values, nil
}

// flowScalar returns the text of a top-level scalar. Kestra reads `id` and `namespace` as
// strings, so an unquoted `id: 2024` is the flow "2024", not the integer 2024.
func flowScalar(document *yaml.Node, key string) (string, bool) {
	root := document.Content[0]
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value != key {
			continue
		}
		value := root.Content[i+1]
		if value.Kind != yaml.ScalarNode || value.Tag == "!!null" {
			return "", false
		}
		return value.Value, true
	}
	return "", false
}

// flowSourcesEqual compares two flow sources semantically, so formatting and comments never
// show up as a change. A source that does not parse is never equal: the difference then
// surfaces in the plan.
func flowSourcesEqual(left, right string) bool {
	var a, b map[string]interface{}
	if err := yaml.Unmarshal([]byte(left), &a); err != nil {
		return false
	}
	if err := yaml.Unmarshal([]byte(right), &b); err != nil {
		return false
	}
	return reflect.DeepEqual(a, b)
}
