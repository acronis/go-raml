package raml

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Node struct {
	Id    string
	Value any

	Link *Node

	Location string
	Position
}

func MakeNode(node *yaml.Node, location string) (*Node, error) {
	n := &Node{Location: location, Position: Position{node.Line, node.Column}}

	switch node.Kind {
	default:
		return nil, fmt.Errorf("unexpected kind %v", node.Kind)
	case yaml.ScalarNode:
		switch node.Tag {
		default:
			if err := node.Decode(&n.Value); err != nil {
				return nil, err
			}
		case "!!str":
			if node.Value != "" && node.Value[0] == '{' {
				if err := json.Unmarshal([]byte(node.Value), &n.Value); err != nil {
					return nil, err
				}
			} else {
				n.Value = node.Value
			}
		// TODO: In case with includes that are explicitly required to be string value, probably need to introduce a new tag.
		// !includestr sounds like a good candidate.
		case "!include":
			baseDir := filepath.Dir(location)
			fragmentPath := filepath.Join(baseDir, node.Value)
			// TODO: Need to refactor and move out IO logic from this function.
			r, err := ReadRawFile(filepath.Join(baseDir, node.Value))
			if err != nil {
				return nil, err
			}
			defer func(r io.ReadCloser) {
				err = r.Close()
				if err != nil {
					log.Fatal(fmt.Errorf("close file error: %w", err))
				}
			}(r)
			// TODO: This logic should be more complex because content type may depend on the header reported by remote server.
			link := &Node{Location: fragmentPath}
			ext := filepath.Ext(node.Value)
			if ext == ".json" {
				d := json.NewDecoder(r)
				if err := d.Decode(&link.Value); err != nil {
					return nil, err
				}
			} else if ext == ".yaml" || ext == ".yml" {
				d := yaml.NewDecoder(r)
				if err := d.Decode(&link.Value); err != nil {
					return nil, err
				}
			} else {
				v, err := io.ReadAll(r)
				if err != nil {
					return nil, err
				}
				link.Value = v
			}
			n.Link = link
		}
		return n, nil
	case yaml.MappingNode:
		if err := node.Decode(&n.Value); err != nil {
			return nil, err
		}
	case yaml.SequenceNode:
		if err := node.Decode(&n.Value); err != nil {
			return nil, err
		}
	}
	return n, nil
}
