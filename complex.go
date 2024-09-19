package raml

import (
	"fmt"
	"regexp"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-raml/stacktrace"
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
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Items == nil {
		s.Items = ss.Items
	} else if ss.Items != nil {
		_, err := s.raml.Inherit(*s.Items, *ss.Items)
		if err != nil {
			return nil, stacktrace.NewWrapped("merge array items", err, s.Location,
				stacktrace.WithPosition(&(*s.Items).Base().Position))
		}
	}
	if s.MinItems == nil {
		s.MinItems = ss.MinItems
	} else if ss.MinItems != nil && *s.MinItems > *ss.MinItems {
		return nil, stacktrace.New("minItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MinItems),
			stacktrace.WithInfo("target", *s.MinItems))
	}
	if s.MaxItems == nil {
		s.MaxItems = ss.MaxItems
	} else if ss.MaxItems != nil && *s.MaxItems < *ss.MaxItems {
		return nil, stacktrace.New("maxItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MaxItems),
			stacktrace.WithInfo("target", *s.MaxItems))
	}
	if s.UniqueItems == nil {
		s.UniqueItems = ss.UniqueItems
	} else if ss.UniqueItems != nil && *ss.UniqueItems && !*s.UniqueItems {
		return nil, stacktrace.New("uniqueItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.UniqueItems),
			stacktrace.WithInfo("target", *s.UniqueItems))
	}
	return s, nil
}

