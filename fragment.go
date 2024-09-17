package raml

import (
	"path/filepath"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-raml/stacktrace"
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
	AnnotationTypes *orderedmap.OrderedMap[string, *Shape]
	// TODO: Specific to API fragments. Not supported yet.
	// ResourceTypes   map[string]interface{} `yaml:"resourceTypes"`
	Types *orderedmap.OrderedMap[string, *Shape]
	Uses  *orderedmap.OrderedMap[string, *LibraryLink]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

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
	stacktrace.Position
}

// UnmarshalYAML unmarshals a Library from a yaml.Node, implementing the yaml.Unmarshaler interface
func (l *Library) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return stacktrace.New("must be map", l.Location, stacktrace.WithNodePosition(value))
	}
	l.CustomDomainProperties = orderedmap.New[string, *DomainExtension](0)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := l.raml.unmarshalCustomDomainExtension(l.Location, node, valueNode)
			if err != nil {
				return stacktrace.NewWrapped("unmarshal custom domain extension", err, l.Location, stacktrace.WithNodePosition(valueNode))
			}
			l.CustomDomainProperties.Set(name, de)
		} else if node.Value == "uses" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Uses = orderedmap.New[string, *LibraryLink](len(valueNode.Content) / 2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1]
				l.Uses.Set(name, &LibraryLink{
					Value:    path.Value,
					Location: l.Location,
					Position: stacktrace.Position{path.Line, path.Column},
				})
			}
		} else if node.Value == "types" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.Types = orderedmap.New[string, *Shape](len(valueNode.Content) / 2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := l.raml.makeShape(data, name, l.Location)
				if err != nil {
					return stacktrace.NewWrapped("parse types: make shape", err, l.Location, stacktrace.WithNodePosition(data))
				}
				l.Types.Set(name, shape)
				l.raml.PutTypeIntoFragment(name, l.Location, shape)
				l.raml.PutShapePtr(shape)
			}
		} else if node.Value == "annotationTypes" {
			if valueNode.Tag == "!!null" {
				continue
			}

			l.AnnotationTypes = orderedmap.New[string, *Shape](len(valueNode.Content) / 2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := l.raml.makeShape(data, name, l.Location)
				if err != nil {
					return stacktrace.NewWrapped("parse annotation types: make shape", err, l.Location, stacktrace.WithNodePosition(data))
				}
				l.AnnotationTypes.Set(name, shape)
				l.raml.PutAnnotationTypeIntoFragment(name, l.Location, shape)
				l.raml.PutShapePtr(shape)
			}
		} else if node.Value == "usage" {
			if err := valueNode.Decode(&l.Usage); err != nil {
				return stacktrace.NewWrapped("parse usage: value node decode", err, l.Location, stacktrace.WithNodePosition(valueNode))
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
	Uses  *orderedmap.OrderedMap[string, *LibraryLink]
	Shape *Shape

	Location string
	raml     *RAML
}

func (dt *DataType) GetLocation() string {
	return dt.Location
}

func (dt *DataType) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return stacktrace.New("must be map", dt.Location, stacktrace.WithNodePosition(value))
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

			dt.Uses = orderedmap.New[string, *LibraryLink](len(valueNode.Content) / 2)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				path := valueNode.Content[j+1]
				dt.Uses.Set(name, &LibraryLink{
					Value:    path.Value,
					Location: dt.Location,
					Position: stacktrace.Position{path.Line, path.Column},
				})
			}
		} else {
			shapeValue.Content = append(shapeValue.Content, node, valueNode)
		}
	}
	shape, err := dt.raml.makeShape(shapeValue, filepath.Base(dt.Location), dt.Location)
	if err != nil {
		return stacktrace.NewWrapped("parse types: make shape", err, dt.Location, stacktrace.WithNodePosition(shapeValue))
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
		return nil, stacktrace.NewWrapped("decode fragment", err, path)
	}
	return dt, nil
}

// NamedExample is the RAML 1.0 NamedExample
type NamedExample struct {
	Id  string
	Map *orderedmap.OrderedMap[string, *Example]

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
		return stacktrace.New("must be map", ne.Location, stacktrace.WithNodePosition(value))
	}
	examples := orderedmap.New[string, *Example](len(value.Content) / 2)
	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		example, err := ne.raml.makeExample(valueNode, node.Value, ne.Location)
		if err != nil {
			return stacktrace.NewWrapped("make example", err, ne.Location, stacktrace.WithNodePosition(valueNode))
		}
		examples.Set(node.Value, example)
	}
	ne.Map = examples

	return nil
}
