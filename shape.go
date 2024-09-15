package raml

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ShapeVisitor[T any] interface {
	VisitObjectShape(s *ObjectShape) T
	VisitArrayShape(s *ArrayShape) T
	VisitStringShape(s *StringShape) T
	VisitNumberShape(s *NumberShape) T
	VisitIntegerShape(s *IntegerShape) T
	VisitBooleanShape(s *BooleanShape) T
	VisitFileShape(s *FileShape) T
	VisitUnionShape(s *UnionShape) T
	VisitDateTimeShape(s *DateTimeShape) T
	VisitDateTimeOnlyShape(s *DateTimeOnlyShape) T
	VisitDateOnlyShape(s *DateOnlyShape) T
	VisitTimeOnlyShape(s *TimeOnlyShape) T
	VisitRecursiveShape(s *RecursiveShape) T
	VisitJSONShape(s *JSONShape) T
	VisitAnyShape(s *AnyShape) T
	VisitNilShape(s *NilShape) T
}

type BaseShape struct {
	Id          string
	Name        string
	DisplayName *string
	Description *string
	Type        string
	TypeLabel   string // Used to store the label either the link or type value
	Example     *Example
	Examples    *Examples
	Inherits    []*Shape
	Alias       *Shape
	Default     *Node
	Required    *bool
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
	Id  string
	Map map[string]*Example

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
	Validate(v interface{}, ctxPath string) error
}

// ShapeInheritor is the interface that represents an inheritor of a RAML shape.
type ShapeInheritor interface {
	// TODO: inplace option?
	Inherit(source Shape) (Shape, error)
}

// ShapeJSONSchema is the interface that provide clone implementation for a RAML shape.
type ShapeCloner interface {
	// Public entry point
	Clone() Shape

	// NOTE: Private clone performs a deep copy where applicable.
	// Public method calls private method to simplify public interface.
	clone(history []Shape) Shape
}

type ShapeChecker interface {
	Check() error
}

// yamlNodesUnmarshaller is the interface that represents an unmarshaller of a RAML shape from YAML nodes.
type yamlNodesUnmarshaller interface {
	unmarshalYAMLNodes(v []*yaml.Node) error
}

