package raml

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/xeipuuv/gojsonschema"
	"github.com/zeebo/xxh3"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

type noScalarShape struct{}

func (noScalarShape) IsScalar() bool {
	return false
}

// ArrayFacets contains constraints for array shapes.
type ArrayFacets struct {
	Items       *BaseShape
	MinItems    *uint64
	MaxItems    *uint64
	UniqueItems *bool
}

// ArrayShape represents an array shape.
type ArrayShape struct {
	noScalarShape
	*BaseShape

	ArrayFacets
}

// Base returns the base shape.
func (s *ArrayShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *ArrayShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *ArrayShape) clone(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	if c.Items != nil {
		c.Items = c.Items.clone(clonedMap)
	}
	return &c
}

func (s *ArrayShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*ArrayShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Items = ss.Items
	s.MinItems = ss.MinItems
	s.MaxItems = ss.MaxItems
	s.UniqueItems = ss.UniqueItems
	return s, nil
}

func hashInterfaceFast(v interface{}) (uint64, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return 0, fmt.Errorf("marshal: %w", err)
	}

	// Use xxhash for fast hashing.
	return xxh3.Hash(data), nil
}

func (s *ArrayShape) validate(v interface{}, ctxPath string) error {
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
	uniqueItems := make(map[uint64]struct{})
	for ii, item := range i {
		ctxPathA := ctxPath + "[" + strconv.Itoa(ii) + "]"
		if s.Items != nil {
			if err := s.Items.Shape.validate(item, ctxPathA); err != nil {
				return fmt.Errorf("validate array item %s: %w", ctxPathA, err)
			}
		}
		if validateUniqueItems {
			itemHash, err := hashInterfaceFast(item)
			if err != nil {
				return fmt.Errorf("hash array item %s: %w", ctxPathA, err)
			}

			uniqueItems[itemHash] = struct{}{}
		}
	}
	if validateUniqueItems && len(uniqueItems) != len(i) {
		return fmt.Errorf("array contains duplicate items")
	}

	return nil
}

// Inherit merges the source shape into the target shape.
func (s *ArrayShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*ArrayShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Type))
	}
	if s.Items == nil {
		s.Items = ss.Items
	} else if ss.Items != nil {
		_, err := s.Items.Inherit(ss.Items)
		if err != nil {
			return nil, StacktraceNewWrapped("merge array items", err, s.Location,
				stacktrace.WithPosition(&s.Items.Position))
		}
	}
	if s.MinItems == nil {
		s.MinItems = ss.MinItems
	} else if ss.MinItems != nil && *s.MinItems < *ss.MinItems {
		return nil, StacktraceNew("minItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MinItems),
			stacktrace.WithInfo("target", *s.MinItems))
	}
	if s.MaxItems == nil {
		s.MaxItems = ss.MaxItems
	} else if ss.MaxItems != nil && *s.MaxItems > *ss.MaxItems {
		return nil, StacktraceNew("maxItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.MaxItems),
			stacktrace.WithInfo("target", *s.MaxItems))
	}
	if s.UniqueItems == nil {
		s.UniqueItems = ss.UniqueItems
	} else if ss.UniqueItems != nil && *ss.UniqueItems && !*s.UniqueItems {
		return nil, StacktraceNew("uniqueItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", *ss.UniqueItems),
			stacktrace.WithInfo("target", *s.UniqueItems))
	}
	return s, nil
}

