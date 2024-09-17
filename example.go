package raml

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-raml/stacktrace"
)

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
				if IsCustomDomainExtensionNode(node.Value) {
					name, de, err := r.unmarshalCustomDomainExtension(location, node, valueNode)
					if err != nil {
						return nil, stacktrace.NewWrapped("unmarshal custom domain extension", err, location, stacktrace.WithNodePosition(valueNode))
					}
					ex.CustomDomainProperties.Set(name, de)
				} else if node.Value == "strict" {
					if err := valueNode.Decode(&ex.Strict); err != nil {
						return nil, stacktrace.NewWrapped("decode strict", err, location, stacktrace.WithNodePosition(valueNode))
					}
				} else if node.Value == "displayName" {
					if err := valueNode.Decode(&ex.DisplayName); err != nil {
						return nil, stacktrace.NewWrapped("decode displayName", err, location, stacktrace.WithNodePosition(valueNode))
					}
				} else if node.Value == "description" {
					if err := valueNode.Decode(&ex.Description); err != nil {
						return nil, stacktrace.NewWrapped("decode description", err, location, stacktrace.WithNodePosition(valueNode))
					}
				}
			}
			n, err := r.makeNode(valueKey, location)
			if err != nil {
				return nil, stacktrace.NewWrapped("make node", err, location, stacktrace.WithNodePosition(valueKey))
			}
			ex.Data = n
			return ex, nil
		}
	}
	// In all other cases, the example is considered as a value node
	n, err := r.makeNode(value, location)
	if err != nil {
		return nil, stacktrace.NewWrapped("make node", err, location, stacktrace.WithNodePosition(value))
	}
	ex.Data = n
	return ex, nil
}

// Example represents an example of a shape
type Example struct {
	Id          string
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
