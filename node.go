package raml

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/acronis/go-raml/stacktrace"
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
	stacktrace.Position
	raml *RAML
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (r *RAML) makeRootNode(node *yaml.Node, location string) (*Node, error) {
	if node.Tag == "!include" {
		baseDir := filepath.Dir(location)
		fragmentPath := filepath.Join(baseDir, node.Value)
		rdr, err := ReadRawFile(fragmentPath)
		if err != nil {
			return nil, stacktrace.NewWrapped("include: read raw file", err, location, stacktrace.WithNodePosition(node), stacktrace.WithInfo("path", fragmentPath))
		}
		defer func(rdr io.ReadCloser) {
			err = rdr.Close()
			if err != nil {
				log.Fatal(fmt.Errorf("close file error: %w", err))
			}
		}(rdr)
		var value any
		ext := filepath.Ext(node.Value)
		switch ext {
		default:
			v, err := io.ReadAll(rdr)
			if err != nil {
				return nil, stacktrace.NewWrapped("include: read all", err, fragmentPath, stacktrace.WithNodePosition(node))
			}
			value = string(v)
		case ".json":
			var data any
			d := json.NewDecoder(rdr)
			if err := d.Decode(&data); err != nil {
				return nil, stacktrace.NewWrapped("include: json decode", err, fragmentPath, stacktrace.WithNodePosition(node))
			}
			value = data
		case ".yaml", ".yml":
			var data yaml.Node
			d := yaml.NewDecoder(rdr)
			if err := d.Decode(&data); err != nil {
				return nil, stacktrace.NewWrapped("include: yaml decode", err, fragmentPath, stacktrace.WithNodePosition(node))
			}
			value, err = yamlNodeToDataNode(&data, fragmentPath, false)
			if err != nil {
				return nil, stacktrace.NewWrapped("include: yaml node to data node", err, fragmentPath, stacktrace.WithNodePosition(node))
			}
		}
		return &Node{
			Value:    value,
			Location: fragmentPath,
			Position: stacktrace.Position{node.Line, node.Column},
			raml:     r,
		}, nil
	} else if node.Value != "" && node.Value[0] == '{' {
		var value any
		if err := json.Unmarshal([]byte(node.Value), &value); err != nil {
			return nil, stacktrace.NewWrapped("json unmarshal", err, location, stacktrace.WithNodePosition(node))
		}
		return &Node{
			Value:    value,
			Location: location,
			Position: stacktrace.Position{node.Line, node.Column},
			raml:     r,
		}, nil
	}

	return r.makeYamlNode(node, location)
}

func (r *RAML) makeYamlNode(node *yaml.Node, location string) (*Node, error) {
	data, err := yamlNodeToDataNode(node, location, false)
	if err != nil {
		return nil, stacktrace.NewWrapped("yaml node to data node", err, location, stacktrace.WithNodePosition(node))
	}
	return &Node{
		Value:    data,
		Location: location,
		Position: stacktrace.Position{node.Line, node.Column},
		raml:     r,
	}, nil
}

func yamlNodeToDataNode(node *yaml.Node, location string, isInclude bool) (any, error) {
	switch node.Kind {
	default:
		return nil, stacktrace.New("unexpected kind", location, stacktrace.WithInfo("node.kind", stacktrace.Stringer(node.Kind)), stacktrace.WithNodePosition(node))
	case yaml.DocumentNode:
		return yamlNodeToDataNode(node.Content[0], location, isInclude)
	case yaml.ScalarNode:
		switch node.Tag {
		default:
			var val any
			if err := node.Decode(&val); err != nil {
				return nil, stacktrace.NewWrapped("decode scalar node", err, location)
			}
			return val, nil
		case "!!str":
			return node.Value, nil
		case "!!timestamp":
			return node.Value, nil
		case "!include":
			if isInclude {
				return nil, stacktrace.New("nested includes are not allowed", location, stacktrace.WithNodePosition(node))
			}
			// TODO: In case with includes that are explicitly required to be string value, probably need to introduce a new tag.
			// !includestr sounds like a good candidate.
			baseDir := filepath.Dir(location)
			fragmentPath := filepath.Join(baseDir, node.Value)
			// TODO: Need to refactor and move out IO logic from this function.
			r, err := ReadRawFile(filepath.Join(baseDir, node.Value))
			if err != nil {
				return nil, stacktrace.NewWrapped("include: read raw file", err, location, stacktrace.WithNodePosition(node), stacktrace.WithInfo("path", fragmentPath))
			}
			defer func(r io.ReadCloser) {
				err = r.Close()
				if err != nil {
					log.Fatal(fmt.Errorf("close file error: %w", err))
				}
			}(r)
			// TODO: This logic should be more complex because content type may depend on the header reported by remote server.
			ext := filepath.Ext(node.Value)
			switch ext {
			default:
				v, err := io.ReadAll(r)
				if err != nil {
					return nil, stacktrace.NewWrapped("include: read all", err, fragmentPath, stacktrace.WithNodePosition(node))
				}
				return string(v), nil
			case ".yaml", ".yml":
				var data yaml.Node
				d := yaml.NewDecoder(r)
				if err := d.Decode(&data); err != nil {
					return nil, stacktrace.NewWrapped("include: yaml decode", err, fragmentPath, stacktrace.WithNodePosition(node))
				}
				return yamlNodeToDataNode(&data, fragmentPath, true)
			}
		}
	case yaml.MappingNode:
		properties := make(map[string]any, len(node.Content)/2)
		for i := 0; i != len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]
			data, err := yamlNodeToDataNode(value, location, isInclude)
			if err != nil {
				return nil, stacktrace.NewWrapped("yaml node to data node", err, location, stacktrace.WithNodePosition(value))
			}
			properties[key] = data
		}
		return properties, nil
	case yaml.SequenceNode:
		items := make([]any, len(node.Content))
		for i, item := range node.Content {
			data, err := yamlNodeToDataNode(item, location, isInclude)
			if err != nil {
				return nil, stacktrace.NewWrapped("yaml node to data node", err, location, stacktrace.WithNodePosition(item))
			}
			items[i] = data
		}
		return items, nil
	}
}
