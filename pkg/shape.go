package goraml

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

type Shape interface {
	Base() *BaseShape
	// Validate(v interface{}) error
	// Merge(v interface{}) error
	Clone() Shape

	UnmarshalYAML(v []*yaml.Node) error

	// ToJSONSchema() interface{}
	// ToRAMLDataType() interface{}
}

func identifyShapeType(shapeFacets []*yaml.Node) string {
	var t string = STRING
	for i := 0; i != len(shapeFacets); i += 2 {
		node := shapeFacets[i]
		if _, ok := STRING_FACETS[node.Value]; ok {
			t = STRING
		} else if _, ok := INTEGER_FACETS[node.Value]; ok {
			t = INTEGER
		} else if _, ok := FILE_FACETS[node.Value]; ok {
			t = FILE
			break // File has a unique facet
		} else if _, ok := INTEGER_FACETS[node.Value]; ok {
			t = NUMBER
			break // Number has a unique facet
		} else if _, ok := OBJECT_FACETS[node.Value]; ok {
			t = OBJECT
			break
		} else if _, ok := ARRAY_FACETS[node.Value]; ok {
			t = ARRAY
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
	case ANY:
		shape = &AnyShape{BaseShape: *base}
	case NIL:
		shape = &NilShape{BaseShape: *base}
	case OBJECT:
		shape = &ObjectShape{BaseShape: *base}
	case ARRAY:
		shape = &ArrayShape{BaseShape: *base}
	case STRING:
		shape = &StringShape{BaseShape: *base}
	case INTEGER:
		shape = &IntegerShape{BaseShape: *base}
	case NUMBER:
		shape = &NumberShape{BaseShape: *base}
	case DATETIME:
		shape = &DateTimeShape{BaseShape: *base}
	case DATETIME_ONLY:
		shape = &DateTimeOnlyShape{BaseShape: *base}
	case DATE_ONLY:
		shape = &DateOnlyShape{BaseShape: *base}
	case TIME_ONLY:
		shape = &TimeOnlyShape{BaseShape: *base}
	case FILE:
		shape = &FileShape{BaseShape: *base}
	case BOOLEAN:
		shape = &BooleanShape{BaseShape: *base}
	case UNION:
		shape = &UnionShape{BaseShape: *base}
	case JSON:
		shape = &JSONShape{BaseShape: *base}
	}

	if err := shape.UnmarshalYAML(shapeFacets); err != nil {
		return nil, err
	}

	return shape, nil
}

func MakeShape(v *yaml.Node, name string, location string) (*Shape, error) {
	base := BaseShape{Name: name, Location: location, Position: Position{Line: v.Line, Column: v.Column}}

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
					shapeType = JSON
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
			var inherits []*Shape = make([]*Shape, len(shapeTypeNode.Content))
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
			shapeType = COMPOSITE
		}
	}

	s, err := MakeConcreteShape(&base, shapeType, shapeFacets)
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

	s.CustomShapeFacets = make(CustomShapeFacets)
	s.CustomDomainProperties = make(CustomDomainProperties)
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
