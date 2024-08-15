package raml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Node struct {
	Id string
	*ScalarNode
	*ObjectNode
	*ArrayNode

	// TODO: Possibly needs !include support?
	Location string
	Position
}

func MakeNode(node *yaml.Node, location string) (*Node, error) {
	n := &Node{
		Location: location,
	}
	if err := n.UnmarshalYAML(node); err != nil {
		return nil, fmt.Errorf("parse node: %w", err)
	}
	return n, nil
}

func (dn *Node) UnmarshalYAML(value *yaml.Node) error {
	node := YamlNodeToDataNode(value)
	switch t := node.(type) {
	default:
		return fmt.Errorf("unexpected type %T", t)
	case *ScalarNode:
		dn.ScalarNode = t
	case *ObjectNode:
		dn.ObjectNode = t
	case *ArrayNode:
		dn.ArrayNode = t
	}
	return nil
}

type ScalarNode struct {
	Value string
}

type ObjectNode struct {
	Properties map[string]any
}

type ArrayNode struct {
	Items []any
}

func YamlNodeToDataNode(node *yaml.Node) any {
	switch node.Kind {
	default:
		return nil
	case yaml.ScalarNode:
		return &ScalarNode{
			Value: node.Value,
		}
	case yaml.MappingNode:
		properties := make(map[string]any)
		for i := 0; i != len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]
			properties[key] = YamlNodeToDataNode(value)
		}
		return &ObjectNode{Properties: properties}
	case yaml.SequenceNode:
		items := make([]any, len(node.Content))
		for i, item := range node.Content {
			items[i] = YamlNodeToDataNode(item)
		}
		return &ArrayNode{Items: items}
	}
}
