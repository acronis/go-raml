package raml

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type BaseShape struct {
	Id          string
	Name        string
	DisplayName *string
	Description *string
	Type        string
	Example     *Example
	Examples    *Examples
	Inherits    []*Shape
	Default     *Node
	Required    *bool
	resolved    bool
	unwrapped   bool

	// To support !include of DataType fragment
	Link *DataType

	CustomShapeFacets           CustomShapeFacets           // Map of custom facets with values
	CustomShapeFacetDefinitions CustomShapeFacetDefinitions // Map of custom facet definitions
	CustomDomainProperties      CustomDomainProperties      // Map of custom annotations

	raml *RAML

	Location string
	Position
}

// IsResolved returns true if the shape is resolved.
func (s *BaseShape) IsResolved() bool {
	return s.resolved
}

// IsUnwrapped returns true if the shape is unwrapped.
func (s *BaseShape) IsUnwrapped() bool {
	return s.unwrapped
}

// String implements fmt.Stringer.
func (s *BaseShape) String() string {
	r := fmt.Sprintf("Type: %s: Name: %s",
		s.Type,
		s.Name,
	)
	if len(s.Inherits) > 0 {
		r = fmt.Sprintf("%s: Inherits", r)
		for _, i := range s.Inherits {
			base := (*i).Base()
			r = fmt.Sprintf("%s: %s", r, base.Name)
		}
	}
	return r
}

// Examples represents a collection of examples.
type Examples struct {
	Id       string
	Examples map[string]*Example

	// To support !include of NamedExample fragment
	Link *NamedExample

	Location string
	Position

	raml *RAML
}

// ShapeBaser is the interface that represents a retriever of a base shape.
type ShapeBaser interface {
	Base() *BaseShape
}

// ShapeValidator is the interface that represents a validator of a RAML shape.
type ShapeValidator interface {
	Validate(v interface{}) error
}

// ShapeInheritor is the interface that represents an inheritor of a RAML shape.
type ShapeInheritor interface {
	// TODO: inplace option?
	Inherit(source Shape) (Shape, error)
}

// ShapeJsonSchema is the interface that represents a maker of JSON schema from a RAML shape.
type ShapeJsonSchema interface {
	ToJSONSchema() interface{}
}

// ShapeRAMLDataType is the interface that represents a maker of RAML data type from a RAML shape.
type ShapeRAMLDataType interface {
	ToRAMLDataType() interface{}
}

// yamlNodesUnmarshaller is the interface that represents an unmarshaller of a RAML shape from YAML nodes.
type yamlNodesUnmarshaller interface {
	unmarshalYAMLNodes(v []*yaml.Node) error
}

// Shape is the interface that represents a RAML shape.
type Shape interface {
	Clone() Shape
	Check() error

	// Inherit is a ShapeInheritor
	Inherit(source Shape) (Shape, error)
	ShapeBaser
	yamlNodesUnmarshaller
	fmt.Stringer
}

// identifyShapeType identifies the type of the shape.
func identifyShapeType(shapeFacets []*yaml.Node) string {
	var t = TypeString
	for i := 0; i != len(shapeFacets); i += 2 {
		node := shapeFacets[i]
		if _, ok := SetOfStringFacets[node.Value]; ok {
			t = TypeString
		} else if _, ok := SetOfNumberFacets[node.Value]; ok {
			t = TypeInteger
			break
		} else if _, ok := SetOfFileFacets[node.Value]; ok {
			t = TypeFile
			break // File has a unique facet
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

// makeConcreteShape creates a new concrete shape.
func (r *RAML) makeConcreteShape(base *BaseShape, shapeType string, shapeFacets []*yaml.Node) (Shape, error) {
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

	if err := shape.unmarshalYAMLNodes(shapeFacets); err != nil {
		return nil, NewWrappedError("unmarshal yaml nodes", err, base.Location,
			WithPosition(&base.Position), WithInfo("shape type", shapeType))
	}

	return shape, nil
}

// MakeBaseShape creates a new base shape which is a base for all shapes.
func (r *RAML) MakeBaseShape(name string, location string, position *Position) *BaseShape {
	return &BaseShape{
		Id:       generateShapeId(),
		Name:     name,
		Location: location,
		Position: *position,

		raml:                        r,
		CustomDomainProperties:      make(CustomDomainProperties),
		CustomShapeFacets:           make(CustomShapeFacets),
		CustomShapeFacetDefinitions: make(CustomShapeFacetDefinitions),
	}
}

// TODO: Temporary workaround
var idCounter int = 1

func generateShapeId() string {
	id := "#" + fmt.Sprint(idCounter)
	idCounter++
	return id
}

// makeShape creates a new shape from the given YAML node.
func (r *RAML) makeShape(v *yaml.Node, name string, location string) (*Shape, error) {
	base := r.MakeBaseShape(name, location, &Position{Line: v.Line, Column: v.Column})

	shapeTypeNode, shapeFacets, err := base.decode(v)
	if err != nil {
		return nil, NewWrappedError("decode", err, location, WithNodePosition(v))
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
				dt, err := r.parseDataType(filepath.Join(baseDir, shapeTypeNode.Value))
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
					return nil, NewError("node kind must be scalar", location, WithNodePosition(node))
				} else if node.Tag == "!include" {
					return nil, NewError("!include is not allowed in multiple inheritance", location, WithNodePosition(node))
				}
				s, err := r.makeShape(node, name, location)
				if err != nil {
					return nil, NewWrappedError("make shape", err, location, WithNodePosition(node))
				}
				inherits[i] = s
			}
			base.Inherits = inherits
			shapeType = TypeComposite
		}
	}

	s, err := r.makeConcreteShape(base, shapeType, shapeFacets)
	if err != nil {
		return nil, NewWrappedError("make concrete shape", err, base.Location, WithPosition(&base.Position))
	}
	ptr := &s
	return ptr, nil
}