func (s *ArrayShape) check() error {
	if s.MinItems != nil && s.MaxItems != nil && *s.MinItems > *s.MaxItems {
		return StacktraceNew("minItems must be less than or equal to maxItems", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	if s.Items != nil {
		if err := s.Items.Check(); err != nil {
			return StacktraceNewWrapped("check items", err, s.Location,
				stacktrace.WithPosition(&s.Items.Position))
		}
	}
	return nil
}

// UnmarshalYAMLNodes unmarshals the array shape from YAML nodes.
func (s *ArrayShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location, stacktrace.WithPosition(&s.Position))
	}
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]

		switch node.Value {
		case FacetMinItems:
			if err := valueNode.Decode(&s.MinItems); err != nil {
				return StacktraceNewWrapped("decode", err, s.Location,
					WithNodePosition(valueNode),
					stacktrace.WithInfo("facet", FacetMinItems))
			}
		case FacetMaxItems:
			if err := valueNode.Decode(&s.MaxItems); err != nil {
				return StacktraceNewWrapped("decode", err, s.Location,
					WithNodePosition(valueNode),
					stacktrace.WithInfo("facet", FacetMaxItems))
			}
		case FacetItems:
			shape, err := s.raml.makeNewShapeYAML(valueNode, FacetItems, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make shape", err, s.Location,
					WithNodePosition(valueNode),
					stacktrace.WithInfo("facet", FacetItems))
			}
			s.Items = shape
		case FacetUniqueItems:
			if err := valueNode.Decode(&s.UniqueItems); err != nil {
				return StacktraceNewWrapped("decode", err, s.Location,
					WithNodePosition(valueNode),
					stacktrace.WithInfo("facet", FacetUniqueItems))
			}
		default:
			n, err := s.raml.makeRootNode(valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
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
	noScalarShape
	*BaseShape

	ObjectFacets
}

func (s *ObjectShape) unmarshalPatternProperties(
	nodeName, propertyName string, data *yaml.Node, hasImplicitOptional bool) error {
	if s.PatternProperties == nil {
		s.PatternProperties = orderedmap.New[string, PatternProperty]()
	}
	property, err := s.raml.makePatternProperty(nodeName, propertyName, data, s.Location,
		hasImplicitOptional)
	if err != nil {
		return StacktraceNewWrapped("make pattern property", err, s.Location,
			WithNodePosition(data))
	}
	s.PatternProperties.Set(propertyName, property)
	return nil
}

func (s *ObjectShape) unmarshalProperty(nodeName string, data *yaml.Node) error {
	propertyName, hasImplicitOptional := s.raml.chompImplicitOptional(nodeName)
	if len(propertyName) > 1 && propertyName[0] == '/' && propertyName[len(propertyName)-1] == '/' {
		return s.unmarshalPatternProperties(nodeName, propertyName, data, hasImplicitOptional)
	}

	if s.Properties == nil {
		s.Properties = orderedmap.New[string, Property]()
	}
	property, err := s.raml.makeProperty(nodeName, propertyName, data, s.Location, hasImplicitOptional)
	if err != nil {
		return StacktraceNewWrapped("make property", err, s.Location, WithNodePosition(data))
	}
	s.Properties.Set(property.Name, property)
	return nil
}

func (s *ObjectShape) unmarshalYAMLNode(node, valueNode *yaml.Node) error {
	switch node.Value {
	case FacetAdditionalProperties:
		if err := valueNode.Decode(&s.AdditionalProperties); err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetAdditionalProperties))
		}
	case FacetDiscriminator:
		if err := valueNode.Decode(&s.Discriminator); err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetDiscriminator))
		}
	case FacetDiscriminatorValue:
		if err := valueNode.Decode(&s.DiscriminatorValue); err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetDiscriminatorValue))
		}
	case FacetMinProperties:
		if err := valueNode.Decode(&s.MinProperties); err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetMinProperties))
		}
	case FacetMaxProperties:
		if err := valueNode.Decode(&s.MaxProperties); err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetMaxProperties))
		}
	case FacetProperties:
		if len(valueNode.Content)%2 != 0 {
			return StacktraceNew("odd number of nodes", s.Location, stacktrace.WithPosition(&s.Position))
		}
		for j := 0; j != len(valueNode.Content); j += 2 {
			nodeName := valueNode.Content[j].Value
			data := valueNode.Content[j+1]

			if err := s.unmarshalProperty(nodeName, data); err != nil {
				return fmt.Errorf("unmarshal property: %w", err)
			}
		}
	default:
		n, err := s.raml.makeRootNode(valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(node.Value, n)
	}
	return nil
}

