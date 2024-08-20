package raml

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FragmentKind int

const (
	FragmentUnknown FragmentKind = iota - 1
	FragmentLibrary
	FragmentDataType
	FragmentNamedExample
)

// Library is the RAML 1.0 Library
type Library struct {
	Id              string
	Usage           string
	AnnotationTypes map[string]*Shape
	// TODO: Specific to API fragments. Not supported yet.
	// ResourceTypes   map[string]interface{} `yaml:"resourceTypes"`
	Types map[string]*Shape
	Uses  map[string]*LibraryLink

	CustomDomainProperties CustomDomainProperties

	Location string
}

type LibraryLink struct {
	Id    string
	Value string

	Link *Library

	Location string
	Position
}

func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("must be map")
	}
	l.CustomDomainProperties = make(CustomDomainProperties)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := UnmarshalCustomDomainExtension(l.Location, node, valueNode)
			if err != nil {
				return NewWrappedError("unmarshal custom domain extension", err, l.Location, WithNodePosition(valueNode))
			}
			l.CustomDomainProperties[name] = de
		} else if node.Value == "uses" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Uses = make(map[string]*LibraryLink, len(valueNode.Content)/2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1]
				l.Uses[name] = &LibraryLink{
					Value:    path.Value,
					Location: l.Location,
					Position: Position{path.Line, path.Column},
				}
			}
		} else if node.Value == "types" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Types = make(map[string]*Shape, len(valueNode.Content)/2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := MakeShape(data, name, l.Location)
				if err != nil {
					return NewWrappedError("parse types: make shape", err, l.Location, WithNodePosition(data))
				}
				l.Types[name] = shape
				GetRegistry().PutIntoFragment(name, l.Location, shape)
			}
		} else if node.Value == "annotationTypes" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.AnnotationTypes = make(map[string]*Shape, len(valueNode.Content)/2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := MakeShape(data, name, l.Location)
				if err != nil {
					return NewWrappedError("parse annotation types: make shape", err, l.Location, WithNodePosition(data))
				}
				l.AnnotationTypes[name] = shape
				// GetRegistry().Put(name, l.Location, shape)
			}
		} else if node.Value == "usage" {
			if err := valueNode.Decode(&l.Usage); err != nil {
				return NewWrappedError("parse usage: value node decode", err, l.Location, WithNodePosition(valueNode))
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

// DataType is the RAML 1.0 DataType
type DataType struct {
	Id    string
	Usage string
	Uses  map[string]*LibraryLink
	Shape *Shape

	Location string
}

func (dt *DataType) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("value kind must be map")
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

			dt.Uses = make(map[string]*LibraryLink, len(valueNode.Content)/2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1]
				dt.Uses[name] = &LibraryLink{
					Value:    path.Value,
					Location: dt.Location,
					Position: Position{path.Line, path.Column},
				}
			}
		} else {
			shapeValue.Content = append(shapeValue.Content, node, valueNode)
		}
	}
	shape, err := MakeShape(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return NewWrappedError("parse types: make shape", err, dt.Location, WithNodePosition(shapeValue))
	}
	dt.Shape = shape
	return nil
}

func MakeDataType(path string) *DataType {
	return &DataType{
		Location: path,
	}
}

func MakeJsonDataType(value []byte, path string) (*DataType, error) {
	dt := MakeDataType(path)
	// Convert to yaml node to reuse the same data node creation interface
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{
				Kind:  yaml.ScalarNode,
				Value: "type",
				Tag:   "!!str",
			},
			{
				Kind:  yaml.ScalarNode,
				Value: string(value),
				Tag:   "!!str",
			},
		},
	}
	if err := node.Decode(&dt); err != nil {
		return nil, NewWrappedError("decode fragment", err, path)
	}
	return dt, nil
}

// NamedExample is the RAML 1.0 NamedExample
type NamedExample struct {
	Id       string
	Examples map[string]*Example

	Location string
}

func MakeNamedExample(path string) *NamedExample {
	return &NamedExample{
		Location: path,
	}
}

func (ne *NamedExample) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("must be map")
	}
	examples := make(map[string]*Example, len(value.Content)/2)
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		example, err := MakeExample(valueNode, node.Value, ne.Location)
		if err != nil {
			return NewWrappedError("make example", err, ne.Location, WithNodePosition(valueNode))
		}
		examples[node.Value] = example
	}
	ne.Examples = examples

	return nil
}
