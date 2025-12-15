package raml

import (
	"errors"
	"fmt"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

const (
	ExampleValue = "value"
)

var ErrValueKeyNotFound = errors.New("value key not found")

func (ex *Example) decode(node *yaml.Node, valueNode *yaml.Node, location string) error {
	switch node.Value {
	case FacetStrict:
		if err := valueNode.Decode(&ex.Strict); err != nil {
			return StacktraceNewWrapped("decode strict", err, location, WithNodePosition(valueNode))
		}
	case FacetDisplayName:
		if err := valueNode.Decode(&ex.DisplayName); err != nil {
			return StacktraceNewWrapped("decode displayName", err, location, WithNodePosition(valueNode))
		}
	case FacetDescription:
		if err := valueNode.Decode(&ex.Description); err != nil {
			return StacktraceNewWrapped("decode description", err, location, WithNodePosition(valueNode))
		}
	case FacetValue:
		n, err := ex.raml.makeRootNode(valueNode, location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, location, WithNodePosition(valueNode))
		}
		ex.Data = n
	default:
		if IsCustomDomainExtensionNode(node.Value) {
			deName, de, err := ex.raml.unmarshalCustomDomainExtension(location, node, valueNode)
			if err != nil {
				return StacktraceNewWrapped("unmarshal custom domain extension", err, location, WithNodePosition(valueNode))
			}
			ex.CustomDomainProperties.Set(deName, de)
		} else {
			return StacktraceNew("unknown field", location, WithNodePosition(valueNode),
				stacktrace.WithInfo("field", node.Value))
		}
	}
	return nil
}

func (ex *Example) fill(location string, value *yaml.Node) error {
	var valueKey *yaml.Node
	// First lookup for the "value" key.
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if node.Value == ExampleValue {
			valueKey = valueNode
			break
		}
	}

	if valueKey == nil {
		return ErrValueKeyNotFound
	}

	// If "value" key is found, then the example is considered as a map with additional properties
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if err := ex.decode(node, valueNode, location); err != nil {
			return fmt.Errorf("decode example: %w", err)
		}
	}
	return nil
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
		err := ex.fill(location, value)
		if err == nil {
			return ex, nil
		} else if !errors.Is(err, ErrValueKeyNotFound) {
			return nil, fmt.Errorf("fill example from mapping node: %w", err)
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
