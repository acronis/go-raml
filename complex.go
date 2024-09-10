package raml

import (
	"fmt"
	"regexp"
	"strconv"
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

func (s *ArrayShape) Validate(v interface{}, ctxPath string) error {
	i, ok := v.([]interface{})
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected []interface{}", v)
	}

	arrayLen := uint64(len(i))
	if s.MinItems != nil && arrayLen < *s.MinItems {
		return fmt.Errorf("array must have at least %d items", *s.MinItems)
	}
	if s.MaxItems != nil && arrayLen > *s.MaxItems {
		return fmt.Errorf("array must have not more than %d items", *s.MaxItems)
	}
	validateUniqueItems := s.UniqueItems != nil && *s.UniqueItems
	uniqueItems := make(map[interface{}]struct{})
	for ii, item := range i {
		ctxPath := ctxPath + "[" + strconv.Itoa(ii) + "]"
		if s.Items != nil {
			if err := (*s.Items).Validate(item, ctxPath); err != nil {
				return fmt.Errorf("validate array item %s: %w", ctxPath, err)
			}
		}
		if validateUniqueItems {
			uniqueItems[item] = struct{}{}
		}
	}
	if validateUniqueItems && len(uniqueItems) != len(i) {
		return fmt.Errorf("array contains duplicate items")
	}

	return nil
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

func (s *ArrayShape) Check() error {
	if s.MinItems != nil && s.MaxItems != nil && *s.MinItems > *s.MaxItems {
		return NewError("minItems must be less than or equal to maxItems", s.Location, WithPosition(&s.Position))
	}
	if s.Items != nil {
		if err := (*s.Items).Check(); err != nil {
			return NewWrappedError("check items", err, s.Location, WithPosition(&(*s.Items).Base().Position))
		}
	}
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
	PatternProperties    map[string]PatternProperty
	MinProperties        *uint64
	MaxProperties        *uint64
}

// ObjectShape represents an object shape.
type ObjectShape struct {
	BaseShape

	ObjectFacets
}

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
			for j := 0; j != len(valueNode.Content); j += 2 {
				name := valueNode.Content[j].Value
				data := valueNode.Content[j+1]
				if name != "" && name[0] == '/' && name[len(name)-1] == '/' {
					if s.PatternProperties == nil {
						s.PatternProperties = make(map[string]PatternProperty)
					}
					property, err := s.raml.makePatternProperty(name, data, s.Location)
					if err != nil {
						return NewWrappedError("make pattern property", err, s.Location, WithNodePosition(data))
					}
					s.PatternProperties[name] = property
					// s.raml.PutTypeIntoFragment(s.Name+"#"+property.Name, s.Location, property.Shape)
					s.raml.PutShapePtr(property.Shape)
				} else {
					if s.Properties == nil {
						s.Properties = make(map[string]Property)
					}
					property, err := s.raml.makeProperty(name, data, s.Location)
					if err != nil {
						return NewWrappedError("make property", err, s.Location, WithNodePosition(data))
					}
					s.Properties[property.Name] = property
					// s.raml.PutTypeIntoFragment(s.Name+"#"+property.Name, s.Location, property.Shape)
					s.raml.PutShapePtr(property.Shape)
				}
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
	if c.PatternProperties != nil {
		c.PatternProperties = make(map[string]PatternProperty, len(s.PatternProperties))
		for k, v := range s.PatternProperties {
			p := (*v.Shape).clone(history)
			v.Shape = &p
			c.PatternProperties[k] = v
		}
	}
	return ptr
}

func (s *ObjectShape) Validate(v interface{}, ctxPath string) error {
	i, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected map[string]interface{}", v)
	}

	mapLen := uint64(len(i))
	if s.MinProperties != nil && mapLen < *s.MinProperties {
		return fmt.Errorf("object must have at least %d properties", *s.MinProperties)
	}
	if s.MaxProperties != nil && mapLen > *s.MaxProperties {
		return fmt.Errorf("object must have not more than %d properties", *s.MaxProperties)
	}
	restrictedAdditionalProperties := s.AdditionalProperties != nil && !*s.AdditionalProperties
	for k, item := range i {
		// Explicitly defined properties have priority over pattern properties.
		if s.Properties != nil {
			ctxPath := ctxPath + "." + k
			if p, ok := s.Properties[k]; ok {
				ps := *p.Shape
				if err := ps.Validate(item, ctxPath); err != nil {
					return fmt.Errorf("validate property %s: %w", ctxPath, err)
				}
			} else if restrictedAdditionalProperties {
				// Will never happen if PatternProperties are present because additional properties false is not allowed together with pattern properties.
				return fmt.Errorf("unexpected additional property \"%s\"", k)
			}
		} else if s.PatternProperties != nil {
			ctxPath := ctxPath + "." + k
			validated := false
			// TODO: Collect errors
			for _, pp := range s.PatternProperties {
				if pp.Pattern.MatchString(k) {
					ps := *pp.Shape
					// TODO: The first pattern to validate prevails. However, since pattern properties is a map, the validation can be random.
					if err := ps.Validate(item, ctxPath); err == nil {
						validated = true
						break
					}
				}
			}
			if !validated {
				return fmt.Errorf("property \"%s\" failed to match or validate against any pattern property", k)
			}
		} else if restrictedAdditionalProperties {
			// Special case when an object doesn't define any properties but specify additional properties false
			return fmt.Errorf("unexpected additional property \"%s\"", k)
		}
	}

	return nil
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
					return nil, NewWrappedError("inherit property", err, s.Base().Location, WithPosition(&(*targetProp.Shape).Base().Position), WithInfo("property", k))
				}
			} else {
				s.Properties[k] = sourceProp
			}
		}
	}
	if s.PatternProperties == nil {
		s.PatternProperties = ss.PatternProperties
	} else if ss.PatternProperties != nil {
		for k, sourceProp := range ss.PatternProperties {
			if targetProp, ok := s.PatternProperties[k]; ok {
				_, err := s.raml.Inherit(*sourceProp.Shape, *s.PatternProperties[k].Shape)
				if err != nil {
					return nil, NewWrappedError("inherit pattern property", err, s.Base().Location, WithPosition(&(*targetProp.Shape).Base().Position), WithInfo("property", k))
				}
			} else {
				s.PatternProperties[k] = sourceProp
			}
		}
	}
	return s, nil
}

