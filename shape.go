package raml

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync/atomic"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xeipuuv/gojsonschema"
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

// type ShapeGetter interface {
// 	Shape() Shape
// }

type ShapeSetter interface {
	SetShape(Shape)
}

type BaseShape struct {
	Shape
	ID          int64
	Name        string
	DisplayName *string
	Description *string
	// TODO: Move Type to underlying Shape
	Type      string
	TypeLabel string // Used to store the label either the link or type value
	Example   *Example
	Examples  *Examples
	Inherits  []*BaseShape
	Alias     *BaseShape
	Default   *Node
	Required  *bool

	// To support !include of DataType fragment
	Link *DataTypeFragment

	// CustomShapeFacets is a map of custom facets with values
	CustomShapeFacets *orderedmap.OrderedMap[string, *Node]
	// CustomShapeFacetDefinitions is an object properties share the same syntax with custom shape facets.
	CustomShapeFacetDefinitions *orderedmap.OrderedMap[string, Property]
	// CustomDomainProperties is a map of custom annotations
	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	// Controlled by UnwrapShape
	unwrapped bool
	// NOTE: Not thread safe and should be used only in one method simultaneously.
	ShapeVisited bool

	raml *RAML

	Location string
	stacktrace.Position
}

func (s *BaseShape) callRAMLHooks(key HookKey, params ...any) error {
	if s.raml == nil {
		return nil
	}
	params = append([]any{s}, params...)
	return s.raml.callHooks(key, params...)
}

func (s *BaseShape) AppendRAMLHook(key HookKey, hook HookFunc) {
	if s.raml == nil {
		s.raml = New(context.Background())
	}
	s.raml.AppendHook(key, hook)
}

func (s *BaseShape) RemoveRAMLHook(key HookKey, hook HookFunc) {
	if s.raml == nil {
		return
	}
	s.raml.RemoveHook(key, hook)
}

func (s *BaseShape) PrepenRAMLHook(key HookKey, hook HookFunc) {
	if s.raml == nil {
		s.raml = New(context.Background())
	}
	s.raml.PrependHook(key, hook)
}

func (s *BaseShape) ClearRAMLHooks(key HookKey) {
	if s.raml == nil {
		return
	}
	s.raml.ClearHooks(key)
}

func (s *BaseShape) SetShape(shape Shape) {
	s.Shape = shape
}

func (s *BaseShape) Validate(v interface{}) error {
	return s.Shape.validate(v, "$")
}

const HookBeforeBaseShapeInherit = "BaseShape.Inherit"

func (s *BaseShape) Inherit(sourceBase *BaseShape) (*BaseShape, error) {
	if err := s.callRAMLHooks(HookBeforeBaseShapeInherit, sourceBase); err != nil {
		return nil, err
	}

	// Avoid recursion caused by inheritance chain
	if sourceBase.ShapeVisited {
		// NOTE: We do not mark any recursions here. External code must handle this case.
		return sourceBase, nil
	}
	sourceBase.ShapeVisited = true

	source := sourceBase.Shape
	target := s.Shape

	if s.Description == nil {
		s.Description = sourceBase.Description
	}

	// Inherit custom shape facets
	if s.CustomShapeFacets == nil {
		s.CustomShapeFacets = sourceBase.CustomShapeFacets
	} else if sourceBase.CustomShapeFacets != nil {
		for pair := sourceBase.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
			k, sourceNode := pair.Key, pair.Value
			if _, present := s.CustomShapeFacets.Get(k); !present {
				s.CustomShapeFacets.Set(k, sourceNode)
			}
		}
	}

	// If source type is any, return target as is
	if _, ok := source.(*AnyShape); ok {
		return s, nil
	}

	sourceUnion, isSourceUnion := source.(*UnionShape)
	targetUnion, isTargetUnion := target.(*UnionShape)

	switch {
	case isSourceUnion && !isTargetUnion:
		return s.inheritUnionSource(sourceUnion)

	case isTargetUnion && !isSourceUnion:
		return s.inheritUnionTarget(targetUnion)
	}
	// Homogenous types produce same type
	_, err := target.inherit(source)
	if err != nil {
		return nil, StacktraceNewWrapped("merge shapes", err, target.Base().Location,
			stacktrace.WithPosition(&target.Base().Position))
	}
	sourceBase.ShapeVisited = false
	return s, nil
}