// Shape is the interface that represents a RAML shape.
type Shape interface {
	ShapeInheritor
	ShapeBaser
	ShapeChecker
	ShapeCloner
	ShapeValidator

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

func (r *RAML) MakeJSONShape(base *BaseShape, rawSchema string) (Shape, error) {
	base.Type = "json"

	var schema *JSONSchema
	err := json.Unmarshal([]byte(rawSchema), &schema)
	if err != nil {
		return nil, NewWrappedError("unmarshal json", err, base.Location, WithPosition(&base.Position))
	}

	return &JSONShape{BaseShape: *base, Raw: rawSchema, Schema: schema}, nil
}

// MakeConcreteShape creates a new concrete shape.
func (r *RAML) MakeConcreteShape(base *BaseShape, shapeType string, shapeFacets []*yaml.Node) (Shape, error) {
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
			return nil, NewError("type must be string or array", location, WithNodePosition(shapeTypeNode))
		case yaml.ScalarNode:
			if shapeTypeNode.Tag == "!!str" {
				shapeType = shapeTypeNode.Value
				if shapeType == "" {
					shapeType = identifyShapeType(shapeFacets)
				} else if shapeType[0] == '{' {
					s, err := r.MakeJSONShape(base, shapeType)
					if err != nil {
						return nil, NewWrappedError("make json shape", err, location, WithNodePosition(shapeTypeNode))
					}
					return &s, nil
				}
			} else if shapeTypeNode.Tag == "!include" {
				baseDir := filepath.Dir(location)
				dt, err := r.parseDataType(filepath.Join(baseDir, shapeTypeNode.Value))
				if err != nil {
					return nil, NewWrappedError("parse data", err, location, WithNodePosition(shapeTypeNode))
				}
				base.TypeLabel = shapeTypeNode.Value
				base.Link = dt
			} else {
				return nil, NewError("type must be string", location, WithNodePosition(shapeTypeNode))
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

	s, err := r.MakeConcreteShape(base, shapeType, shapeFacets)
	if err != nil {
		return nil, NewWrappedError("make concrete shape", err, base.Location, WithPosition(&base.Position))
	}
	ptr := &s
	if _, ok := s.(*UnknownShape); ok {
		r.unresolvedShapes.PushBack(ptr)
	}
	return ptr, nil
}

// decode decodes the shape from the YAML node.
func (s *BaseShape) decode(value *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	// For inline type declaration
	if value.Kind == yaml.ScalarNode || value.Kind == yaml.SequenceNode {
		return value, nil, nil
	}

	if value.Kind != yaml.MappingNode {
		return nil, nil, NewError("value kind must be map", s.Location, WithNodePosition(value))
	}

	var shapeTypeNode *yaml.Node
	shapeFacets := make([]*yaml.Node, 0)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := s.raml.unmarshalCustomDomainExtension(s.Location, node, valueNode)
			if err != nil {
				return nil, nil, NewWrappedError("unmarshal custom domain extension", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomDomainProperties[name] = de
		} else if node.Value == "type" {
			shapeTypeNode = valueNode
		} else if node.Value == "displayName" {
			if err := valueNode.Decode(&s.DisplayName); err != nil {
				return nil, nil, NewWrappedError("decode display name", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "description" {
			if err := valueNode.Decode(&s.Description); err != nil {
				return nil, nil, NewWrappedError("decode description", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "required" {
			if err := valueNode.Decode(&s.Required); err != nil {
				return nil, nil, NewWrappedError("decode required", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "facets" {
			s.CustomShapeFacetDefinitions = make(CustomShapeFacetDefinitions, len(valueNode.Content)/2)
			for j := 0; j != len(valueNode.Content); j += 2 {
				nodeName := valueNode.Content[j].Value
				data := valueNode.Content[j+1]

				propertyName, hasImplicitOptional := s.raml.chompImplicitOptional(nodeName)
				property, err := s.raml.makeProperty(nodeName, propertyName, data, s.Location, hasImplicitOptional)
				if err != nil {
					return nil, nil, NewWrappedError("make property", err, s.Location, WithNodePosition(data))
				}
				s.CustomShapeFacetDefinitions[property.Name] = property
			}
		} else if node.Value == "example" {
			if s.Examples != nil {
				return nil, nil, NewError("example and examples cannot be defined together", s.Location, WithNodePosition(valueNode))
			}
			example, err := s.raml.makeExample(valueNode, "", s.Location)
			if err != nil {
				return nil, nil, NewWrappedError("make example", err, s.Location, WithNodePosition(valueNode))
			}
			s.Example = example
		} else if node.Value == "examples" {
			if s.Example != nil {
				return nil, nil, NewError("example and examples cannot be defined together", s.Location, WithNodePosition(valueNode))
			}
			if valueNode.Kind == yaml.ScalarNode && valueNode.Tag == "!include" {
				baseDir := filepath.Dir(s.Location)
				n, err := s.raml.parseNamedExample(filepath.Join(baseDir, valueNode.Value))
				if err != nil {
					return nil, nil, NewWrappedError("parse named example", err, s.Location, WithNodePosition(valueNode))
				}
				s.Examples = &Examples{Link: n, Location: s.Location}
				continue
			} else if valueNode.Kind != yaml.MappingNode {
				return nil, nil, NewError("examples must be map", s.Location, WithNodePosition(valueNode))
			}
			examples := make(map[string]*Example, len(valueNode.Content)/2)
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				example, err := s.raml.makeExample(data, name, s.Location)
				if err != nil {
					return nil, nil, NewWrappedError(fmt.Sprintf("make examples: [%d]", j), err, s.Location, WithNodePosition(data))
				}
				examples[name] = example
			}
			s.Examples = &Examples{Map: examples, Location: s.Location}
		} else if node.Value == "default" {
			n, err := s.raml.makeNode(valueNode, s.Location)
			if err != nil {
				return nil, nil, NewWrappedError("make node default", err, s.Location, WithNodePosition(valueNode))
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