func (s *ObjectShape) Check() error {
	if s.MinProperties != nil && s.MaxProperties != nil && *s.MinProperties > *s.MaxProperties {
		return NewError("minProperties must be less than or equal to maxProperties", s.Location, WithPosition(&s.Position))
	}
	if s.PatternProperties != nil {
		if s.AdditionalProperties != nil && !*s.AdditionalProperties {
			return NewError("pattern properties are not allowed with additionalProperties false", s.Location, WithPosition(&s.Position))
		}
		for _, prop := range s.PatternProperties {
			if err := (*prop.Shape).Check(); err != nil {
				return NewWrappedError("check pattern property", err, s.Location, WithPosition(&(*prop.Shape).Base().Position), WithInfo("property", prop.Pattern.String()))
			}
		}
	}
	if s.Properties != nil {
		for _, prop := range s.Properties {
			if err := (*prop.Shape).Check(); err != nil {
				return NewWrappedError("check property", err, s.Location, WithPosition(&(*prop.Shape).Base().Position), WithInfo("property", prop.Name))
			}
		}
		// FIXME: Need to validate on which level the discriminator is applied to avoid potential false positives.
		// Inline definitions with discriminator are not allowed.
		// TODO: Setting discriminator should be allowed only on scalar shapes.
		if s.Discriminator != nil {
			prop, ok := s.Properties[*s.Discriminator]
			if !ok {
				return NewError("discriminator property not found", s.Location, WithPosition(&s.Position), WithInfo("discriminator", *s.Discriminator))
			}
			discriminatorValue := s.DiscriminatorValue
			if discriminatorValue == nil {
				discriminatorValue = s.Base().Name
			}
			ps := *prop.Shape
			if err := ps.Validate(discriminatorValue, "$"); err != nil {
				return NewWrappedError("validate discriminator value", err, s.Location, WithPosition(&s.Base().Position), WithInfo("discriminator", *s.Discriminator))
			}
		}
	} else if s.Discriminator != nil {
		return NewError("discriminator without properties", s.Location, WithPosition(&s.Position))
	}
	return nil
}