const HookBeforeBaseShapeInheritUnionSource = "BaseShape.inheritUnionSource"

func (s *BaseShape) inheritUnionSource(sourceUnion *UnionShape) (*BaseShape, error) {
	if err := s.callRAMLHooks(HookBeforeBaseShapeInheritUnionSource, sourceUnion); err != nil {
		return nil, err
	}
	var filtered []*BaseShape
	var st *stacktrace.StackTrace
	for _, source := range sourceUnion.AnyOf {
		// If at least one union member has any type, the whole union is considered as any type.
		if _, ok := source.Shape.(*AnyShape); ok {
			return s, nil
		}
		if source.Type == s.Type {
			// Deep copy with ID change is required since we create new union members from source members
			tc := s.CloneDetached()
			tc.ID = s.raml.generateShapeID()
			is, err := tc.Inherit(source)
			if err != nil {
				se := StacktraceNewWrapped("merge shapes", err, s.Location,
					stacktrace.WithPosition(&s.Position))
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
				// Skip shapes that didn't pass inheritance check
				continue
			}
			filtered = append(filtered, is)
		}
	}
	if len(filtered) == 0 {
		se := StacktraceNew("failed to find compatible union member", s.Location,
			stacktrace.WithPosition(&s.Position))
		if st != nil {
			se = se.Append(st)
		}
		return nil, se
	}
	// If only one union member remains - simplify to target type
	if len(filtered) == 1 {
		return filtered[0], nil
	}
	// Convert target to union
	s.Type = TypeUnion
	s.SetShape(&UnionShape{
		BaseShape: s,
		UnionFacets: UnionFacets{
			AnyOf: filtered,
		},
	})
	return s, nil
}

const HookBeforeBaseShapeInheritUnionTarget = "BaseShape.inheritUnionTarget"

func (s *BaseShape) inheritUnionTarget(targetUnion *UnionShape) (*BaseShape, error) {
	if err := s.callRAMLHooks(HookBeforeBaseShapeInheritUnionTarget, targetUnion); err != nil {
		return nil, err
	}
	var st *stacktrace.StackTrace
	for _, item := range targetUnion.AnyOf {
		// Merge will raise an error in case any of union members has incompatible type
		_, err := item.Inherit(s)
		if err != nil {
			se := StacktraceNewWrapped("merge shapes", err, targetUnion.Base().Location,
				stacktrace.WithPosition(&targetUnion.Base().Position))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
	}
	if st != nil {
		return nil, st
	}
	return targetUnion.Base(), nil
}

func (s *BaseShape) AliasTo(source *BaseShape) (*BaseShape, error) {
	_, err := s.Shape.alias(source.Shape)
	if err != nil {
		return nil, StacktraceNewWrapped("alias shape", err, s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	s.DisplayName = source.DisplayName
	s.Description = source.Description
	s.Example = source.Example
	s.Examples = source.Examples
	s.Inherits = source.Inherits
	s.Alias = source.Alias
	s.Default = source.Default
	s.Required = source.Required
	s.CustomShapeFacets = source.CustomShapeFacets
	s.CustomShapeFacetDefinitions = source.CustomShapeFacetDefinitions
	s.CustomDomainProperties = source.CustomDomainProperties
	return s, nil
}

// CloneShallow creates a shallow copy of the shape.
func (s *BaseShape) CloneShallow() *BaseShape {
	c := *s
	ptr := &c

	c.CustomDomainProperties = orderedmap.New[string, *DomainExtension](s.CustomDomainProperties.Len())
	for pair := s.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		c.CustomDomainProperties.Set(pair.Key, pair.Value)
	}

	c.CustomShapeFacets = orderedmap.New[string, *Node](s.CustomShapeFacets.Len())
	for pair := s.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		c.CustomShapeFacets.Set(pair.Key, pair.Value)
	}

	c.CustomShapeFacetDefinitions = orderedmap.New[string, Property](s.CustomShapeFacetDefinitions.Len())
	for pair := s.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		c.CustomShapeFacetDefinitions.Set(pair.Key, prop)
	}

	c.Shape = s.Shape.cloneShallow(ptr)
	return ptr
}