// UnmarshalYAMLNodes unmarshals the object shape from YAML nodes.
func (s *ObjectShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v)%2 != 0 {
		return StacktraceNew("odd number of nodes", s.Location, stacktrace.WithPosition(&s.Position))
	}
	for i := 0; i != len(v); i += 2 {
		node := v[i]
		valueNode := v[i+1]
		if err := s.unmarshalYAMLNode(node, valueNode); err != nil {
			return fmt.Errorf("unmarshal object facet: %w", err)
		}
	}
	return nil
}

// Base returns the base shape.
func (s *ObjectShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *ObjectShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *ObjectShape) clone(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	if c.Properties != nil {
		c.Properties = orderedmap.New[string, Property](s.Properties.Len())
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			k, prop := pair.Key, pair.Value
			prop.Base = prop.Base.clone(clonedMap)
			c.Properties.Set(k, prop)
		}
	}
	if c.PatternProperties != nil {
		c.PatternProperties = orderedmap.New[string, PatternProperty](s.PatternProperties.Len())
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, prop := pair.Key, pair.Value
			prop.Base = prop.Base.clone(clonedMap)
			c.PatternProperties.Set(k, prop)
		}
	}
	return &c
}

func (s *ObjectShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*ObjectShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Properties = ss.Properties
	s.PatternProperties = ss.PatternProperties
	s.MinProperties = ss.MinProperties
	s.MaxProperties = ss.MaxProperties
	s.AdditionalProperties = ss.AdditionalProperties
	s.Discriminator = ss.Discriminator
	s.DiscriminatorValue = ss.DiscriminatorValue
	return s, nil
}

func (s *ObjectShape) validatePatternProperty(
	k string,
	item interface{},
	ctxPathK string,
) (bool, error) {
	if s.PatternProperties == nil {
		return false, nil
	}
	var st *stacktrace.StackTrace
	for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		pp := pair.Value
		if !pp.Pattern.MatchString(k) {
			continue
		}
		// NOTE: The first defined pattern property to validate prevails.
		err := pp.Base.Shape.validate(item, ctxPathK)
		if err == nil {
			return true, nil
		}
		se := StacktraceNewWrapped("validate pattern property", err, s.Location,
			stacktrace.WithPosition(&pp.Base.Position),
			stacktrace.WithInfo("property", pp.Pattern.String()))
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
	}
	if st != nil {
		return true, st
	}
	return false, nil
}

func (s *ObjectShape) validateProperty(
	k string,
	item interface{},
	ctxPathK string,
) (bool, error) {
	if s.Properties == nil {
		return false, nil
	}
	p, present := s.Properties.Get(k)
	if !present {
		return false, nil
	}
	if err := p.Base.Shape.validate(item, ctxPathK); err != nil {
		return true, fmt.Errorf("validate property %s: %w", ctxPathK, err)
	}
	return true, nil
}

func (s *ObjectShape) validateProperties(ctxPath string, props map[string]interface{}) error {
	if missing := s.findMissingRequired(props); len(missing) > 0 {
		return fmt.Errorf("missing required properties: %s", strings.Join(missing, ", "))
	}

	restrictedAdditionalProperties := s.AdditionalProperties != nil && !*s.AdditionalProperties
	for k, item := range props {
		// Explicitly defined properties have priority over pattern properties.
		ctxPathK := ctxPath + "." + k

		found, err := s.validateProperty(k, item, ctxPathK)
		if err != nil {
			return fmt.Errorf("validate property %s: %w", ctxPathK, err)
		}
		if found {
			continue
		}
		if restrictedAdditionalProperties {
			return fmt.Errorf("unexpected additional property \"%s\"", k)
		}

		_, err = s.validatePatternProperty(k, item, ctxPathK)
		if err != nil {
			return fmt.Errorf("validate pattern property %s: %w", ctxPathK, err)
		}
	}
	return nil
}

func (s *ObjectShape) findMissingRequired(props map[string]interface{}) []string {
	var missing []string
	for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if !pair.Value.Required {
			continue
		}
		if _, ok := props[pair.Key]; !ok {
			missing = append(missing, pair.Key)
		}
	}
	return missing
}

