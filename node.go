package raml

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Nodes []*Node

func (n Nodes) String() string {
	vals := make([]string, len(n))
	for i, node := range n {
		vals[i] = node.String()
	}
	return strings.Join(vals, ", ")
}

type Node struct {
	Id    string
	Value any

	// TODO: Probably not required but could be useful for reusing raw data fragments.
	//Link *Node

	Location string
	Position
	raml *RAML
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (r *RAML) makeNode(node *yaml.Node, location string) (*Node, error) {
	data, err := yamlNodeToDataNode(node, location, false)
	if err != nil {
		return nil, NewWrappedError("yaml node to data node", err, location)
	}
	return &Node{
		Value:    data,
		Location: location,
		Position: Position{node.Line, node.Column},
		raml:     r,
	}, nil
}

func yamlNodeToDataNode(node *yaml.Node, location string, isInclude bool) (any, error) {
	switch node.Kind {
	default:
		return nil, NewError("unexpected kind", location, WithInfo("node.kind", Stringer(node.Kind)), WithNodePosition(node))
	case yaml.DocumentNode:
		return yamlNodeToDataNode(node.Content[0], location, isInclude)
	case yaml.ScalarNode:
		switch node.Tag {
		default:
			var val any
			if err := node.Decode(&val); err != nil {
				return nil, NewWrappedError("decode scalar node", err, location)
			}
			return val, nil
		case "!!str":
			// TODO: Unmarshalling JSON string may be not always desirable, but it's extensively used in RAML.
			if node.Value != "" && node.Value[0] == '{' {
				var val any
				if err := json.Unmarshal([]byte(node.Value), &val); err != nil {
					return nil, NewWrappedError("json unmarshal", err, location, WithNodePosition(node))
				}
				return val, nil
			}
			return node.Value, nil
		case "!!timestamp":
			return node.Value, nil
		case "!include":
			if isInclude {
				return nil, NewError("nested includes are not allowed", location, WithNodePosition(node))
			}
			// TODO: In case with includes that are explicitly required to be string value, probably need to introduce a new tag.
			// !includestr sounds like a good candidate.
			baseDir := filepath.Dir(location)
			fragmentPath := filepath.Join(baseDir, node.Value)
			// TODO: Need to refactor and move out IO logic from this function.
			r, err := ReadRawFile(filepath.Join(baseDir, node.Value))
			if err != nil {
				return nil, NewWrappedError("include: read raw file", err, location, WithNodePosition(node), WithInfo("path", fragmentPath))
			}
			defer func(r io.ReadCloser) {
				err = r.Close()
				if err != nil {
					log.Fatal(fmt.Errorf("close file error: %w", err))
				}
			}(r)
			// TODO: This logic should be more complex because content type may depend on the header reported by remote server.
			ext := filepath.Ext(node.Value)
			if ext == ".json" {
				var data any
				d := json.NewDecoder(r)
				if err := d.Decode(&data); err != nil {
					return nil, NewWrappedError("include: json decode", err, fragmentPath, WithNodePosition(node))
				}
				return data, nil
			} else if ext == ".yaml" || ext == ".yml" {
				var data yaml.Node
				d := yaml.NewDecoder(r)
				if err := d.Decode(&data); err != nil {
					return nil, NewWrappedError("include: yaml decode", err, fragmentPath, WithNodePosition(node))
				}
				return yamlNodeToDataNode(&data, fragmentPath, true)
			} else {
				v, err := io.ReadAll(r)
				if err != nil {
					return nil, NewWrappedError("include: read all", err, fragmentPath, WithNodePosition(node))
				}
				return string(v), nil
			}
		}
	case yaml.MappingNode:
		properties := make(map[string]any, len(node.Content)/2)
		for i := 0; i != len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]
			data, err := yamlNodeToDataNode(value, location, isInclude)
			if err != nil {
				return nil, NewWrappedError("yaml node to data node", err, location, WithNodePosition(value))
			}
			properties[key] = data
		}
		return properties, nil
	case yaml.SequenceNode:
		items := make([]any, len(node.Content))
		for i, item := range node.Content {
			data, err := yamlNodeToDataNode(item, location, isInclude)
			if err != nil {
				return nil, NewWrappedError("yaml node to data node", err, location, WithNodePosition(item))
			}
			items[i] = data
		}
		return items, nil
	}
}