// Clone creates a deep copy of the shape.
//
// Use this method to make a deep copy of the shape while preserving the relationships between shapes.
// Passed cloned map will be populated with cloned shape IDs that can be reused in subsequent Clone calls.
//
// NOTE: If you need a completely independent copy of a shape, use CloneDetached method.
func (s *BaseShape) Clone(clonedMap map[int64]*BaseShape) *BaseShape {
	return s.clone(clonedMap)
}

// CloneDetached creates a detached deep copy of the shape.
//
// Detached copy makes a deep copy of the shape, including parents, links and aliases.
// This makes the copied shape and its references completely independent from the original tree.
//
// NOTE: To avoid excessive memory copies and allocation, this method must be used only
// when an independent shape copy is required. Otherwise, use Clone method.
func (s *BaseShape) CloneDetached() *BaseShape {
	return s.clone(make(map[int64]*BaseShape))
}

func (s *BaseShape) clone(clonedMap map[int64]*BaseShape) *BaseShape {
	if shape, ok := clonedMap[s.ID]; ok {
		return shape
	}

	c := *s
	clonedMap[s.ID] = &c

	// TODO: Node is not deep copied yet, but it's not mutated anyway
	c.CustomShapeFacets = orderedmap.New[string, *Node](s.CustomShapeFacets.Len())
	for pair := s.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		c.CustomShapeFacets.Set(pair.Key, pair.Value)
	}

	c.CustomShapeFacetDefinitions = orderedmap.New[string, Property](s.CustomShapeFacetDefinitions.Len())
	for pair := s.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		prop.Base = prop.Base.clone(clonedMap)
		c.CustomShapeFacetDefinitions.Set(pair.Key, prop)
	}

	// TODO: DomainExtension is not deep copied yet, but it's not mutated anyway
	c.CustomDomainProperties = orderedmap.New[string, *DomainExtension](s.CustomDomainProperties.Len())
	for pair := s.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		c.CustomDomainProperties.Set(pair.Key, pair.Value)
	}

	if s.Alias != nil {
		c.Alias = s.Alias.clone(clonedMap)
	}
	if s.Inherits != nil {
		c.Inherits = make([]*BaseShape, len(s.Inherits))
		for i, v := range s.Inherits {
			c.Inherits[i] = v.clone(clonedMap)
		}
	}
	if s.Link != nil {
		l := *s.Link
		c.Link = &l
		c.Link.Shape = s.Link.Shape.clone(clonedMap)
	}
	c.Shape = s.Shape.clone(&c, clonedMap)

	return &c
}

// Check returns an error if type shape is invalid.
func (s *BaseShape) Check() error {
	return s.Shape.check()
}

// IsUnwrapped returns true if the shape is unwrapped.
func (s *BaseShape) IsUnwrapped() bool {
	return s.unwrapped
}

func (s *BaseShape) SetUnwrapped() {
	s.unwrapped = true
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
			r = fmt.Sprintf("%s: %s", r, i.Name)
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
	validate(v interface{}, ctxPath string) error
}

// ShapeInheritor is the interface that represents an inheritor of a RAML shape.
type ShapeInheritor interface {
	inherit(source Shape) (Shape, error)
}

