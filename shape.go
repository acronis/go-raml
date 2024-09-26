package raml

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
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
	ID          string
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

	// CustomShapeFacets is a map of custom facets with values
	CustomShapeFacets *orderedmap.OrderedMap[string, *Node]
	// CustomShapeFacetDefinitions is an object properties share the same syntax with custom shape facets.
	CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
	// CustomDomainProperties is a map of custom annotations
	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	raml *RAML

	Location string
	stacktrace.Position
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
	ID  string
	Map *orderedmap.OrderedMap[string, *Example]

	// To support !include of NamedExample fragment
	Link *NamedExample

	Location string
	stacktrace.Position
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
	// Inherit is a ShapeInheritor
	Inherit(source Shape) (Shape, error)
	ShapeBaser
	ShapeChecker
	// Clone is a ShapeCloner
	Clone() Shape
	// clone is a ShapeCloner
	clone(history []Shape) Shape
	ShapeValidator

	yamlNodesUnmarshaller
	fmt.Stringer
}

// identifyShapeType identifies the type of the shape.
func identifyShapeType(shapeFacets []*yaml.Node) string {
	var t = TypeString
	for i := 0; i != len(shapeFacets); i += 2 {
		node := shapeFacets[i]
		if _, isString := SetOfStringFacets[node.Value]; isString {
			t = TypeString
		} else if _, isInteger := SetOfNumberFacets[node.Value]; isInteger {
			t = TypeInteger
			break
		} else if _, isFile := SetOfFileFacets[node.Value]; isFile {
			t = TypeFile
			break // File has a unique facet
		} else if _, isObj := SetOfObjectFacets[node.Value]; isObj {
			t = TypeObject
			break
		} else if _, isArray := SetOfArrayFacets[node.Value]; isArray {
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
		return nil, StacktraceNewWrapped("unmarshal json", err, base.Location,
			stacktrace.WithPosition(&base.Position))
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
		return nil, StacktraceNewWrapped("unmarshal yaml nodes", err, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithInfo("shape type", shapeType))
	}

	return shape, nil
}

// MakeBaseShape creates a new base shape which is a base for all shapes.
func (r *RAML) MakeBaseShape(name string, location string, position *stacktrace.Position) *BaseShape {
	return &BaseShape{
		ID:       generateShapeID(),
		Name:     name,
		Location: location,
		Position: *position,

		raml:                        r,
		CustomDomainProperties:      orderedmap.New[string, *DomainExtension](0),
		CustomShapeFacets:           orderedmap.New[string, *Node](0),
		CustomShapeFacetDefinitions: orderedmap.New[string, Property](0),
	}
}

// TODO: Temporary workaround
var idCounter = 1

func generateShapeID() string {
	id := "#" + strconv.Itoa(idCounter)
	idCounter++
	return id
}

func (r *RAML) makeShapeType(
	shapeTypeNode *yaml.Node, shapeFacets []*yaml.Node,
	name, location string, base *BaseShape) (string, *Shape, error) {
	var shape *Shape
	var shapeType string

	switch shapeTypeNode.Kind {
	default:
		return shapeType, shape, stacktrace.New("type must be string or array", location,
			WithNodePosition(shapeTypeNode))
	case yaml.DocumentNode:
		return shapeType, shape, stacktrace.New("document node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.MappingNode:
		return shapeType, shape, stacktrace.New("mapping node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.AliasNode:
		return shapeType, shape, stacktrace.New("alias node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.ScalarNode:
		switch shapeTypeNode.Tag {
		case TagStr:
			shapeType = shapeTypeNode.Value
			if shapeType == "" {
				shapeType = identifyShapeType(shapeFacets)
			} else if shapeType[0] == '{' {
				s, errMake := r.MakeJSONShape(base, shapeType)
				if errMake != nil {
					return "", nil, StacktraceNewWrapped("make json shape", errMake, location,
						WithNodePosition(shapeTypeNode))
				}
				return shapeType, &s, nil
			}
		case TagInclude:
			baseDir := filepath.Dir(location)
			dt, errParse := r.parseDataType(filepath.Join(baseDir, shapeTypeNode.Value))
			if errParse != nil {
				return shapeType, shape, StacktraceNewWrapped("parse data", errParse, location,
					WithNodePosition(shapeTypeNode))
			}
			base.TypeLabel = shapeTypeNode.Value
			base.Link = dt
		case TagNull:
			shapeType = TypeString
		default:
			return shapeType, shape, stacktrace.New("type must be string", location,
				WithNodePosition(shapeTypeNode))
		}
	case yaml.SequenceNode:
		var inherits = make([]*Shape, len(shapeTypeNode.Content))
		for i, node := range shapeTypeNode.Content {
			if node.Kind != yaml.ScalarNode {
				return shapeType, shape, stacktrace.New("node kind must be scalar", location,
					WithNodePosition(node))
			} else if node.Tag == "!include" {
				return shapeType, shape, stacktrace.New("!include is not allowed in multiple inheritance",
					location, WithNodePosition(node))
			}
			s, errMake := r.makeShape(node, name, location)
			if errMake != nil {
				return shapeType, shape, StacktraceNewWrapped("make shape", errMake, location,
					WithNodePosition(node))
			}
			inherits[i] = s
		}
		base.Inherits = inherits
		shapeType = TypeComposite
	}
	return shapeType, shape, nil
}

// makeShape creates a new shape from the given YAML node.
func (r *RAML) makeShape(v *yaml.Node, name string, location string) (*Shape, error) {
	base := r.MakeBaseShape(name, location, &stacktrace.Position{Line: v.Line, Column: v.Column})

	shapeTypeNode, shapeFacets, err := base.decode(v)
	if err != nil {
		return nil, StacktraceNewWrapped("decode", err, location, WithNodePosition(v))
	}

	var shapeType string
	if shapeTypeNode == nil {
		shapeType = identifyShapeType(shapeFacets)
	} else {
		var shape *Shape
		shapeType, shape, err = r.makeShapeType(shapeTypeNode, shapeFacets, name, location, base)
		if err != nil {
			return nil, fmt.Errorf("make shape type: %w", err)
		}
		if shape != nil {
			return shape, nil
		}
	}

	s, err := r.MakeConcreteShape(base, shapeType, shapeFacets)
	if err != nil {
		return nil, StacktraceNewWrapped("make concrete shape", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}
	ptr := &s
	if _, ok := s.(*UnknownShape); ok {
		r.unresolvedShapes.PushBack(ptr)
	}
	return ptr, nil
}

func (s *BaseShape) decodeExamples(valueNode *yaml.Node) error {
	if s.Example != nil {
		return stacktrace.New("example and examples cannot be defined together", s.Location,
			WithNodePosition(valueNode))
	}
	if valueNode.Kind == yaml.ScalarNode && valueNode.Tag == "!include" {
		baseDir := filepath.Dir(s.Location)
		n, err := s.raml.parseNamedExample(filepath.Join(baseDir, valueNode.Value))
		if err != nil {
			return StacktraceNewWrapped("parse named example", err, s.Location,
				WithNodePosition(valueNode))
		}
		s.Examples = &Examples{Link: n, Location: s.Location}
		return nil
	} else if valueNode.Kind != yaml.MappingNode {
		return stacktrace.New("examples must be map", s.Location,
			WithNodePosition(valueNode))
	}
	examples := orderedmap.New[string, *Example](len(valueNode.Content) / 2)
	for j := 0; j != len(valueNode.Content); j += 2 {
		name := valueNode.Content[j].Value
		data := valueNode.Content[j+1]
		example, err := s.raml.makeExample(data, name, s.Location)
		if err != nil {
			return StacktraceNewWrapped(fmt.Sprintf("make examples: [%d]", j),
				err, s.Location, WithNodePosition(data))
		}
		examples.Set(name, example)
	}
	s.Examples = &Examples{Map: examples, Location: s.Location}
	return nil
}

func (s *BaseShape) decodeFacets(valueNode *yaml.Node) error {
	s.CustomShapeFacetDefinitions = orderedmap.New[string, Property](len(valueNode.Content) / 2)
	for j := 0; j != len(valueNode.Content); j += 2 {
		nodeName := valueNode.Content[j].Value
		data := valueNode.Content[j+1]

		propertyName, hasImplicitOptional := s.raml.chompImplicitOptional(nodeName)
		property, err := s.raml.makeProperty(nodeName, propertyName, data, s.Location, hasImplicitOptional)
		if err != nil {
			return StacktraceNewWrapped("make property", err, s.Location,
				WithNodePosition(data))
		}
		s.CustomShapeFacetDefinitions.Set(property.Name, property)
	}
	return nil
}

func (s *BaseShape) decodeExample(valueNode *yaml.Node) error {
	if s.Examples != nil {
		return stacktrace.New("example and examples cannot be defined together", s.Location,
			WithNodePosition(valueNode))
	}
	example, err := s.raml.makeExample(valueNode, "", s.Location)
	if err != nil {
		return StacktraceNewWrapped("make example", err, s.Location,
			WithNodePosition(valueNode))
	}
	s.Example = example
	return nil
}

func (s *BaseShape) decodeValueNode(node, valueNode *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	var shapeTypeNode *yaml.Node
	shapeFacets := make([]*yaml.Node, 0)

	switch node.Value {
	case "type":
		shapeTypeNode = valueNode
	case "displayName":
		if err := valueNode.Decode(&s.DisplayName); err != nil {
			return nil, nil, StacktraceNewWrapped("decode display name", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "description":
		if err := valueNode.Decode(&s.Description); err != nil {
			return nil, nil, StacktraceNewWrapped("decode description", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "required":
		if err := valueNode.Decode(&s.Required); err != nil {
			return nil, nil, StacktraceNewWrapped("decode required", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "facets":
		if err := s.decodeFacets(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode facets", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "example":
		if err := s.decodeExample(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode example", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "examples":
		if err := s.decodeExamples(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode example", err, s.Location,
				WithNodePosition(valueNode))
		}
	case "default":
		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return nil, nil, StacktraceNewWrapped("make node default", err, s.Location,
				WithNodePosition(valueNode))
		}
		s.Default = n
	case "allowedTargets":
		// TODO: Included by annotationTypes
	default:
		if IsCustomDomainExtensionNode(node.Value) {
			name, de, err := s.raml.unmarshalCustomDomainExtension(s.Location, node, valueNode)
			if err != nil {
				return nil, nil, StacktraceNewWrapped("unmarshal custom domain extension", err, s.Location,
					WithNodePosition(valueNode))
			}
			s.CustomDomainProperties.Set(name, de)
		} else {
			shapeFacets = append(shapeFacets, node, valueNode)
		}
	}
	return shapeTypeNode, shapeFacets, nil
}

// decode decodes the shape from the YAML node.
func (s *BaseShape) decode(value *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	// For inline type declaration
	if value.Kind == yaml.ScalarNode || value.Kind == yaml.SequenceNode {
		return value, nil, nil
	}

	if value.Kind != yaml.MappingNode {
		return nil, nil, stacktrace.New("value kind must be map", s.Location, WithNodePosition(value))
	}

	var shapeTypeNode *yaml.Node
	shapeFacets := make([]*yaml.Node, 0)

	for i := 0; i != len(value.Content); i += 2 {
		node := value.Content[i]
		valueNode := value.Content[i+1]
		t, f, err := s.decodeValueNode(node, valueNode)
		if err != nil {
			return nil, nil, fmt.Errorf("decode value node: %w", err)
		}
		if t != nil {
			shapeTypeNode = t
		}
		if len(f) > 0 {
			shapeFacets = append(shapeFacets, f...)
		}
	}

	return shapeTypeNode, shapeFacets, nil
}
