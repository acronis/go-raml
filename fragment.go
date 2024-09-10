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

type LocationGetter interface {
	GetLocation() string
}

type Fragment interface {
	LocationGetter
}

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
	raml     *RAML
}

func (l *Library) GetLocation() string {
	return l.Location
}

type LibraryLink struct {
	Id    string
	Value string

	Link *Library

	Location string
	Position
}

// UnmarshalYAML unmarshals a Library from a yaml.Node, implementing the yaml.Unmarshaler interface
func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("must be map")
	}
	l.CustomDomainProperties = make(CustomDomainProperties)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := l.raml.unmarshalCustomDomainExtension(l.Location, node, valueNode)
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
				shape, err := l.raml.makeShape(data, name, l.Location)
				if err != nil {
					return NewWrappedError("parse types: make shape", err, l.Location, WithNodePosition(data))
				}
				l.Types[name] = shape
				l.raml.PutTypeIntoFragment(name, l.Location, shape)
				l.raml.PutShapePtr(shape)
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
				shape, err := l.raml.makeShape(data, name, l.Location)
				if err != nil {
					return NewWrappedError("parse annotation types: make shape", err, l.Location, WithNodePosition(data))
				}
				l.AnnotationTypes[name] = shape
				l.raml.PutAnnotationTypeIntoFragment(name, l.Location, shape)
				l.raml.PutShapePtr(shape)
			}
		} else if node.Value == "usage" {
			if err := valueNode.Decode(&l.Usage); err != nil {
				return NewWrappedError("parse usage: value node decode", err, l.Location, WithNodePosition(valueNode))
			}
		}
	}

	return nil
}

func (r *RAML) MakeLibrary(path string) *Library {
	return &Library{
		Location: path,
		raml:     r,
	}
}

// DataType is the RAML 1.0 DataType
type DataType struct {
	Id    string
	Usage string
	Uses  map[string]*LibraryLink
	Shape *Shape

	Location string
	raml     *RAML
}

func (dt *DataType) GetLocation() string {
	return dt.Location
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
	shape, err := dt.raml.makeShape(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return NewWrappedError("parse types: make shape", err, dt.Location, WithNodePosition(shapeValue))
	}
	dt.Shape = shape
	return nil
}

func (r *RAML) MakeDataType(path string) *DataType {
	return &DataType{
		Location: path,
		raml:     r,
	}
}

func (r *RAML) MakeJsonDataType(value []byte, path string) (*DataType, error) {
	dt := r.MakeDataType(path)
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
	Id  string
	Map map[string]*Example

	Location string
	raml     *RAML
}

func (ne *NamedExample) GetLocation() string {
	return ne.Location
}

func (r *RAML) MakeNamedExample(path string) *NamedExample {
	return &NamedExample{
		Location: path,
		raml:     r,
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
		example, err := ne.raml.makeExample(valueNode, node.Value, ne.Location)
		if err != nil {
			return NewWrappedError("make example", err, ne.Location, WithNodePosition(valueNode))
		}
		examples[node.Value] = example
	}
	ne.Map = examples

	return nil
}
