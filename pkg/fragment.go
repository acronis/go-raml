package goraml

import (
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FragmentKind int

const (
	UNKNOWN FragmentKind = iota - 1
	LIBRARY
	DATATYPE
	NAMED_EXAMPLE
)

// RAML 1.0 Library
type Library struct {
	Id              string
	Usage           string
	AnnotationTypes map[string]*Shape
	// TODO: Specific to API fragments. Not supported yet.
	// ResourceTypes   map[string]interface{} `yaml:"resourceTypes"`
	Types map[string]*Shape
	Uses  map[string]*Library

	CustomDomainProperties CustomDomainProperties

	Location string
}

func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("must be map")
	}

	l.CustomDomainProperties = make(CustomDomainProperties)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := UnmarshalCustomDomainExtension(l.Location, node, valueNode)
			if err != nil {
				return err
			}
			l.CustomDomainProperties[name] = de
		} else if node.Value == "uses" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Uses = make(map[string]*Library)
			baseDir := filepath.Dir(l.Location)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1].Value
				lib, err := ParseLibrary(filepath.Join(baseDir, path))
				if err != nil {
					return err
				}
				l.Uses[name] = lib
			}
		} else if node.Value == "types" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Types = make(map[string]*Shape)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := MakeShape(data, name, l.Location)
				if err != nil {
					return fmt.Errorf("parse types: %w", err)
				}
				l.Types[name] = shape
				GetRegistry().PutIntoFragment(name, l.Location, shape)
			}
		} else if node.Value == "annotationTypes" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.AnnotationTypes = make(map[string]*Shape)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := MakeShape(data, name, l.Location)
				if err != nil {
					return fmt.Errorf("parse types: %w", err)
				}
				l.AnnotationTypes[name] = shape
				// GetRegistry().Put(name, l.Location, shape)
			}
		} else if node.Value == "usage" {
			if err := valueNode.Decode(&l.Usage); err != nil {
				return err
			}
		}
	}

	return nil
}

func MakeLibrary(path string) *Library {
	return &Library{
		Location: path,
	}
}

// RAML 1.0 DataType
type DataType struct {
	Id    string
	Usage string
	Uses  map[string]*Library
	Shape *Shape

	Location string
}

func (dt *DataType) UnmarshalJSON(value []byte) error {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "type",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: string(value),
				Tag:   "!!str",
			},
		},
	}
	shape, err := MakeShape(node, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return fmt.Errorf("parse types: %w", err)
	}
	dt.Shape = shape
	return nil
}

func (dt *DataType) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return errors.New("must be map")
	}

	shapeValue := &yaml.Node{
		Kind: yaml.MappingNode,
	}
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if node.Value == "usage" {
			dt.Usage = valueNode.Value
		} else if node.Value == "uses" {
			if valueNode.Tag == "!!null" {
				continue
			}

			dt.Uses = make(map[string]*Library)
			baseDir := filepath.Dir(dt.Location)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1].Value
				lib, err := ParseLibrary(filepath.Join(baseDir, path))
				if err != nil {
					return err
				}
				dt.Uses[name] = lib
			}
		} else {
			shapeValue.Content = append(shapeValue.Content, node, valueNode)
		}
	}
	shape, err := MakeShape(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return fmt.Errorf("parse types: %w", err)
	}
	dt.Shape = shape
	return nil
}

func MakeDataType(path string) *DataType {
	return &DataType{
		Location: path,
	}
}

// RAML 1.0 NamedExample
// type NamedExample struct {
// 	Id string
// 	example.Example
// 	Location string
// }
