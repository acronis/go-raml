package raml

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
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
	ID    string
	Value any

	// TODO: Probably not required but could be useful for reusing raw data fragments.
	// Link *Node

	Location string
	stacktrace.Position
	// positions locates the values inside Value in the source, see
	// nodePositions. It is nil for a Node that was not read from YAML.
	positions nodePositions
	raml      *RAML
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.Value)
}

func (r *RAML) makeIncludedNode(node *yaml.Node, location string) (*Node, error) {
	baseDir := filepath.Dir(location)
	fragmentPath := filepath.Join(baseDir, node.Value)
	rdr, err := ReadRawFile(fragmentPath)
	if err != nil {
		return nil, StacktraceNewWrapped("include: read raw file", err, location, WithNodePosition(node),
			stacktrace.WithInfo("path", fragmentPath))
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
		v, errDecode := io.ReadAll(rdr)
		if errDecode != nil {
			return nil, StacktraceNewWrapped("include: read all", errDecode, fragmentPath, WithNodePosition(node))
		}
		value = string(v)
	case ".json":
		var data any
		d := json.NewDecoder(rdr)
		if errDecode := d.Decode(&data); errDecode != nil {
			return nil, StacktraceNewWrapped("include: json decode", errDecode, fragmentPath, WithNodePosition(node))
		}
		value = data
	case ".yaml", ".yml":
		var data yaml.Node
		d := yaml.NewDecoder(rdr)
		if errDecode := d.Decode(&data); errDecode != nil {
			return nil, StacktraceNewWrapped("include: yaml decode", errDecode, fragmentPath, WithNodePosition(node))
		}
		value, err = yamlNodeToDataNode(&data, fragmentPath, false)
		if err != nil {
			return nil, StacktraceNewWrapped("include: yaml node to data node", err, fragmentPath,
				WithNodePosition(node))
		}
	}
	return &Node{
		Value:    value,
		Location: fragmentPath,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		raml:     r,
	}, nil
}

func (r *RAML) makeRootNode(node *yaml.Node, location string) (*Node, error) {
	if node.Tag == TagInclude {
		return r.makeIncludedNode(node, location)
	}

	if node.Value != "" && node.Value[0] == '{' {
		var value any
		if err := json.Unmarshal([]byte(node.Value), &value); err != nil {
			return nil, StacktraceNewWrapped("json unmarshal", err, location, WithNodePosition(node))
		}
		return &Node{
			Value:    value,
			Location: location,
			Position: stacktrace.Position{Line: node.Line, Column: node.Column},
			raml:     r,
		}, nil
	}

	return r.makeYamlNode(node, location)
}

func (r *RAML) makeYamlNode(node *yaml.Node, location string) (*Node, error) {
	positions := make(nodePositions)
	data, err := yamlNodeToDataNodeAt(node, location, false, "$", positions)
	if err != nil {
		return nil, StacktraceNewWrapped("yaml node to data node", err, location, WithNodePosition(node))
	}
	return &Node{
		Value:     data,
		Location:  location,
		Position:  stacktrace.Position{Line: node.Line, Column: node.Column},
		positions: positions,
		raml:      r,
	}, nil
}

func scalarNodeToDataNode(node *yaml.Node, location string, isInclude bool) (any, error) {
	switch node.Tag {
	default:
		var val any
		if err := node.Decode(&val); err != nil {
			return nil, StacktraceNewWrapped("decode scalar node", err, location)
		}
		return val, nil
	case TagStr:
		return node.Value, nil
	case TagTimestamp:
		return node.Value, nil
	case TagInclude:
		if isInclude {
			return nil, StacktraceNew("nested includes are not allowed", location, WithNodePosition(node))
		}
		// TODO: In case with includes that are explicitly required to be string value, probably need to introduce
		//  a new tag.
		// !includestr sounds like a good candidate.
		baseDir := filepath.Dir(location)
		fragmentPath := filepath.Join(baseDir, node.Value)
		// TODO: Need to refactor and move out IO logic from this function.
		r, err := ReadRawFile(filepath.Join(baseDir, node.Value))
		if err != nil {
			return nil, StacktraceNewWrapped("include: read raw file", err, location, WithNodePosition(node),
				stacktrace.WithInfo("path", fragmentPath))
		}
		defer func(r io.ReadCloser) {
			err = r.Close()
			if err != nil {
				log.Fatal(fmt.Errorf("close file error: %w", err))
			}
		}(r)
		// TODO: This logic should be more complex because content type may depend on the header reported
		//  by remote server.
		ext := filepath.Ext(node.Value)
		switch ext {
		default:
			v, errRead := io.ReadAll(r)
			if errRead != nil {
				return nil, StacktraceNewWrapped("include: read all", errRead, fragmentPath,
					WithNodePosition(node))
			}
			return string(v), nil
		case ".yaml", ".yml":
			var data yaml.Node
			d := yaml.NewDecoder(r)
			if errDecode := d.Decode(&data); errDecode != nil {
				return nil, StacktraceNewWrapped("include: yaml decode", errDecode, fragmentPath,
					WithNodePosition(node))
			}
			return yamlNodeToDataNode(&data, fragmentPath, true)
		}
	}
}

func yamlNodeToDataNode(node *yaml.Node, location string, isInclude bool) (any, error) {
	return yamlNodeToDataNodeAt(node, location, isInclude, "$", nil)
}

// yamlNodeToDataNodeAt converts the node to data. When positions is not nil,
// it records the position of every mapping key and sequence item under its
// JSON path, built as validation builds ctxPath: "$", "$[0]", "$.key".
func yamlNodeToDataNodeAt(
	node *yaml.Node, location string, isInclude bool, path string, positions nodePositions,
) (any, error) {
	switch node.Kind {
	default:
		return nil, StacktraceNew("unexpected kind", location,
			stacktrace.WithInfo("node.kind", stacktrace.Stringer(node.Kind)), WithNodePosition(node))
	case yaml.AliasNode:
		return nil, StacktraceNew("alias nodes are not supported", location, WithNodePosition(node))
	case yaml.DocumentNode:
		return yamlNodeToDataNodeAt(node.Content[0], location, isInclude, path, positions)
	case yaml.ScalarNode:
		return scalarNodeToDataNode(node, location, isInclude)
	case yaml.MappingNode:
		properties := make(map[string]any, len(node.Content)/2)
		if len(node.Content)%2 != 0 {
			return nil, StacktraceNew("mapping node should have even number of children", location,
				WithNodePosition(node))
		}
		for i := 0; i != len(node.Content); i += 2 {
			key := node.Content[i].Value
			value := node.Content[i+1]
			keyPath := path + "." + key
			if positions != nil {
				positions[keyPath] = stacktrace.Position{Line: node.Content[i].Line, Column: node.Content[i].Column}
			}
			data, err := yamlNodeToDataNodeAt(value, location, isInclude, keyPath, positions)
			if err != nil {
				return nil, StacktraceNewWrapped("yaml node to data node", err, location,
					WithNodePosition(value))
			}
			properties[key] = data
		}
		return properties, nil
	case yaml.SequenceNode:
		items := make([]any, len(node.Content))
		for i, item := range node.Content {
			itemPath := path + "[" + strconv.Itoa(i) + "]"
			if positions != nil {
				positions[itemPath] = stacktrace.Position{Line: item.Line, Column: item.Column}
			}
			data, err := yamlNodeToDataNodeAt(item, location, isInclude, itemPath, positions)
			if err != nil {
				return nil, StacktraceNewWrapped("yaml node to data node", err, location, WithNodePosition(item))
			}
			items[i] = data
		}
		return items, nil
	}
}