// ShapeCloner is the interface that provide clone implementation for a RAML shape.
type ShapeCloner interface {
	clone(base *BaseShape, clonedMap map[int64]*BaseShape) Shape
	cloneShallow(base *BaseShape) Shape
}

// ShapeAliaser is the interface that provides alias implementation for a RAML shape.
type ShapeAliaser interface {
	alias(source Shape) (Shape, error)
}

type ShapeChecker interface {
	check() error
}

// yamlNodesUnmarshaller is the interface that represents an unmarshaller of a RAML shape from YAML nodes.
type yamlNodesUnmarshaller interface {
	unmarshalYAMLNodes(v []*yaml.Node) error
}

// Shape is the interface that represents a RAML shape.
type Shape interface {
	// Inherit is a ShapeInheritor
	ShapeInheritor
	ShapeBaser
	ShapeChecker
	// ShapeCloner Clones the shape and its children and points to specified base shape.
	ShapeCloner
	ShapeAliaser
	ShapeValidator

	yamlNodesUnmarshaller
	fmt.Stringer
	IsScalar() bool
}

// identifyShapeType identifies the type of the shape by facets.
func identifyShapeType(shapeFacets []*yaml.Node) (string, error) {
	var t = ""
	var stringOnly bool
	for i := 0; i != len(shapeFacets); i += 2 {
		node := shapeFacets[i]
		var ft string
		var ok bool
		switch node.Value {
		case FacetMinLength, FacetMaxLength:
			ok = true
			ft = TypeString
		case FacetPattern:
			ok = true
			ft = TypeString
			stringOnly = true
		case FacetMinimum, FacetMaximum, FacetMultipleOf:
			ok = true
			ft = TypeNumber
		case FacetMinItems, FacetMaxItems, FacetUniqueItems, FacetItems:
			ok = true
			ft = TypeArray
		case FacetMinProperties, FacetMaxProperties, FacetAdditionalProperties, FacetProperties, FacetDiscriminator:
			ok = true
			ft = TypeObject
		case FacetFileTypes:
			ok = true
			ft = TypeFile
		}
		if ok {
			switch t {
			case TypeString:
				if ft == TypeFile && !stringOnly {
					t = TypeFile
					ft = TypeFile
				}
			case TypeFile:
				if ft == TypeString && !stringOnly {
					ft = TypeFile
					t = TypeFile
				}
			}
			if t != "" && ft != t {
				return "", fmt.Errorf("detected types by facets are not equal: %s and %s", t, ft)
			}
			t = ft
		}
	}
	if t == "" {
		t = TypeString
	}
	return t, nil
}

func (r *RAML) MakeRecursiveShape(headBase *BaseShape) *BaseShape {
	recursiveBase := r.MakeBaseShape(headBase.Name, headBase.Location, headBase.Position)
	recursiveBase.Name = headBase.Name
	recursiveBase.Type = TypeRecursive
	recursiveBase.Description = headBase.Description
	recursiveBase.CustomDomainProperties = headBase.CustomDomainProperties
	recursiveBase.CustomShapeFacets = headBase.CustomShapeFacets
	recursiveBase.CustomShapeFacetDefinitions = headBase.CustomShapeFacetDefinitions
	s := &RecursiveShape{BaseShape: recursiveBase, Head: headBase}
	recursiveBase.SetShape(s)
	return recursiveBase
}

