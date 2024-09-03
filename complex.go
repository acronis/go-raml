package raml

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// ArrayFacets contains constraints for array shapes.
type ArrayFacets struct {
	Items       *Shape
	MinItems    *uint64
	MaxItems    *uint64
	UniqueItems *bool
}

// ArrayShape represents an array shape.
type ArrayShape struct {
	BaseShape

	ArrayFacets
}

// Base returns the base shape.
func (s *ArrayShape) Base() *BaseShape {
	return &s.BaseShape
}

// func (s *ArrayShape) Validate(v interface{}) error {
// 	a, ok := v.([]interface{})
// 	if !ok {
// 		return NewError("value is not an array", s.Location)
// 	}
// 	if s.MinItems > uint64(len(a)) {
// 		return NewError("minItems constraint violation", s.Location)
// 	}
// 	if s.MaxItems < uint64(len(a)) {
// 		return NewError("maxItems constraint violation", s.Location)
// 	}
// 	if s.UniqueItems {
// 		seen := make(map[interface{}]struct{})
// 		for _, item := range a {
// 			if _, ok := seen[item]; ok {
// 				return NewError("uniqueItems constraint violation", s.Location)
// 			}
// 			seen[item] = struct{}{}
// 		}
// 	}
// 	for _, item := range a {
// 		if err := (*s.Items).Validate(item); err != nil {
// 			return NewWrappedError("validate item", err, s.Location)
// 		}
// 	}
// 	return nil
// }

// Clone returns a clone of the shape.
func (s *ArrayShape) Clone() Shape {
	return s.clone(make([]Shape, 0))
}

func (s *ArrayShape) clone(history []Shape) Shape {
	for _, item := range history {
		if item.Base().Id == s.Id {
			return item
		}
	}
	c := *s
	ptr := &c
	history = append(history, ptr)
	if c.Items != nil {
		items := (*c.Items).clone(history)
		c.Items = &items
	}
	return ptr
}

// Inherit merges the source shape into the target shape.
func (s *ArrayShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*ArrayShape)
	if !ok {
		return nil, NewError("cannot inherit from different type", s.Location, WithPosition(&s.Position), WithInfo("source", source.Base().Type), WithInfo("target", s.Base().Type))
	}
	if s.Items == nil {
		s.Items = ss.Items
	} else if ss.Items != nil {
		_, err := s.raml.Inherit(*s.Items, *ss.Items)
		if err != nil {
			return nil, NewWrappedError("merge array items", err, s.Location)
		}
	}
	if s.MinItems == nil {
		s.MinItems = ss.MinItems
	} else if ss.MinItems != nil && *s.MinItems > *ss.MinItems {
		return nil, NewError("minItems constraint violation", s.Location, WithPosition(&s.Position), WithInfo("source", *ss.MinItems), WithInfo("target", *s.MinItems))
	}
	if s.MaxItems == nil {
		s.MaxItems = ss.MaxItems
	} else if ss.MaxItems != nil && *s.MaxItems < *ss.MaxItems {
		return nil, NewError("maxItems constraint violation", s.Location, WithPosition(&s.Position), WithInfo("source", *ss.MaxItems), WithInfo("target", *s.MaxItems))
	}
	if s.UniqueItems == nil {
		s.UniqueItems = ss.UniqueItems
	} else if ss.UniqueItems != nil && *ss.UniqueItems && !*s.UniqueItems {
		return nil, NewError("uniqueItems constraint violation", s.Location, WithPosition(&s.Position), WithInfo("source", *ss.UniqueItems), WithInfo("target", *s.UniqueItems))
	}
	return s, nil
}

func (s *ArrayShape) ToJSONSchema() *JSONSchema {
	schema := &JSONSchema{
		ID:       s.Id,
		Type:     "array",
		MinItems: s.MinItems,
		MaxItems: s.MaxItems,
		Extras:   make(map[string]interface{}),
	}
	if s.UniqueItems != nil {
		schema.UniqueItems = *s.UniqueItems
	}
	if s.Items != nil {
		schema.Items = (*s.Items).ToJSONSchema()
	}
	schema.WithRamlData(s.Base())
	return schema
}

func (s *ArrayShape) Check() error {
	return nil
}

