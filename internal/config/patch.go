package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func PatchYAMLFile(filename, key, value string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	parts := strings.Split(key, ".")
	docNode := findDocNode(&root)
	if docNode == nil {
		return fmt.Errorf("config file has no document node")
	}

	leaf, err := findOrCreateNode(docNode, parts)
	if err != nil {
		return err
	}

	fieldType, err := resolveFieldType(parts)
	if err != nil {
		return err
	}

	setNodeValue(leaf, value, fieldType)

	tmpFile, err := os.CreateTemp(filepath.Dir(filename), ".config-*.yaml")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()

	encoder := yaml.NewEncoder(tmpFile)
	encoder.SetIndent(2)
	if err := encoder.Encode(&root); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to encode config: %w", err)
	}
	if err := encoder.Close(); err != nil {
		tmpFile.Close()
		os.Remove(tmpName)
		return fmt.Errorf("failed to flush config: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpName, filename); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func findDocNode(root *yaml.Node) *yaml.Node {
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		return root.Content[0]
	}
	if root.Kind == yaml.MappingNode {
		return root
	}
	return nil
}

func findOrCreateNode(parent *yaml.Node, parts []string) (*yaml.Node, error) {
	current := parent
	for i, part := range parts {
		isLast := i == len(parts)-1
		child := findMappingKey(current, part)
		if child == nil {
			if isLast {
				valueNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: "",
					Tag:   "",
				}
				keyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: part,
				}
				current.Content = append(current.Content, keyNode, valueNode)
				return valueNode, nil
			}
			mappingNode := &yaml.Node{
				Kind: yaml.MappingNode,
			}
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: part,
			}
			current.Content = append(current.Content, keyNode, mappingNode)
			current = mappingNode
			continue
		}

		var valueNode *yaml.Node
		idx := indexOfMappingKey(current, part)
		if idx >= 0 && idx+1 < len(current.Content) {
			valueNode = current.Content[idx+1]
		}

		if isLast {
			return valueNode, nil
		}

		if valueNode.Kind != yaml.MappingNode {
			newMapping := &yaml.Node{
				Kind: yaml.MappingNode,
			}
			if valueNode.HeadComment != "" {
				newMapping.HeadComment = valueNode.HeadComment
				valueNode.HeadComment = ""
			}
			if valueNode.LineComment != "" {
				newMapping.LineComment = valueNode.LineComment
				valueNode.LineComment = ""
			}
			if valueNode.FootComment != "" {
				newMapping.FootComment = valueNode.FootComment
				valueNode.FootComment = ""
			}
			current.Content[idx+1] = newMapping
			current = newMapping
		} else {
			current = valueNode
		}
	}
	return nil, fmt.Errorf("unexpected end of path")
}

func findMappingKey(node *yaml.Node, key string) *yaml.Node {
	if node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i]
		}
	}
	return nil
}

func indexOfMappingKey(node *yaml.Node, key string) int {
	if node.Kind != yaml.MappingNode {
		return -1
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return i
		}
	}
	return -1
}

func resolveFieldType(parts []string) (reflect.Type, error) {
	t := reflect.TypeOf(Config{})
	for i, part := range parts {
		field, found := findFieldTypeByYAMLTag(t, part)
		if !found {
			return nil, fmt.Errorf("config key %q not found at part %q", strings.Join(parts, "."), part)
		}
		if field.Type.Kind() == reflect.Ptr {
			field.Type = field.Type.Elem()
		}
		if i == len(parts)-1 {
			return field.Type, nil
		}
		t = field.Type
	}
	return nil, fmt.Errorf("config key %q not found", strings.Join(parts, "."))
}

func findFieldTypeByYAMLTag(t reflect.Type, tag string) (reflect.StructField, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "" || yamlTag == "-" {
			continue
		}
		name := strings.SplitN(yamlTag, ",", 2)[0]
		if name == tag {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

func setNodeValue(node *yaml.Node, value string, fieldType reflect.Type) {
	if fieldType == reflect.TypeOf(Duration(0)) {
		node.Kind = yaml.ScalarNode
		node.Value = value
		node.Tag = ""
		return
	}

	switch fieldType.Kind() {
	case reflect.String:
		node.Kind = yaml.ScalarNode
		node.Value = value
		node.Tag = ""
	case reflect.Int:
		if _, err := strconv.Atoi(value); err == nil {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = "!!int"
		} else {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = ""
		}
	case reflect.Float64:
		if _, err := strconv.ParseFloat(value, 64); err == nil {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = "!!float"
		} else {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = ""
		}
	case reflect.Bool:
		if _, err := strconv.ParseBool(value); err == nil {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = "!!bool"
		} else {
			node.Kind = yaml.ScalarNode
			node.Value = value
			node.Tag = ""
		}
	default:
		node.Kind = yaml.ScalarNode
		node.Value = value
		node.Tag = ""
	}
}