func (s *ObjectShape) validate(v interface{}, ctxPath string) error {
	props, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected map[string]interface{}", v)
	}

	mapLen := uint64(len(props))
	if s.MinProperties != nil && mapLen < *s.MinProperties {
		return fmt.Errorf("object must have at least %d properties", *s.MinProperties)
	}
	if s.MaxProperties != nil && mapLen > *s.MaxProperties {
		return fmt.Errorf("object must have not more than %d properties", *s.MaxProperties)
	}

	if err := s.validateProperties(ctxPath, props); err != nil {
		return fmt.Errorf("validate properties: %w", err)
	}

	return nil
}

func (s *ObjectShape) inheritMinProperties(source *ObjectShape) error {
	if s.MinProperties == nil {
		s.MinProperties = source.MinProperties
	} else if source.MinProperties != nil && *s.MinProperties < *source.MinProperties {
		return StacktraceNew("minProperties constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *source.MinProperties),
			stacktrace.WithInfo("target", *s.MinProperties))
	}
	return nil
}

func (s *ObjectShape) inheritMaxProperties(source *ObjectShape) error {
	if s.MaxProperties == nil {
		s.MaxProperties = source.MaxProperties
	} else if source.MaxProperties != nil && *s.MaxProperties > *source.MaxProperties {
		return StacktraceNew("maxProperties constraint violation", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", *source.MaxProperties),
			stacktrace.WithInfo("target", *s.MaxProperties))
	}
	return nil
}

func (s *ObjectShape) inheritProperties(source *ObjectShape) error {
	if s.Properties == nil {
		s.Properties = source.Properties
		return nil
	}

	if source.Properties == nil {
		return nil
	}

	for pair := source.Properties.Oldest(); pair != nil; pair = pair.Next() {
		k, sourceProp := pair.Key, pair.Value
		if targetProp, present := s.Properties.Get(k); present {
			if sourceProp.Required && !targetProp.Required {
				return StacktraceNew("cannot make required property optional", s.Location,
					stacktrace.WithPosition(&targetProp.Base.Position),
					stacktrace.WithInfo("property", k),
					stacktrace.WithInfo("source", sourceProp.Required),
					stacktrace.WithInfo("target", targetProp.Required),
					stacktrace.WithType(StacktraceTypeUnwrapping))
			}
			_, err := targetProp.Base.Inherit(sourceProp.Base)
			if err != nil {
				return StacktraceNewWrapped("inherit property", err, s.Location,
					stacktrace.WithPosition(&targetProp.Base.Position),
					stacktrace.WithInfo("property", k),
					stacktrace.WithType(StacktraceTypeUnwrapping))
			}
		} else {
			s.Properties.Set(k, sourceProp)
		}
	}
	return nil
}

func (s *ObjectShape) inheritPatternProperties(source *ObjectShape) error {
	if s.PatternProperties == nil {
		s.PatternProperties = source.PatternProperties
		return nil
	}
	if source.PatternProperties != nil {
		for pair := source.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, sourceProp := pair.Key, pair.Value
			if targetProp, present := s.PatternProperties.Get(k); present {
				_, err := targetProp.Base.Inherit(sourceProp.Base)
				if err != nil {
					return StacktraceNewWrapped("inherit pattern property", err, s.Location,
						stacktrace.WithPosition(&targetProp.Base.Position),
						stacktrace.WithInfo("property", k),
						stacktrace.WithType(StacktraceTypeUnwrapping))
				}
			} else {
				s.PatternProperties.Set(k, sourceProp)
			}
		}
	}
	return nil
}