// decode decodes the shape from the YAML node.
func (s *BaseShape) decode(value *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	// For inline type declaration
	if value.Kind == yaml.ScalarNode || value.Kind == yaml.SequenceNode {
		return value, nil, nil
	}

	if value.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("value kind must be map")
	}

	var shapeTypeNode *yaml.Node
	var shapeFacets []*yaml.Node

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := s.raml.unmarshalCustomDomainExtension(s.Location, node, valueNode)
			if err != nil {
				return nil, nil, fmt.Errorf("unmarshal custom domain extension: %w", err)
			}
			s.CustomDomainProperties[name] = de
		} else if node.Value == "type" {
			shapeTypeNode = valueNode
		} else if node.Value == "displayName" {
			if err := valueNode.Decode(&s.DisplayName); err != nil {
				return nil, nil, fmt.Errorf("decode display name: %w", err)
			}
		} else if node.Value == "description" {
			if err := valueNode.Decode(&s.Description); err != nil {
				return nil, nil, fmt.Errorf("decode description: %w", err)
			}
		} else if node.Value == "required" {
			if err := valueNode.Decode(&s.Required); err != nil {
				return nil, nil, fmt.Errorf("decode required: %w", err)
			}
		} else if node.Value == "facets" {
			s.CustomShapeFacetDefinitions = make(CustomShapeFacetDefinitions, len(valueNode.Content)/2)
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				property, err := s.raml.makeProperty(name, data, s.Location)
				if err != nil {
					return nil, nil, NewWrappedError("make property", err, s.Location, WithNodePosition(data))
				}
				s.CustomShapeFacetDefinitions[name] = property
			}
		} else if node.Value == "example" {
			if s.Examples != nil {
				return nil, nil, fmt.Errorf("example and examples cannot be defined together")
			}
			example, err := s.raml.makeExample(valueNode, "", s.Location)
			if err != nil {
				return nil, nil, fmt.Errorf("make example: %w", err)
			}
			s.Example = example
		} else if node.Value == "examples" {
			if s.Example != nil {
				return nil, nil, fmt.Errorf("example and examples cannot be defined together")
			}
			if valueNode.Kind == yaml.ScalarNode && valueNode.Tag == "!include" {
				baseDir := filepath.Dir(s.Location)
				n, err := s.raml.parseNamedExample(filepath.Join(baseDir, valueNode.Value))
				if err != nil {
					return nil, nil, fmt.Errorf("parse named example: %w", err)
				}
				s.Examples = &Examples{Link: n, Location: s.Location}
				continue
			} else if valueNode.Kind != yaml.MappingNode {
				return nil, nil, fmt.Errorf("examples must be map")
			}
			examples := make(map[string]*Example, len(valueNode.Content)/2)
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				example, err := s.raml.makeExample(data, name, s.Location)
				if err != nil {
					return nil, nil, fmt.Errorf("make examples: [%d]: %w", j, err)
				}
				examples[name] = example
			}
			s.Examples = &Examples{Examples: examples, Location: s.Location}
		} else if node.Value == "default" {
			n, err := s.raml.makeNode(valueNode, s.Location)
			if err != nil {
				return nil, nil, fmt.Errorf("make node default: %w", err)
			}
			s.Default = n
		} else if node.Value == "allowedTargets" {
			// TODO: Included by annotationTypes
		} else {
			shapeFacets = append(shapeFacets, node, valueNode)
		}
	}

	return shapeTypeNode, shapeFacets, nil
}
