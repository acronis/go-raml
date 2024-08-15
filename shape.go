package raml

import (
	"errors"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BaseShape struct {
	Id          string // computed
	Name        string // computed
	DisplayName *string
	Description *string
	Type        string
	Examples    []*Example
	Inherits    []*Shape // computed
	Default     *Node
	Required    *bool

	// To support !include
	Link *DataType // computed

	CustomShapeFacets           CustomShapeFacets // computed. Map of custom facets with values
	CustomShapeFacetDefinitions CustomShapeFacetDefinitions
	CustomDomainProperties      CustomDomainProperties // computed. Map of custom annotations

	Location string // computed.
	Position        // computed.
}

// ShapeBaser is the interface that represents a retriever of a base shape.
type ShapeBaser interface {
	Base() *BaseShape
}

// ShapeValidator is the interface that represents a validator of a RAML shape.
type ShapeValidator interface {
	Validate(v interface{}) error
}

// ShapeMerger is the interface that represents a merger of a RAML shape.
type ShapeMerger interface {
	Merge(v interface{}) error
}

// ShapeCloner is the interface that represents a maker of a clone of a RAML shape.
type ShapeCloner interface {
	Clone() Shape
}

// ShapeJsonSchema is the interface that represents a maker of JSON schema from a RAML shape.
type ShapeJsonSchema interface {
	ToJSONSchema() interface{}
}

// ShapeRAMLDataType is the interface that represents a maker of RAML data type from a RAML shape.
type ShapeRAMLDataType interface {
	ToRAMLDataType() interface{}
}

// YAMLNodesUnmarshaller is the interface that represents an unmarshaller of a RAML shape from YAML nodes.
type YAMLNodesUnmarshaller interface {
	UnmarshalYAMLNodes(v []*yaml.Node) error
}

// Shape is the interface that represents a RAML shape.
type Shape interface {
	ShapeBaser
	ShapeCloner
	YAMLNodesUnmarshaller
}

func identifyShapeType(shapeFacets []*yaml.Node) string {
	var t = TypeString
	for i := 0; i != len(shapeFacets); i += 2 {
		node := shapeFacets[i]
		if _, ok := SetOfStringFacets[node.Value]; ok {
			t = TypeString
		} else if _, ok := SetOfIntegerFacets[node.Value]; ok {
			t = TypeInteger
		} else if _, ok := SetOfFileFacets[node.Value]; ok {
			t = TypeFile
			break // File has a unique facet
		} else if _, ok := SetOfIntegerFacets[node.Value]; ok {
			t = TypeNumber
			break // Number has a unique facet
		} else if _, ok := SetOfObjectFacets[node.Value]; ok {
			t = TypeObject
			break
		} else if _, ok := SetOfArrayFacets[node.Value]; ok {
			t = TypeArray
			break
		}
	}
	return t
}

func MakeConcreteShape(base *BaseShape, shapeType string, shapeFacets []*yaml.Node) (Shape, error) {
	base.Type = shapeType

	// NOTE: Shape resolution is performed in a separate stage.
	var shape Shape
	switch shapeType {
	default:
		// NOTE: UnknownShape is a special type of shape that will be resolved later.
		shape = &UnknownShape{BaseShape: *base}
	case TypeAny:
		shape = &AnyShape{BaseShape: *base}
	case TypeNil:
		shape = &NilShape{BaseShape: *base}
	case TypeObject:
		shape = &ObjectShape{BaseShape: *base}
	case TypeArray:
		shape = &ArrayShape{BaseShape: *base}
	case TypeString:
		shape = &StringShape{BaseShape: *base}
	case TypeInteger:
		shape = &IntegerShape{BaseShape: *base}
	case TypeNumber:
		shape = &NumberShape{BaseShape: *base}
	case TypeDatetime:
		shape = &DateTimeShape{BaseShape: *base}
	case TypeDatetimeOnly:
		shape = &DateTimeOnlyShape{BaseShape: *base}
	case TypeDateOnly:
		shape = &DateOnlyShape{BaseShape: *base}
	case TypeTimeOnly:
		shape = &TimeOnlyShape{BaseShape: *base}
	case TypeFile:
		shape = &FileShape{BaseShape: *base}
	case TypeBoolean:
		shape = &BooleanShape{BaseShape: *base}
	case TypeUnion:
		shape = &UnionShape{BaseShape: *base}
	case TypeJSON:
		shape = &JSONShape{BaseShape: *base}
	}

	if err := shape.UnmarshalYAMLNodes(shapeFacets); err != nil {
		return nil, fmt.Errorf("unmarshal yaml nodes: shape type: %s: err: %w", shapeType, err)
	}

	return shape, nil
}

func MakeBaseShape(name string, location string, position *Position) *BaseShape {
	return &BaseShape{
		Name:     name,
		Location: location,
		Position: *position,

		CustomDomainProperties: make(CustomDomainProperties),
		CustomShapeFacets:      make(CustomShapeFacets),
	}
}

func MakeShape(v *yaml.Node, name string, location string) (*Shape, error) {
	base := MakeBaseShape(name, location, &Position{Line: v.Line, Column: v.Column})

	shapeTypeNode, shapeFacets, err := base.Decode(v)
	if err != nil {
		return nil, err
	}

	var shapeType string
	if shapeTypeNode == nil {
		shapeType = identifyShapeType(shapeFacets)
	} else {
		switch shapeTypeNode.Kind {
		default:
			return nil, fmt.Errorf("type must be string or array")
		case yaml.ScalarNode:
			if shapeTypeNode.Tag == "!!str" {
				shapeType = shapeTypeNode.Value
				if shapeType == "" {
					shapeType = identifyShapeType(shapeFacets)
				} else if shapeType[0] == '{' {
					shapeType = TypeJSON
				}
			} else if shapeTypeNode.Tag == "!include" {
				baseDir := filepath.Dir(location)
				dt, err := ParseDataType(filepath.Join(baseDir, shapeTypeNode.Value))
				if err != nil {
					return nil, fmt.Errorf("parse data: %w", err)
				}
				base.Link = dt
			} else {
				return nil, fmt.Errorf("type must be string")
			}
		case yaml.SequenceNode:
			var inherits = make([]*Shape, len(shapeTypeNode.Content))
			for i, node := range shapeTypeNode.Content {
				if node.Kind != yaml.ScalarNode {
					return nil, errors.New("must be string")
				}
				s, err := MakeShape(node, name, location)
				if err != nil {
					return nil, err
				}
				inherits[i] = s
			}
			base.Inherits = inherits
			shapeType = TypeComposite
		}
	}

	s, err := MakeConcreteShape(base, shapeType, shapeFacets)
	if err != nil {
		return nil, err
	}
	ptr := &s
	if _, ok := s.(*UnknownShape); ok {
		GetRegistry().AppendUnresolvedShape(ptr)
	}
	return ptr, nil
}

func (s *BaseShape) Decode(value *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	// For inline type declaration
	if value.Kind == yaml.ScalarNode || value.Kind == yaml.SequenceNode {
		return value, nil, nil
	}

	if value.Kind != yaml.MappingNode {
		return nil, nil, errors.New("must be map")
	}

	var shapeTypeNode *yaml.Node
	var shapeFacets []*yaml.Node

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := UnmarshalCustomDomainExtension(s.Location, node, valueNode)
			if err != nil {
				return nil, nil, err
			}
			s.CustomDomainProperties[name] = de
		} else if node.Value == "type" {
			shapeTypeNode = valueNode
		} else if node.Value == "displayName" {
			if err := valueNode.Decode(&s.DisplayName); err != nil {
				return nil, nil, err
			}
		} else if node.Value == "description" {
			if err := valueNode.Decode(&s.Description); err != nil {
				return nil, nil, err
			}
		} else if node.Value == "required" {
			if err := valueNode.Decode(&s.Required); err != nil {
				return nil, nil, err
			}
		} else if node.Value == "facets" {
			s.CustomShapeFacetDefinitions = make(CustomShapeFacetDefinitions)
			// Map nodes come in pairs in order [key, value]
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				shape, err := MakeShape(data, name, s.Location)
				if err != nil {
					return nil, nil, err
				}
				s.CustomShapeFacetDefinitions[name] = shape
			}
		} else if node.Value == "example" {
			example := &Example{
				Location: s.Location,
			}
			if err := example.UnmarshalYAML(valueNode); err != nil {
				return nil, nil, err
			}
			s.Examples = append(s.Examples, example)
		} else if node.Value == "examples" {
			// TODO: Add NamedExample support
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				example := &Example{
					Name:     name,
					Location: s.Location,
				}
				if err := example.UnmarshalYAML(data); err != nil {
					return nil, nil, err
				}
				s.Examples = append(s.Examples, example)
			}
		} else if node.Value == "default" {
			if err := valueNode.Decode(&s.Default); err != nil {
				return nil, nil, err
			}
		} else if node.Value == "allowedTargets" {
			// TODO: Included by annotationTypes
		} else {
			shapeFacets = append(shapeFacets, node, valueNode)
		}
	}

	return shapeTypeNode, shapeFacets, nil
}