// Inherit merges the source shape into the target shape.
func (s *ObjectShape) inherit(source Shape) (Shape, error) {
	if ss, ok := source.(*RecursiveShape); ok {
		source = ss.Head.Shape
	}
	ss, ok := source.(*ObjectShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
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

	if err := s.inheritMinProperties(ss); err != nil {
		return nil, fmt.Errorf("inherit minProperties: %w", err)
	}

	if err := s.inheritMaxProperties(ss); err != nil {
		return nil, fmt.Errorf("inherit maxProperties: %w", err)
	}

	if err := s.inheritProperties(ss); err != nil {
		return nil, fmt.Errorf("inherit properties: %w", err)
	}

	if err := s.inheritPatternProperties(ss); err != nil {
		return nil, fmt.Errorf("inherit pattern properties: %w", err)
	}

	return s, nil
}

func (s *ObjectShape) checkPatternProperties() error {
	if s.PatternProperties == nil {
		return nil
	}
	if s.AdditionalProperties != nil && !*s.AdditionalProperties {
		// TODO: We actually can allow pattern properties with "additionalProperties: false" for stricter
		// 	validation.
		// This will contradict RAML 1.0 spec, but JSON Schema allows that.
		// https://json-schema.org/understanding-json-schema/reference/object#additionalproperties
		return StacktraceNew("pattern properties are not allowed with \"additionalProperties: false\"",
			s.Location, stacktrace.WithPosition(&s.Position))
	}
	for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		if err := prop.Base.Check(); err != nil {
			return StacktraceNewWrapped("check pattern property", err, s.Location,
				stacktrace.WithPosition(&prop.Base.Position),
				stacktrace.WithInfo("property", prop.Pattern.String()))
		}
	}
	return nil
}

func (s *ObjectShape) checkProperties() error {
	if s.Properties == nil {
		return nil
	}

	for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		if err := prop.Base.Check(); err != nil {
			return StacktraceNewWrapped("check property", err, s.Location,
				stacktrace.WithPosition(&prop.Base.Position),
				stacktrace.WithInfo("property", prop.Name))
		}
	}
	// FIXME: Need to validate on which level the discriminator is applied to avoid potential false positives.
	// Inline definitions with discriminator are not allowed.
	//nolint:nestif // Contains simple checks.
	if s.Discriminator != nil {
		prop, ok := s.Properties.Get(*s.Discriminator)
		if !ok {
			return StacktraceNew("discriminator property not found", s.Location,
				stacktrace.WithPosition(&s.Position),
				stacktrace.WithInfo("discriminator", *s.Discriminator))
		}
		if !prop.Base.IsScalar() {
			return StacktraceNew("discriminator property type must be a scalar", s.Location,
				stacktrace.WithPosition(&prop.Base.Position),
				stacktrace.WithInfo("discriminator", *s.Discriminator))
		}
		discriminatorValue := s.DiscriminatorValue
		// If discriminatorValue is set explicitly - validate it against the discriminator type
		if discriminatorValue != nil {
			if err := prop.Base.Validate(discriminatorValue); err != nil {
				return StacktraceNewWrapped("validate discriminator value", err, s.Location,
					stacktrace.WithPosition(&s.Base().Position),
					stacktrace.WithInfo("discriminator", *s.Discriminator))
			}
		}
	}

	return nil
}