// UnmarshalYAMLNodes unmarshals the array shape from YAML nodes.
func (s *ArrayShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "minItems" {
			if err := valueNode.Decode(&s.MinItems); err != nil {
				return NewWrappedError("decode minItems", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "maxItems" {
			if err := valueNode.Decode(&s.MaxItems); err != nil {
				return NewWrappedError("decode maxItems: %w", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "items" {
			name := "items"
			shape, err := s.raml.makeShape(valueNode, name, s.Location)
			if err != nil {
				return NewWrappedError("make shape", err, s.Location, WithNodePosition(valueNode))
			}
			s.Items = shape
			//s.raml.PutTypeIntoFragment(s.Name+"#items", s.Location, s.Items)
			s.raml.PutShapePtr(s.Items)
		} else if node.Value == "uniqueItems" {
			if err := valueNode.Decode(&s.UniqueItems); err != nil {
				return NewWrappedError("decode uniqueItems", err, s.Location, WithNodePosition(valueNode))
			}
		} else {
			n, err := s.raml.makeNode(valueNode, s.Location)
			if err != nil {
				return NewWrappedError("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets[node.Value] = n
		}
	}
	return nil
}

// ObjectFacets contains constraints for object shapes.
type ObjectFacets struct {
	Discriminator        *string
	DiscriminatorValue   any
	AdditionalProperties *bool
	Properties           map[string]Property
	MinProperties        *uint64
	MaxProperties        *uint64
}

// ObjectShape represents an object shape.
type ObjectShape struct {
	BaseShape

	ObjectFacets
}

// func (s *ObjectShape) Validate(v interface{}) error {
// 	m, ok := v.(map[string]interface{})
// 	if !ok {
// 		return NewError("value is not a map", s.Location)
// 	}
// 	if s.MinProperties > uint64(len(m)) {
// 		return NewError("minProperties constraint violation", s.Location)
// 	}
// 	if s.MaxProperties < uint64(len(m)) {
// 		return NewError("maxProperties constraint violation", s.Location)
// 	}
// 	for k, v := range m {
// 		if p, ok := s.Properties[k]; ok {
// 			if err := (*p.Shape).Validate(v); err != nil {
// 				return NewWrappedError("validate property", err, s.Location)
// 			}
// 		} else if !s.AdditionalProperties {
// 			return NewError("additionalProperties constraint violation", s.Location)
// 		}
// 	}
// 	return nil
// }

// UnmarshalYAMLNodes unmarshals the object shape from YAML nodes.
func (s *ObjectShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		if node.Value == "additionalProperties" {
			if err := valueNode.Decode(&s.AdditionalProperties); err != nil {
				return NewWrappedError("decode additionalProperties", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "discriminator" {
			if err := valueNode.Decode(&s.Discriminator); err != nil {
				return NewWrappedError("decode discriminator", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "discriminatorValue" {
			if err := valueNode.Decode(&s.DiscriminatorValue); err != nil {
				return NewWrappedError("decode discriminatorValue", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "minProperties" {
			if err := valueNode.Decode(&s.MinProperties); err != nil {
				return NewWrappedError("decode minProperties", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "maxProperties" {
			if err := valueNode.Decode(&s.MaxProperties); err != nil {
				return NewWrappedError("decode maxProperties", err, s.Location, WithNodePosition(valueNode))
			}
		} else if node.Value == "properties" {
			s.Properties = make(map[string]Property, len(valueNode.Content)/2)
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				property, err := s.raml.makeProperty(name, data, s.Location)
				if err != nil {
					return NewWrappedError("make property", err, s.Location, WithNodePosition(data))
				}
				s.Properties[property.Name] = property
				// s.raml.PutTypeIntoFragment(s.Name+"#"+property.Name, s.Location, property.Shape)
				s.raml.PutShapePtr(property.Shape)
			}
		} else {
			n, err := s.raml.makeNode(valueNode, s.Location)
			if err != nil {
				return NewWrappedError("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets[node.Value] = n
		}
	}
	return nil
}

// Base returns the base shape.
func (s *ObjectShape) Base() *BaseShape {
	return &s.BaseShape
}

// Clone returns a clone of the object shape.
func (s *ObjectShape) Clone() Shape {
	return s.clone(make([]Shape, 0))
}

func (s *ObjectShape) clone(history []Shape) Shape {
	for _, item := range history {
		if item.Base().Id == s.Id {
			return item
		}
	}
	c := *s
	ptr := &c
	history = append(history, ptr)
	if c.Properties != nil {
		c.Properties = make(map[string]Property, len(s.Properties))
		for k, v := range s.Properties {
			p := (*v.Shape).clone(history)
			v.Shape = &p
			c.Properties[k] = v
		}
	}
	return ptr
}

// Inherit merges the source shape into the target shape.
func (s *ObjectShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*ObjectShape)
	if !ok {
		return nil, NewError("cannot inherit from different type", s.Location, WithPosition(&s.Position), WithInfo("source", source.Base().Type), WithInfo("target", s.Base().Type))
	}

	// Discriminator and AdditionalProperties are inherited as is
	if s.AdditionalProperties == nil {
		s.AdditionalProperties = ss.AdditionalProperties
	}
	if s.Discriminator == nil {
		s.Discriminator = ss.Discriminator
	}

	if s.MinProperties == nil {
		s.MinProperties = ss.MinProperties
	} else if ss.MinProperties != nil && *s.MinProperties < *ss.MinProperties {
		return nil, NewError("minProperties constraint violation", s.Location, WithPosition(&s.Position), WithInfo("source", *ss.MinProperties), WithInfo("target", *s.MinProperties))
	}
	if s.MaxProperties == nil {
		s.MaxProperties = ss.MaxProperties
	} else if ss.MaxProperties != nil && *s.MaxProperties > *ss.MaxProperties {
		return nil, NewError("maxProperties constraint violation", s.Location, WithPosition(&s.Position), WithInfo("source", *ss.MaxProperties), WithInfo("target", *s.MaxProperties))
	}

	if s.Properties == nil {
		s.Properties = ss.Properties
	} else if ss.Properties != nil {
		for k, sourceProp := range ss.Properties {
			if targetProp, ok := s.Properties[k]; ok {
				if sourceProp.Required && !targetProp.Required {
					return nil, NewError("cannot make required property optional", s.Location, WithPosition(&(*targetProp.Shape).Base().Position), WithInfo("property", k), WithInfo("source", sourceProp.Required), WithInfo("target", targetProp.Required))
				}
				_, err := s.raml.Inherit(*sourceProp.Shape, *s.Properties[k].Shape)
				if err != nil {
					return nil, NewWrappedError("merge object property", err, s.Base().Location, WithPosition(&(*targetProp.Shape).Base().Position), WithInfo("property", k))
				}
			} else {
				s.Properties[k] = sourceProp
			}
		}
	}
	return s, nil
}

func (s *ObjectShape) ToJSONSchema() *JSONSchema {
	schema := &JSONSchema{
		ID:            s.Id,
		Type:          "object",
		MinProperties: s.MinProperties,
		MaxProperties: s.MaxProperties,
		Extras:        make(map[string]interface{}),
	}
	if s.AdditionalProperties != nil && !*s.AdditionalProperties {
		schema.AdditionalProperties = s.AdditionalProperties
	}
	if s.Properties != nil {
		schema.PatternProperties = make(map[string]*JSONSchema, len(s.Properties))
		schema.Properties = make(map[string]*JSONSchema, len(s.Properties))
		for k, v := range s.Properties {
			if k[0] == '/' && k[len(k)-1] == '/' {
				k = k[1 : len(k)-1]
				schema.PatternProperties[k] = (*v.Shape).ToJSONSchema()
			} else {
				schema.Properties[k] = (*v.Shape).ToJSONSchema()
			}
			if v.Required {
				schema.Required = append(schema.Required, k)
			}
		}
	}
	schema.WithRamlData(s.Base())
	return schema
}

func (s *ObjectShape) Check() error {
	return nil
}

// makeProperty creates a property from a YAML node.
func (r *RAML) makeProperty(name string, v *yaml.Node, location string) (Property, error) {
	shape, err := r.makeShape(v, name, location)
	if err != nil {
		return Property{}, NewWrappedError("make shape", err, location, WithNodePosition(v))
	}
	propertyName := name
	var required bool
	// Pattern properties are always optional
	if propertyName[0] == '/' && propertyName[len(propertyName)-1] == '/' {
		required = false
	} else {
		shapeRequired := (*shape).Base().Required
		if shapeRequired == nil {
			if strings.HasSuffix(propertyName, "?") {
				required = false
				propertyName = propertyName[:len(propertyName)-1]
			} else {
				required = true
			}
		} else {
			required = *shapeRequired
		}
	}
	return Property{
		Name:     propertyName,
		Shape:    shape,
		Required: required,
		raml:     r,
	}, nil
}

// Property represents a property of an object shape.
type Property struct {
	Name     string
	Shape    *Shape
	Required bool
	raml     *RAML
}

// UnionFacets contains constraints for union shapes.
type UnionFacets struct {
	AnyOf []*Shape
}

// UnionShape represents a union shape.
type UnionShape struct {
	BaseShape

	EnumFacets
	UnionFacets
}

// UnmarshalYAMLNodes unmarshals the union shape from YAML nodes.
func (s *UnionShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	return nil
}

// Base returns the base shape.
func (s *UnionShape) Base() *BaseShape {
	return &s.BaseShape
}

// Clone returns a clone of the union shape.
func (s *UnionShape) Clone() Shape {
	return s.clone(make([]Shape, 0))
}

func (s *UnionShape) clone(history []Shape) Shape {
	for _, item := range history {
		if item.Base().Id == s.Id {
			return item
		}
	}
	c := *s
	ptr := &c
	history = append(history, ptr)
	c.AnyOf = make([]*Shape, len(s.AnyOf))
	for i, item := range s.AnyOf {
		an := (*item).clone(history)
		c.AnyOf[i] = &an
	}
	return ptr
}

// Inherit merges the source shape into the target shape.
func (s *UnionShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*UnionShape)
	if !ok {
		return nil, NewError("cannot inherit from different type", s.Location, WithPosition(&s.Position), WithInfo("source", source.Base().Type), WithInfo("target", s.Base().Type))
	}
	// TODO: Facets need merging
	// TODO: This can be optimized
	var sourceUnionTypes map[string]struct{} = make(map[string]struct{})
	var filtered []*Shape
	for _, sourceMember := range ss.AnyOf {
		sourceUnionTypes[(*sourceMember).Base().Type] = struct{}{}
		for _, targetMember := range s.AnyOf {
			if (*sourceMember).Base().Type == (*targetMember).Base().Type {
				// Clone is required to avoid modifying the original target member shape.
				cs := (*targetMember).Clone()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.Base().Id = generateShapeId()
				ms, err := cs.Inherit(*sourceMember)
				if err != nil {
					return nil, NewWrappedError("merge union member", err, s.Location)
				}
				filtered = append(filtered, &ms)
			}
		}
	}
	for _, targetMember := range s.AnyOf {
		if _, ok := sourceUnionTypes[(*targetMember).Base().Type]; !ok {
			return nil, NewError("target union includes an incompatible type", s.Location, WithPosition(&s.Position), WithInfo("target_type", (*targetMember).Base().Type), WithInfo("source_types", sourceUnionTypes))
		}
	}
	s.AnyOf = filtered
	return s, nil
}

func (s *UnionShape) ToJSONSchema() *JSONSchema {
	schema := &JSONSchema{
		ID:     s.Id,
		AnyOf:  make([]*JSONSchema, len(s.AnyOf)),
		Extras: make(map[string]interface{}),
	}
	for i, item := range s.AnyOf {
		schema.AnyOf[i] = (*item).ToJSONSchema()
	}
	schema.WithRamlData(s.Base())
	return schema
}

func (s *UnionShape) Check() error {
	return nil
}

type JSONShape struct {
	BaseShape
}

func (s *JSONShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *JSONShape) Clone() Shape {
	c := *s
	return &c
}

func (s *JSONShape) clone(history []Shape) Shape {
	return s.Clone()
}

func (s *JSONShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	return nil
}

func (s *JSONShape) Inherit(source Shape) (Shape, error) {
	_, ok := source.(*JSONShape)
	if !ok {
		return nil, NewError("cannot inherit from different type", s.Location, WithPosition(&s.Position), WithInfo("source", source.Base().Type), WithInfo("target", s.Base().Type))
	}
	return s, nil
}

func (s *JSONShape) ToJSONSchema() *JSONSchema {
	return nil
}

func (s *JSONShape) Check() error {
	return nil
}

type UnknownShape struct {
	BaseShape

	facets []*yaml.Node
}

func (s *UnknownShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *UnknownShape) Clone() Shape {
	c := *s
	return &c
}

func (s *UnknownShape) clone(history []Shape) Shape {
	return s.Clone()
}

func (s *UnknownShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	s.facets = v
	return nil
}

func (s *UnknownShape) Inherit(source Shape) (Shape, error) {
	return nil, NewError("cannot inherit from unknown shape", s.Location, WithPosition(&s.Position))
}

func (s *UnknownShape) ToJSONSchema() *JSONSchema {
	return nil
}

func (s *UnknownShape) Check() error {
	return nil
}

type RecursiveShape struct {
	BaseShape

	Head *Shape
}

func (s *RecursiveShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	return nil
}

func (s *RecursiveShape) Base() *BaseShape {
	return &s.BaseShape
}

func (s *RecursiveShape) Clone() Shape {
	// TODO: How to clone recursive shape?
	c := *s
	return &c
}

func (s *RecursiveShape) clone(history []Shape) Shape {
	return s.Clone()
}

func (s *RecursiveShape) Inherit(source Shape) (Shape, error) {
	return nil, NewError("cannot inherit from recursive shape", s.Location, WithPosition(&s.Position))
}

func (s *RecursiveShape) ToJSONSchema() *JSONSchema {
	// TODO: Conversion of recursive schema doesn't make much sense as atomic operation.
	return &JSONSchema{
		Ref: (*s.Head).Base().Id,
	}
}

func (s *RecursiveShape) Check() error {
	return nil
}
