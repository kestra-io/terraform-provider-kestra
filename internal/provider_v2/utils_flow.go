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

// flowIdentityKeys holds the `id` and `namespace` that content leaves out, taken from the
// flow_id and namespace attributes.
type flowIdentityKeys struct {
	Id        *string `yaml:"id,omitempty"`
	Namespace *string `yaml:"namespace,omitempty"`
}

// mergeFlowMetadata adds the attribute-managed keys to the end of the flow source.
func mergeFlowMetadata(source string, metadata flowMetadata) (string, error) {
	if metadata.isEmpty() {
		return source, nil
	}
	return addFlowKeys(source, metadata, false)
}

// addFlowIdentity puts `id` and `namespace` first in a source that leaves them out, as the
// flows written by keep_original_source = false did. A key in content is kept as is.
func addFlowIdentity(source, flowId, namespace string) (string, error) {
	_, values, err := parseFlowSource(source)
	if err != nil {
		return "", err
	}
	var keys flowIdentityKeys
	if _, ok := values["id"]; !ok && flowId != "" {
		keys.Id = &flowId
	}
	if _, ok := values["namespace"]; !ok && namespace != "" {
		keys.Namespace = &namespace
	}
	if keys.Id == nil && keys.Namespace == nil {
		return source, nil
	}
	return addFlowKeys(source, keys, true)
}

// addFlowKeys adds keys to the flow source as text, before or after it, so the source stored
// in Kestra keeps the comments and layout of the configuration. A source where this does not
// yield the expected flow, such as a flow-style root mapping, is re-encoded instead: same
// flow, different formatting.
func addFlowKeys(source string, keys interface{}, first bool) (string, error) {
	document, values, err := parseFlowSource(source)
	if err != nil {
		return "", err
	}

	added, err := yaml.Marshal(keys)
	if err != nil {
		return "", err
	}
	var expected map[string]interface{}
	if err := yaml.Unmarshal(added, &expected); err != nil {
		return "", err
	}
	// a key already in content would end up twice in the source
	for key := range expected {
		if _, inContent := values[key]; inContent {
			return "", fmt.Errorf("`%s` is set both as an attribute and in content", key)
		}
	}
	maps.Copy(expected, values)

	merged := strings.TrimRight(source, "\n") + "\n" + string(added)
	if first {
		merged = string(added) + source
	}
	if _, mergedValues, err := parseFlowSource(merged); err == nil && reflect.DeepEqual(mergedValues, expected) {
		return merged, nil
	}

	var keysNode yaml.Node
	if err := keysNode.Encode(keys); err != nil {
		return "", err
	}
	root := document.Content[0]
	if first {
		root.Content = append(keysNode.Content, root.Content...)
	} else {
		root.Content = append(root.Content, keysNode.Content...)
	}
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