func (r *RAML) MakeJSONShape(base *BaseShape, rawSchema string) (*JSONShape, error) {
	base.Type = "json"

	// TODO: Probably this can be replaced with gojsonschema but it does not expose internal schema structure.
	var schema *JSONSchema
	err := json.Unmarshal([]byte(rawSchema), &schema)
	if err != nil {
		return nil, StacktraceNewWrapped("unmarshal json", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}

	// TODO: This will only work with local files, but currently we work only with local files anyway
	p := "file://" + filepath.ToSlash(base.Location)
	// Load schema using string loader
	l := gojsonschema.NewStringLoader(rawSchema)
	sl := gojsonschema.NewSchemaLoader()
	// Add it to schema loader with URI pointing to RAML file.
	// This will cache the schema and resolve all references against this URI.
	err = sl.AddSchema(p, l)
	if err != nil {
		return nil, StacktraceNewWrapped("add schema", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}
	// Replace StringLoader with ReferenceLoader to support local/remote references resolution.
	// Since the reference is cached, gojsonschema will not attempt to load it from the storage.
	// TODO: Introduce custom reference loader to have possibility to disallow remote schemas.
	l = gojsonschema.NewReferenceLoader(p)
	validator, err := sl.Compile(l)
	if err != nil {
		return nil, StacktraceNewWrapped("new schema", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}

	return &JSONShape{BaseShape: base, Raw: rawSchema, Schema: schema, Validator: validator}, nil
}

const HookBeforeRAMLMakeConcreteShapeYAML = "before:RAML.makeConcreteShapeYAML"

// MakeConcreteShapeYAML creates a new concrete shape and assigns it to the base shape.
func (r *RAML) MakeConcreteShapeYAML(base *BaseShape, shapeType string, shapeFacets []*yaml.Node) (Shape, error) {
	if err := r.callHooks(HookBeforeRAMLMakeConcreteShapeYAML, base, shapeType, shapeFacets); err != nil {
		return nil, err
	}

	base.Type = shapeType

	// NOTE: Shape resolution is performed in a separate stage.
	var shape Shape
	switch shapeType {
	default:
		// NOTE: UnknownShape is a special type of shape that will be resolved later.
		shape = &UnknownShape{BaseShape: base}
	case TypeAny:
		shape = &AnyShape{BaseShape: base}
	case TypeNil:
		shape = &NilShape{BaseShape: base}
	case TypeObject:
		shape = &ObjectShape{BaseShape: base}
	case TypeArray:
		shape = &ArrayShape{BaseShape: base}
	case TypeString:
		shape = &StringShape{BaseShape: base}
	case TypeInteger:
		shape = &IntegerShape{BaseShape: base}
	case TypeNumber:
		shape = &NumberShape{BaseShape: base}
	case TypeDatetime:
		shape = &DateTimeShape{BaseShape: base}
	case TypeDatetimeOnly:
		shape = &DateTimeOnlyShape{BaseShape: base}
	case TypeDateOnly:
		shape = &DateOnlyShape{BaseShape: base}
	case TypeTimeOnly:
		shape = &TimeOnlyShape{BaseShape: base}
	case TypeFile:
		shape = &FileShape{BaseShape: base}
	case TypeBoolean:
		shape = &BooleanShape{BaseShape: base}
	case TypeUnion:
		shape = &UnionShape{BaseShape: base}
	case TypeJSON:
		shape = &JSONShape{BaseShape: base}
	}
	base.SetShape(shape)

	if err := shape.unmarshalYAMLNodes(shapeFacets); err != nil {
		return nil, StacktraceNewWrapped("unmarshal yaml nodes", err, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithInfo("shape type", shapeType))
	}

	return shape, nil
}

// MakeBaseShape creates a new base shape which is a base for all shapes.
func (r *RAML) MakeBaseShape(name string, location string, position stacktrace.Position) *BaseShape {
	// If position is not set, use default position.
	if position.Line == 0 && position.Column == 0 {
		position.Line = 1
	}
	b := &BaseShape{
		ID:       r.generateShapeID(),
		Name:     name,
		Location: location,
		Position: position,

		raml:                        r,
		CustomDomainProperties:      orderedmap.New[string, *DomainExtension](0),
		CustomShapeFacets:           orderedmap.New[string, *Node](0),
		CustomShapeFacetDefinitions: orderedmap.New[string, Property](0),
	}
	r.PutShape(b)
	return b
}

func (r *RAML) generateShapeID() int64 {
	return atomic.AddInt64(&r.idCounter, 1)
}

func (r *RAML) makeShapeType(
	shapeTypeNode *yaml.Node,
	shapeFacets []*yaml.Node,
	name string,
	location string,
	base *BaseShape,
) (string, Shape, error) {
	var shapeType string
	switch shapeTypeNode.Kind {
	default:
		return "", nil, StacktraceNew("type must be string or array", location,
			WithNodePosition(shapeTypeNode))
	case yaml.DocumentNode:
		return "", nil, StacktraceNew("document node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.MappingNode:
		return "", nil, StacktraceNew("mapping node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.AliasNode:
		return "", nil, StacktraceNew("alias node is not allowed", location,
			WithNodePosition(shapeTypeNode))
	case yaml.ScalarNode:
		switch shapeTypeNode.Tag {
		case TagStr:
			shapeType = shapeTypeNode.Value
			if shapeType == "" {
				shapeTypeI, err := identifyShapeType(shapeFacets)
				if err != nil {
					return "", nil, StacktraceNewWrapped("identify shape type", err, location,
						WithNodePosition(shapeTypeNode))
				}
				shapeType = shapeTypeI
			} else if shapeType[0] == '{' {
				s, errMake := r.MakeJSONShape(base, shapeType)
				if errMake != nil {
					return "", nil, StacktraceNewWrapped("make json shape", errMake, location,
						WithNodePosition(shapeTypeNode))
				}
				return shapeType, s, nil
			}
		case TagInclude:
			baseDir := filepath.Dir(location)
			dt, errParse := r.parseDataType(filepath.Join(baseDir, shapeTypeNode.Value))
			if errParse != nil {
				return "", nil, StacktraceNewWrapped("parse data", errParse, location,
					WithNodePosition(shapeTypeNode))
			}
			base.TypeLabel = shapeTypeNode.Value
			base.Link = dt
		case TagNull:
			shapeType = TypeString
		default:
			return "", nil, StacktraceNew("type must be string", location,
				WithNodePosition(shapeTypeNode))
		}
	case yaml.SequenceNode:
		var inherits = make([]*BaseShape, len(shapeTypeNode.Content))
		for i, node := range shapeTypeNode.Content {
			if node.Kind != yaml.ScalarNode {
				return "", nil, StacktraceNew("node kind must be scalar", location,
					WithNodePosition(node))
			} else if node.Tag == TagInclude {
				return "", nil, StacktraceNew("!include is not allowed in multiple inheritance",
					location, WithNodePosition(node))
			}
			s, errMake := r.makeNewShapeYAML(node, name, location)
			if errMake != nil {
				return "", nil, StacktraceNewWrapped("make shape", errMake, location,
					WithNodePosition(node))
			}
			inherits[i] = s
		}
		base.Inherits = inherits
		shapeType = TypeComposite
	}
	return shapeType, nil, nil
}

func (r *RAML) MakeNewShape(
	name string,
	shapeType string,
	location string,
	position stacktrace.Position,
) (*BaseShape, Shape, error) {
	base := r.MakeBaseShape(name, location, position)
	s, err := r.MakeConcreteShapeYAML(base, shapeType, nil)
	if err != nil {
		return nil, nil, StacktraceNewWrapped("make concrete shape", err, location,
			stacktrace.WithPosition(&base.Position))
	}
	return base, s, nil
}

const HookBeforeRAMLMakeNewShapeYAML = "before:RAML.makeNewShapeYAML"

// makeNewShapeYAML creates a new shape from the given YAML node.
func (r *RAML) makeNewShapeYAML(v *yaml.Node, name string, location string) (*BaseShape, error) {
	if err := r.callHooks(HookBeforeRAMLMakeNewShapeYAML, v); err != nil {
		return nil, err
	}

	base := r.MakeBaseShape(name, location, stacktrace.Position{Line: v.Line, Column: v.Column})

	shapeTypeNode, shapeFacets, err := base.decode(v)
	if err != nil {
		return nil, StacktraceNewWrapped("decode", err, location, WithNodePosition(v))
	}

	var shapeType string
	if shapeTypeNode == nil {
		shapeType, err = identifyShapeType(shapeFacets)
		if err != nil {
			return nil, StacktraceNewWrapped("identify shape type", err, location,
				WithNodePosition(v))
		}
	} else {
		var shape Shape
		shapeType, shape, err = r.makeShapeType(shapeTypeNode, shapeFacets, name, location, base)
		if err != nil {
			return nil, fmt.Errorf("make shape type: %w", err)
		}
		if shape != nil {
			base.SetShape(shape)
			return base, nil
		}
	}

	s, err := r.MakeConcreteShapeYAML(base, shapeType, shapeFacets)
	if err != nil {
		return nil, StacktraceNewWrapped("make concrete shape", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}
	if _, ok := s.(*UnknownShape); ok {
		r.unresolvedShapes.PushBack(base)
	}
	return base, nil
}

func (s *BaseShape) decodeExamples(valueNode *yaml.Node) error {
	if s.Example != nil {
		return StacktraceNew("example and examples cannot be defined together", s.Location,
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
		return StacktraceNew("examples must be map", s.Location,
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

// decodeFacets decodes the facet: "facets" from the YAML node.
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
		return StacktraceNew("example and examples cannot be defined together", s.Location,
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

func (s *BaseShape) decodeValueNode(keyNode, valueNode *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	var shapeTypeNode *yaml.Node
	shapeFacets := make([]*yaml.Node, 0)

	switch keyNode.Value {
	case FacetType:
		shapeTypeNode = valueNode
	case FacetDisplayName:
		if err := valueNode.Decode(&s.DisplayName); err != nil {
			return nil, nil, StacktraceNewWrapped("decode display name", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetDescription:
		if err := valueNode.Decode(&s.Description); err != nil {
			return nil, nil, StacktraceNewWrapped("decode description", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetRequired:
		if err := valueNode.Decode(&s.Required); err != nil {
			return nil, nil, StacktraceNewWrapped("decode required", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetFacets:
		if err := s.decodeFacets(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode facets", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetExample:
		if err := s.decodeExample(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode example", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetExamples:
		if err := s.decodeExamples(valueNode); err != nil {
			return nil, nil, StacktraceNewWrapped("decode example", err, s.Location,
				WithNodePosition(valueNode))
		}
	case FacetDefault:
		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return nil, nil, StacktraceNewWrapped("make node default", err, s.Location,
				WithNodePosition(valueNode))
		}
		s.Default = n
	case FacetAllowedTargets:
		// if err := valueNode.Decode(&s.AllowedTargets); err != nil {
		// 	return nil, nil, StacktraceNewWrapped("decode allowed targets", err, s.Location,
		// 		WithNodePosition(valueNode))
		// }
	default:
		if IsCustomDomainExtensionNode(keyNode.Value) {
			name, de, err := s.raml.unmarshalCustomDomainExtension(s.Location, keyNode, valueNode)
			if err != nil {
				return nil, nil, StacktraceNewWrapped("unmarshal custom domain extension", err, s.Location,
					WithNodePosition(valueNode))
			}
			s.CustomDomainProperties.Set(name, de)
		} else {
			shapeFacets = append(shapeFacets, keyNode, valueNode)
		}
	}
	return shapeTypeNode, shapeFacets, nil
}

// decode decodes the shape from the YAML node.
// It returns the shape type node, facets and an error if any.
func (s *BaseShape) decode(value *yaml.Node) (*yaml.Node, []*yaml.Node, error) {
	// For inline type declaration
	if value.Kind == yaml.ScalarNode || value.Kind == yaml.SequenceNode {
		return value, nil, nil
	}

	if value.Kind != yaml.MappingNode {
		return nil, nil, StacktraceNew("value kind must be map", s.Location, WithNodePosition(value))
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