func (s *ArrayShape) Check() error {
	if s.MinItems != nil && s.MaxItems != nil && *s.MinItems > *s.MaxItems {
		return stacktrace.New("minItems must be less than or equal to maxItems", s.Location, stacktrace.WithPosition(&s.Position))
	}
	if s.Items != nil {
		if err := (*s.Items).Check(); err != nil {
			return stacktrace.NewWrapped("check items", err, s.Location, stacktrace.WithPosition(&(*s.Items).Base().Position))
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
				return stacktrace.NewWrapped("decode minItems", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "maxItems" {
			if err := valueNode.Decode(&s.MaxItems); err != nil {
				return stacktrace.NewWrapped("decode maxItems: %w", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "items" {
			name := "items"
			shape, err := s.raml.makeShape(valueNode, name, s.Location)
			if err != nil {
				return stacktrace.NewWrapped("make shape", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
			s.Items = shape
			// s.raml.PutTypeIntoFragment(s.Name+"#items", s.Location, s.Items)
			s.raml.PutShapePtr(s.Items)
		} else if node.Value == "uniqueItems" {
			if err := valueNode.Decode(&s.UniqueItems); err != nil {
				return stacktrace.NewWrapped("decode uniqueItems", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else {
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return stacktrace.NewWrapped("make node", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
		}
	}
	return nil
}

// ObjectFacets contains constraints for object shapes.
type ObjectFacets struct {
	Discriminator        *string
	DiscriminatorValue   any
	AdditionalProperties *bool
	Properties           *orderedmap.OrderedMap[string, Property]
	PatternProperties    *orderedmap.OrderedMap[string, PatternProperty]
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
				return stacktrace.NewWrapped("decode additionalProperties", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "discriminator" {
			if err := valueNode.Decode(&s.Discriminator); err != nil {
				return stacktrace.NewWrapped("decode discriminator", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "discriminatorValue" {
			if err := valueNode.Decode(&s.DiscriminatorValue); err != nil {
				return stacktrace.NewWrapped("decode discriminatorValue", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "minProperties" {
			if err := valueNode.Decode(&s.MinProperties); err != nil {
				return stacktrace.NewWrapped("decode minProperties", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "maxProperties" {
			if err := valueNode.Decode(&s.MaxProperties); err != nil {
				return stacktrace.NewWrapped("decode maxProperties", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
		} else if node.Value == "properties" {
			for j := 0; j != len(valueNode.Content); j += 2 {
				nodeName := valueNode.Content[j].Value
				data := valueNode.Content[j+1]

				propertyName, hasImplicitOptional := s.raml.chompImplicitOptional(nodeName)
				if len(propertyName) > 1 && propertyName[0] == '/' && propertyName[len(propertyName)-1] == '/' {
					if s.PatternProperties == nil {
						s.PatternProperties = orderedmap.New[string, PatternProperty]()
					}
					property, err := s.raml.makePatternProperty(nodeName, propertyName, data, s.Location, hasImplicitOptional)
					if err != nil {
						return stacktrace.NewWrapped("make pattern property", err, s.Location, stacktrace.WithNodePosition(data))
					}
					s.PatternProperties.Set(propertyName, property)
					// s.raml.PutTypeIntoFragment(s.Name+"#"+property.Name, s.Location, property.Shape)
					s.raml.PutShapePtr(property.Shape)
				} else {
					if s.Properties == nil {
						s.Properties = orderedmap.New[string, Property]()
					}
					property, err := s.raml.makeProperty(nodeName, propertyName, data, s.Location, hasImplicitOptional)
					if err != nil {
						return stacktrace.NewWrapped("make property", err, s.Location, stacktrace.WithNodePosition(data))
					}
					s.Properties.Set(property.Name, property)
					// s.raml.PutTypeIntoFragment(s.Name+"#"+property.Name, s.Location, property.Shape)
					s.raml.PutShapePtr(property.Shape)
				}
			}
		} else {
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return stacktrace.NewWrapped("make node", err, s.Location, stacktrace.WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(node.Value, n)
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
		c.Properties = orderedmap.New[string, Property](s.Properties.Len())
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			k, prop := pair.Key, pair.Value
			ps := (*prop.Shape).clone(history)
			prop.Shape = &ps
			c.Properties.Set(k, prop)
		}
	}
	if c.PatternProperties != nil {
		c.PatternProperties = orderedmap.New[string, PatternProperty](s.PatternProperties.Len())
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, prop := pair.Key, pair.Value
			ps := (*prop.Shape).clone(history)
			prop.Shape = &ps
			c.PatternProperties.Set(k, prop)
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
		ctxPath := ctxPath + "." + k
		if s.Properties != nil {
			if p, ok := s.Properties.Get(k); ok {
				ps := *p.Shape
				if err := ps.Validate(item, ctxPath); err != nil {
					return fmt.Errorf("validate property %s: %w", ctxPath, err)
				}
				continue
			}
		}
		if s.PatternProperties != nil {
			found := false
			for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
				pp := pair.Value
				// NOTE: We validate only those keys that match the pattern.
				// The keys that do not match are considered as additional properties and are not validated.
				if pp.Pattern.MatchString(k) {
					ps := *pp.Shape
					// NOTE: The first defined pattern property to validate prevails.
					if err := ps.Validate(item, ctxPath); err == nil {
						found = true
						break
					}
				}
			}
			if found {
				continue
			}
		}
		// Will never happen if pattern properties are present.
		if restrictedAdditionalProperties {
			return fmt.Errorf("unexpected additional property \"%s\"", k)
		}
	}

	return nil
}

// Inherit merges the source shape into the target shape.
func (s *ObjectShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*ObjectShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
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
		return nil, stacktrace.New("minProperties constraint violation", s.Location, stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MinProperties), stacktrace.WithInfo("target", *s.MinProperties))
	}
	if s.MaxProperties == nil {
		s.MaxProperties = ss.MaxProperties
	} else if ss.MaxProperties != nil && *s.MaxProperties > *ss.MaxProperties {
		return nil, stacktrace.New("maxProperties constraint violation", s.Location, stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MaxProperties), stacktrace.WithInfo("target", *s.MaxProperties))
	}

	if s.Properties == nil {
		s.Properties = ss.Properties
	} else if ss.Properties != nil {
		for pair := ss.Properties.Oldest(); pair != nil; pair = pair.Next() {
			k, sourceProp := pair.Key, pair.Value
			if targetProp, ok := s.Properties.Get(k); ok {
				if sourceProp.Required && !targetProp.Required {
					return nil, stacktrace.New("cannot make required property optional", s.Location,
						stacktrace.WithPosition(&(*targetProp.Shape).Base().Position),
						stacktrace.WithInfo("property", k),
						stacktrace.WithInfo("source", sourceProp.Required),
						stacktrace.WithInfo("target", targetProp.Required),
						stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				_, err := s.raml.Inherit(*sourceProp.Shape, *targetProp.Shape)
				if err != nil {
					return nil, stacktrace.NewWrapped("inherit property", err, s.Base().Location,
						stacktrace.WithPosition(&(*targetProp.Shape).Base().Position),
						stacktrace.WithInfo("property", k),
						stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
			} else {
				s.Properties.Set(k, sourceProp)
			}
		}
	}
	if s.PatternProperties == nil {
		s.PatternProperties = ss.PatternProperties
	} else if ss.PatternProperties != nil {
		for pair := ss.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, sourceProp := pair.Key, pair.Value
			if targetProp, ok := s.PatternProperties.Get(k); ok {
				_, err := s.raml.Inherit(*sourceProp.Shape, *targetProp.Shape)
				if err != nil {
					return nil, stacktrace.NewWrapped("inherit pattern property", err, s.Base().Location,
						stacktrace.WithPosition(&(*targetProp.Shape).Base().Position),
						stacktrace.WithInfo("property", k),
						stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
			} else {
				s.PatternProperties.Set(k, sourceProp)
			}
		}
	}
	return s, nil
}

func (s *ObjectShape) Check() error {
	if s.MinProperties != nil && s.MaxProperties != nil && *s.MinProperties > *s.MaxProperties {
		return stacktrace.New("minProperties must be less than or equal to maxProperties", s.Location, stacktrace.WithPosition(&s.Position))
	}
	if s.PatternProperties != nil {
		if s.AdditionalProperties != nil && !*s.AdditionalProperties {
			// TODO: We actually can allow pattern properties with "additionalProperties: false" for stricter validation.
			// This will contradict RAML 1.0 spec, but JSON Schema allows that.
			// https://json-schema.org/understanding-json-schema/reference/object#additionalproperties
			return stacktrace.New("pattern properties are not allowed with \"additionalProperties: false\"", s.Location, stacktrace.WithPosition(&s.Position))
		}
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			if err := (*prop.Shape).Check(); err != nil {
				return stacktrace.NewWrapped("check pattern property", err, s.Location, stacktrace.WithPosition(&(*prop.Shape).Base().Position), stacktrace.WithInfo("property", prop.Pattern.String()))
			}
		}
	}
	if s.Properties != nil {
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			if err := (*prop.Shape).Check(); err != nil {
				return stacktrace.NewWrapped("check property", err, s.Location, stacktrace.WithPosition(&(*prop.Shape).Base().Position), stacktrace.WithInfo("property", prop.Name))
			}
		}
		// FIXME: Need to validate on which level the discriminator is applied to avoid potential false positives.
		// Inline definitions with discriminator are not allowed.
		// TODO: Setting discriminator should be allowed only on scalar shapes.
		if s.Discriminator != nil {
			prop, ok := s.Properties.Get(*s.Discriminator)
			if !ok {
				return stacktrace.New("discriminator property not found", s.Location, stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("discriminator", *s.Discriminator))
			}
			discriminatorValue := s.DiscriminatorValue
			if discriminatorValue == nil {
				discriminatorValue = s.Base().Name
			}
			ps := *prop.Shape
			if err := ps.Validate(discriminatorValue, "$"); err != nil {
				return stacktrace.NewWrapped("validate discriminator value", err, s.Location, stacktrace.WithPosition(&s.Base().Position), stacktrace.WithInfo("discriminator", *s.Discriminator))
			}
		}
	} else if s.Discriminator != nil {
		return stacktrace.New("discriminator without properties", s.Location, stacktrace.WithPosition(&s.Position))
	}
	return nil
}

// makeProperty creates a pattern property from a YAML node.
func (r *RAML) makePatternProperty(nodeName string, propertyName string, v *yaml.Node, location string, hasImplicitOptional bool) (PatternProperty, error) {
	shape, err := r.makeShape(v, nodeName, location)
	if err != nil {
		return PatternProperty{}, stacktrace.NewWrapped("make shape", err, location, stacktrace.WithNodePosition(v))
	}
	// Pattern properties cannot be required
	if (*shape).Base().Required != nil || hasImplicitOptional {
		return PatternProperty{}, stacktrace.New("'required' facet is not supported on pattern property", location, stacktrace.WithNodePosition(v))
	}
	re, err := regexp.Compile(propertyName[1 : len(propertyName)-1])
	if err != nil {
		return PatternProperty{}, stacktrace.NewWrapped("compile pattern", err, location, stacktrace.WithNodePosition(v))
	}
	return PatternProperty{
		Pattern: re,
		Shape:   shape,
		raml:    r,
	}, nil
}

func (r *RAML) chompImplicitOptional(nodeName string) (string, bool) {
	nameLen := len(nodeName)
	if nodeName != "" && nodeName[nameLen-1] == '?' {
		return nodeName[:nameLen-1], true
	}
	return nodeName, false
}

// makeProperty creates a property from a YAML node.
func (r *RAML) makeProperty(nodeName string, propertyName string, v *yaml.Node, location string, hasImplicitOptional bool) (Property, error) {
	shape, err := r.makeShape(v, nodeName, location)
	if err != nil {
		return Property{}, stacktrace.NewWrapped("make shape", err, location, stacktrace.WithNodePosition(v))
	}
	finalName := propertyName
	var required bool
	shapeRequired := (*shape).Base().Required
	if shapeRequired == nil {
		// If shape has no "required" facet, requirement depends only on whether "?"" was used in node name.
		required = !hasImplicitOptional
	} else {
		// If shape explicitly defines "required" facet combined with "?" in node name - explicit definition prevails and property name keeps the node name.
		// Otherwise, keep propertyName that has the last "?" chomped.
		if hasImplicitOptional {
			finalName = nodeName
		}
		required = *shapeRequired
	}
	return Property{
		Name:     finalName,
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
	return stacktrace.New("value does not match any type", s.Location, stacktrace.WithPosition(&s.Position))
}

// Inherit merges the source shape into the target shape.
func (s *UnionShape) Inherit(source Shape) (Shape, error) {
	ss, ok := source.(*UnionShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location, stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type), stacktrace.WithInfo("target", s.Base().Type))
	}
	if len(s.AnyOf) == 0 {
		s.AnyOf = ss.AnyOf
		return s, nil
	}
	// TODO: Implement enum facets inheritance
	var finalFiltered []*Shape
	for _, sourceMember := range ss.AnyOf {
		var filtered []*Shape
		for _, targetMember := range s.AnyOf {
			if (*sourceMember).Base().Type == (*targetMember).Base().Type {
				// Clone is required to avoid modifying the original target member shape.
				cs := (*targetMember).Clone()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.Base().Id = generateShapeId()
				ms, err := cs.Inherit(*sourceMember)
				if err != nil {
					// TODO: Collect errors
					// stacktrace.NewWrapped("merge union member", err, s.Location)
					continue
				}
				filtered = append(filtered, &ms)
			}
		}
		if len(filtered) == 0 {
			return nil, stacktrace.New("failed to find compatible union member", s.Location, stacktrace.WithPosition(&s.Position))
		}
		finalFiltered = append(finalFiltered, filtered...)
	}
	s.AnyOf = finalFiltered
	return s, nil
}

func (s *UnionShape) Check() error {
	for _, item := range s.AnyOf {
		if err := (*item).Check(); err != nil {
			return stacktrace.NewWrapped("check union member", err, s.Location, stacktrace.WithPosition(&(*item).Base().Position))
		}
	}
	// TODO: Unions may have enum facets
	return nil
}

type JSONShape struct {
	BaseShape

	Schema *JSONSchema
	Raw    string
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
	ss, ok := source.(*JSONShape)
	if !ok {
		return nil, stacktrace.New("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if s.Raw != "" && ss.Raw != "" && s.Raw != ss.Raw {
		return nil, stacktrace.New("cannot inherit from different JSON schema", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	s.Schema = ss.Schema
	s.Raw = ss.Raw
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
	return nil, stacktrace.New("cannot inherit from unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *UnknownShape) Check() error {
	return stacktrace.New("cannot check unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
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
	return nil, stacktrace.New("cannot inherit from recursive shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *RecursiveShape) Check() error {
	return nil
}
