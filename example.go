package raml

import (
	"fmt"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

func (ex *Example) decode(node *yaml.Node, valueNode *yaml.Node, location string) error {
	switch node.Value {
	case "strict":
		if err := valueNode.Decode(&ex.Strict); err != nil {
			return StacktraceNewWrapped("decode strict", err, location, WithNodePosition(valueNode))
		}
	case "displayName":
		if err := valueNode.Decode(&ex.DisplayName); err != nil {
			return StacktraceNewWrapped("decode displayName", err, location, WithNodePosition(valueNode))
		}
	case "description":
		if err := valueNode.Decode(&ex.Description); err != nil {
			return StacktraceNewWrapped("decode description", err, location, WithNodePosition(valueNode))
		}
	default:
		if IsCustomDomainExtensionNode(node.Value) {
			deName, de, err := ex.raml.unmarshalCustomDomainExtension(location, node, valueNode)
			if err != nil {
				return StacktraceNewWrapped("unmarshal custom domain extension", err, location, WithNodePosition(valueNode))
			}
			ex.CustomDomainProperties.Set(deName, de)
		}
	}
	return nil
}

func (ex *Example) fill(location string, value *yaml.Node) (*Node, error) {
	var valueKey *yaml.Node
	// First lookup for the "value" key.
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if node.Value == "value" {
			valueKey = valueNode
			break
		}
	}
	// If "value" key is found, then the example is considered as a map with additional properties
	if valueKey != nil {
		for i := 0; i != len(value.Content); i += 2 {
			node := value.Content[i]
			valueNode := value.Content[i+1]
			if err := ex.decode(node, valueNode, location); err != nil {
				return nil, fmt.Errorf("decode example: %w", err)
			}
		}
		n, err := ex.raml.makeRootNode(valueKey, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make node", err, location, WithNodePosition(valueKey))
		}
		return n, nil
	}

	return nil, nil
}

// makeExample creates an example from the given value node
func (r *RAML) makeExample(value *yaml.Node, name string, location string) (*Example, error) {
	ex := &Example{
		Name:                   name,
		Strict:                 true,
		Location:               location,
		Position:               stacktrace.Position{Line: value.Line, Column: value.Column},
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
		raml:                   r,
	}
	// Example can be represented as map in two cases:
	// 1. A value with an example of ObjectShape.
	// 2. A map with the required "value" key that contains the actual example and additional properties of Example.
	if value.Kind == yaml.MappingNode {
		n, err := ex.fill(location, value)
		if err != nil {
			return nil, fmt.Errorf("fill example from mapping node: %w", err)
		}
		if n != nil {
			ex.Data = n
			return ex, nil
		}
	}
	// In all other cases, the example is considered as a value node
	n, err := r.makeRootNode(value, location)
	if err != nil {
		return nil, StacktraceNewWrapped("make node", err, location, WithNodePosition(value))
	}
	ex.Data = n
	return ex, nil
}

// Example represents an example of a shape
type Example struct {
	ID          string
	Name        string
	DisplayName string
	Description string
	Data        *Node

	Strict                 bool
	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}