func (s *ObjectShape) check() error {
	if s.MinProperties != nil && s.MaxProperties != nil && *s.MinProperties > *s.MaxProperties {
		return StacktraceNew("minProperties must be less than or equal to maxProperties",
			s.Location, stacktrace.WithPosition(&s.Position))
	}
	if err := s.checkPatternProperties(); err != nil {
		return fmt.Errorf("check pattern properties: %w", err)
	}
	if err := s.checkProperties(); err != nil {
		return fmt.Errorf("check properties: %w", err)
	}
	if s.Discriminator != nil && s.Properties == nil {
		return StacktraceNew("discriminator without properties", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	return nil
}

// makeProperty creates a pattern property from a YAML node.
func (r *RAML) makePatternProperty(nodeName string, propertyName string, v *yaml.Node, location string,
	hasImplicitOptional bool) (PatternProperty, error) {
	shape, err := r.makeNewShapeYAML(v, nodeName, location)
	if err != nil {
		return PatternProperty{}, StacktraceNewWrapped("make shape", err, location,
			WithNodePosition(v))
	}
	// Pattern properties cannot be required
	if shape.Required != nil || hasImplicitOptional {
		return PatternProperty{}, StacktraceNew("'required' facet is not supported on pattern property",
			location, WithNodePosition(v))
	}
	re, err := regexp.Compile(propertyName[1 : len(propertyName)-1])
	if err != nil {
		return PatternProperty{}, StacktraceNewWrapped("compile pattern", err, location, WithNodePosition(v))
	}
	return PatternProperty{
		Pattern: re,
		Base:    shape,
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
func (r *RAML) makeProperty(nodeName string, propertyName string, v *yaml.Node,
	location string, hasImplicitOptional bool) (Property, error) {
	shape, err := r.makeNewShapeYAML(v, nodeName, location)
	if err != nil {
		return Property{}, StacktraceNewWrapped("make shape", err, location, WithNodePosition(v))
	}
	finalName := propertyName
	var required bool
	shapeRequired := shape.Required
	if shapeRequired == nil {
		// If shape has no "required" facet, requirement depends only on whether "?"" was used in node name.
		required = !hasImplicitOptional
	} else {
		// If shape explicitly defines "required" facet combined with "?" in node name - explicit
		// definition prevails and property name keeps the node name.
		// Otherwise, keep propertyName that has the last "?" chomped.
		if hasImplicitOptional {
			finalName = nodeName
		}
		required = *shapeRequired
	}
	return Property{
		Name:     finalName,
		Base:     shape,
		Required: required,
		raml:     r,
	}, nil
}

// Property represents a property of an object shape.
type Property struct {
	Name     string
	Base     *BaseShape
	Required bool
	raml     *RAML
}

// Property represents a pattern property of an object shape.
type PatternProperty struct {
	Pattern *regexp.Regexp
	Base    *BaseShape
	// Pattern properties are always optional.
	raml *RAML
}

// UnionFacets contains constraints for union shapes.
type UnionFacets struct {
	AnyOf []*BaseShape
}

// UnionShape represents a union shape.
type UnionShape struct {
	noScalarShape
	*BaseShape

	EnumFacets
	UnionFacets
}

// UnmarshalYAMLNodes unmarshals the union shape from YAML nodes.
func (s *UnionShape) unmarshalYAMLNodes(_ []*yaml.Node) error {
	return nil
}

// Base returns the base shape.
func (s *UnionShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *UnionShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *UnionShape) clone(base *BaseShape, clonedMap map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	c.AnyOf = make([]*BaseShape, len(s.AnyOf))
	for i, member := range s.AnyOf {
		c.AnyOf[i] = member.clone(clonedMap)
	}
	return &c
}

func (s *UnionShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*UnionShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.AnyOf = ss.AnyOf
	s.Enum = ss.Enum
	return s, nil
}

func (s *UnionShape) validate(v interface{}, ctxPath string) error {
	st := StacktraceNew("value does not match any type", s.Location,
		stacktrace.WithPosition(&s.Position))
	var err error

	for _, item := range s.AnyOf {
		err = item.Shape.validate(v, ctxPath)
		if err == nil {
			return nil
		}
		st = st.Append(
			StacktraceNewWrapped(
				"validate union member",
				err,
				s.Location,
				stacktrace.WithPosition(&item.Position),
				stacktrace.WithInfo("type", item.Type),
				stacktrace.WithInfo("name", item.Name),
				stacktrace.WithInfo("value", v),
			),
		)
	}
	return st
}

// inherit merges the source shape into the target shape.
func (s *UnionShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*UnionShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	if len(s.AnyOf) == 0 {
		s.AnyOf = ss.AnyOf
		return s, nil
	}
	// TODO: Implement enum facets inheritance
	var finalFiltered []*BaseShape
	for _, sourceMember := range ss.AnyOf {
		var filtered []*BaseShape
		for _, targetMember := range s.AnyOf {
			if sourceMember.Type == targetMember.Type {
				// Clone is required to avoid modifying the original target member shape.
				cs := targetMember.CloneDetached()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.ID = s.raml.generateShapeID()
				ms, err := cs.Inherit(sourceMember)
				if err != nil {
					// TODO: Collect errors
					// StacktraceNewWrapped("merge union member", err, s.Location)
					continue
				}
				filtered = append(filtered, ms)
			}
		}
		if len(filtered) == 0 {
			return nil, StacktraceNew("failed to find compatible union member", s.Location,
				stacktrace.WithPosition(&s.Position))
		}
		finalFiltered = append(finalFiltered, filtered...)
	}
	s.AnyOf = finalFiltered
	return s, nil
}

func (s *UnionShape) check() error {
	for _, item := range s.AnyOf {
		if err := item.Check(); err != nil {
			return StacktraceNewWrapped("check union member", err, s.Location,
				stacktrace.WithPosition(&item.Position))
		}
	}
	return nil
}

type JSONShape struct {
	noScalarShape
	*BaseShape

	Schema    *JSONSchema
	Validator *gojsonschema.Schema
	Raw       string
}

func (s *JSONShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *JSONShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *JSONShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *JSONShape) validate(v interface{}, _ string) error {
	l := gojsonschema.NewGoLoader(v)

	result, err := s.Validator.Validate(l)
	if err != nil {
		return StacktraceNewWrapped("validate JSON schema", err, s.Location, stacktrace.WithPosition(&s.Position))
	}

	if !result.Valid() {
		st := StacktraceNew("failed to validate against JSON schema", s.Location, stacktrace.WithPosition(&s.Position))
		for _, err := range result.Errors() {
			st = st.Append(StacktraceNew(err.String(), s.Location, stacktrace.WithPosition(&s.Position)))
		}
		return st
	}
	return nil
}

func (s *JSONShape) unmarshalYAMLNodes(_ []*yaml.Node) error {
	return nil
}

func (s *JSONShape) inherit(source Shape) (Shape, error) {
	ss, ok := source.(*JSONShape)
	if !ok {
		return nil, StacktraceNew("cannot inherit from different type", s.Location,
			stacktrace.WithPosition(&s.Position), stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	// TODO: Check if the schemas are different more strictly
	if s.Raw != "" && ss.Raw != "" && s.Raw != ss.Raw {
		return nil, StacktraceNew("cannot inherit from different JSON schema", s.Location,
			stacktrace.WithPosition(&s.Position))
	}
	s.Schema = ss.Schema
	s.Raw = ss.Raw
	s.Validator = ss.Validator
	return s, nil
}

func (s *JSONShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*JSONShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Schema = ss.Schema
	s.Raw = ss.Raw
	return s, nil
}

func (s *JSONShape) check() error {
	// TODO: JSON Schema check
	return nil
}

type UnknownShape struct {
	noScalarShape
	*BaseShape

	facets []*yaml.Node
}

func (s *UnknownShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *UnknownShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *UnknownShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *UnknownShape) validate(_ interface{}, _ string) error {
	return StacktraceNew("cannot validate against unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *UnknownShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	s.facets = v
	return nil
}

func (s *UnknownShape) inherit(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot inherit from unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *UnknownShape) alias(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot alias from unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *UnknownShape) check() error {
	return StacktraceNew("cannot check unknown shape", s.Location, stacktrace.WithPosition(&s.Position))
}

type RecursiveShape struct {
	noScalarShape
	*BaseShape

	Head *BaseShape
}

func (s *RecursiveShape) unmarshalYAMLNodes(_ []*yaml.Node) error {
	return nil
}

func (s *RecursiveShape) Base() *BaseShape {
	return s.BaseShape
}

func (s *RecursiveShape) cloneShallow(base *BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *RecursiveShape) clone(base *BaseShape, _ map[int64]*BaseShape) Shape {
	c := *s
	c.BaseShape = base
	return &c
}

func (s *RecursiveShape) alias(source Shape) (Shape, error) {
	ss, ok := source.(*RecursiveShape)
	if !ok {
		return nil, StacktraceNew("cannot make alias from different type", s.Location,
			stacktrace.WithPosition(&s.Position),
			stacktrace.WithInfo("source", source.Base().Type),
			stacktrace.WithInfo("target", s.Base().Type))
	}
	s.Head = ss.Head
	return s, nil
}

func (s *RecursiveShape) validate(v interface{}, ctxPath string) error {
	if err := s.Head.Shape.validate(v, ctxPath); err != nil {
		return fmt.Errorf("validate recursive shape: %w", err)
	}
	return nil
}

// Inherit merges the source shape into the target shape.
func (s *RecursiveShape) inherit(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot inherit from recursive shape", s.Location, stacktrace.WithPosition(&s.Position))
}

func (s *RecursiveShape) check() error {
	return nil
}