// makeProperty creates a pattern property from a YAML node.
func (r *RAML) makePatternProperty(name string, v *yaml.Node, location string) (PatternProperty, error) {
	shape, err := r.makeShape(v, name, location)
	if err != nil {
		return PatternProperty{}, NewWrappedError("make shape", err, location, WithNodePosition(v))
	}
	re, err := regexp.Compile(name[1 : len(name)-1])
	if err != nil {
		return PatternProperty{}, NewWrappedError("compile pattern", err, location, WithNodePosition(v))
	}
	return PatternProperty{
		Pattern: re,
		Shape:   shape,
		raml:    r,
	}, nil
}

// makeProperty creates a property from a YAML node.
func (r *RAML) makeProperty(name string, v *yaml.Node, location string) (Property, error) {
	shape, err := r.makeShape(v, name, location)
	if err != nil {
		return Property{}, NewWrappedError("make shape", err, location, WithNodePosition(v))
	}
	propertyName := name
	var required bool
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

// Property represents a pattern property of an object shape.
type PatternProperty struct {
	Pattern *regexp.Regexp
	Shape   *Shape
	// Pattern properties are always optional.
	raml *RAML
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

func (s *UnionShape) Validate(v interface{}, ctxPath string) error {
	// TODO: Collect errors
	for _, item := range s.AnyOf {
		if err := (*item).Validate(v, ctxPath); err == nil {
			return nil
		}
	}
	return NewError("value does not match any type", s.Location, WithPosition(&s.Position))
}

// Inherit merges the source shape into the target shape.
func (s *UnionShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*UnionShape)
	if !ok {
		return nil, NewError("cannot inherit from different type", s.Location, WithPosition(&s.Position), WithInfo("source", source.Base().Type), WithInfo("target", s.Base().Type))
	}
	if len(s.AnyOf) == 0 {
		s.AnyOf = ss.AnyOf
		return s, nil
	}
	// TODO: Facets need merging
	// TODO: This can be optimized
	sourceUnionTypes := make(map[string]struct{})
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

func (s *UnionShape) Check() error {
	for _, item := range s.AnyOf {
		if err := (*item).Check(); err != nil {
			return NewWrappedError("check union member", err, s.Location, WithPosition(&(*item).Base().Position))
		}
	}
	// TODO: Unions may have enum facets
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

func (s *JSONShape) Validate(v interface{}, ctxPath string) error {
	// TODO: Implement validation with JSON Schema
	return nil
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

func (s *JSONShape) Check() error {
	// TODO: JSON Schema check
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

func (s *UnknownShape) Validate(v interface{}, ctxPath string) error {
	return fmt.Errorf("cannot validate against unknown shape")
}

func (s *UnknownShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	s.facets = v
	return nil
}

func (s *UnknownShape) Inherit(source Shape) (Shape, error) {
	return nil, NewError("cannot inherit from unknown shape", s.Location, WithPosition(&s.Position))
}

func (s *UnknownShape) Check() error {
	return NewError("cannot check unknown shape", s.Location, WithPosition(&s.Position))
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
	// TODO: Should it also copy ref?
	c := *s
	return &c
}

func (s *RecursiveShape) clone(history []Shape) Shape {
	return s.Clone()
}

func (s *RecursiveShape) Validate(v interface{}, ctxPath string) error {
	if err := (*s.Head).Validate(v, ctxPath); err != nil {
		return fmt.Errorf("validate recursive shape: %w", err)
	}
	return nil
}

func (s *RecursiveShape) Inherit(source Shape) (Shape, error) {
	return nil, NewError("cannot inherit from recursive shape", s.Location, WithPosition(&s.Position))
}

func (s *RecursiveShape) Check() error {
	return nil
}
