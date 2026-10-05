package provider_v2

import (
	"bytes"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// flowMetadata holds the top-level flow keys set through resource attributes. A nil field is
// an unset attribute and is left out; a configured empty value such as `labels = {}` is kept.
type flowMetadata struct {
	Description *string            `yaml:"description,omitempty"`
	Disabled    *bool              `yaml:"disabled,omitempty"`
	Labels      *map[string]string `yaml:"labels,omitempty"`
}

func (m flowMetadata) isEmpty() bool {
	return m.Description == nil && m.Disabled == nil && m.Labels == nil
}

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

// mergeFlowMetadata adds the attribute-managed keys to the flow source. They are appended as
// text so the source stored in Kestra keeps the comments and layout of the configuration.
// A source where appending does not yield the expected flow, such as a flow-style root
// mapping, is re-encoded instead: same flow, different formatting.
func mergeFlowMetadata(source string, metadata flowMetadata) (string, error) {
	if metadata.isEmpty() {
		return source, nil
	}
	document, values, err := parseFlowSource(source)
	if err != nil {
		return "", err
	}

	appended, err := yaml.Marshal(metadata)
	if err != nil {
		return "", err
	}
	var expected map[string]interface{}
	if err := yaml.Unmarshal(appended, &expected); err != nil {
		return "", err
	}
	// a key already in content would end up twice in the source
	for key := range expected {
		if _, inContent := values[key]; inContent {
			return "", fmt.Errorf("`%s` is set both as an attribute and in content", key)
		}
	}
	maps.Copy(expected, values)

	merged := strings.TrimRight(source, "\n") + "\n" + string(appended)
	if _, mergedValues, err := parseFlowSource(merged); err == nil && reflect.DeepEqual(mergedValues, expected) {
		return merged, nil
	}

	var metadataNode yaml.Node
	if err := metadataNode.Encode(metadata); err != nil {
		return "", err
	}
	root := document.Content[0]
	root.Content = append(root.Content, metadataNode.Content...)
	return encodeFlowSource(document)
}

func removeFlowKeys(document *yaml.Node, keys []string) {
	root := document.Content[0]
	for i := 0; i+1 < len(root.Content); {
		if slices.Contains(keys, root.Content[i].Value) {
			root.Content = slices.Delete(root.Content, i, i+2)
			continue
		}
		i += 2
	}
}

func encodeFlowSource(document *yaml.Node) (string, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

// flowLabels reads labels in either form Kestra accepts: a map, or a list of key/value objects.
func flowLabels(raw interface{}) (map[string]string, error) {
	invalid := fmt.Errorf("labels must be a map or a list of key/value objects")
	labels := map[string]string{}
	switch typed := raw.(type) {
	case nil:
	case map[string]interface{}:
		for key, value := range typed {
			labels[key] = fmt.Sprint(value)
		}
	case []interface{}:
		for _, item := range typed {
			label, ok := item.(map[string]interface{})
			if !ok {
				return nil, invalid
			}
			key, ok := label["key"].(string)
			if !ok {
				return nil, invalid
			}
			labels[key] = fmt.Sprint(label["value"])
		}
	default:
		return nil, invalid
	}
	return labels, nil
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
