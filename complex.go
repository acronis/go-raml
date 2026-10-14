package raml

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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
	MinItems    *ScalarFacet[uint64]
	MaxItems    *ScalarFacet[uint64]
	UniqueItems *ScalarFacet[bool]
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
	ss, err := checkAliasType[*ArrayShape](s, source)
	if err != nil {
		return nil, err
	}
	s.Items = ss.Items
	s.MinItems = ss.MinItems
	s.MaxItems = ss.MaxItems
	s.UniqueItems = ss.UniqueItems
	return s, nil
}

func (s *ArrayShape) validate(v any, ctxPath string) error {
	i, ok := v.([]any)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected []any", v)
	}

	arrayLen := uint64(len(i))
	if s.MinItems != nil && arrayLen < s.MinItems.Value {
		return fmt.Errorf("array must have at least %d items", s.MinItems.Value)
	}
	if s.MaxItems != nil && arrayLen > s.MaxItems.Value {
		return fmt.Errorf("array must have not more than %d items", s.MaxItems.Value)
	}
	if s.Items != nil {
		for ii, item := range i {
			ctxPathA := ctxPath + "[" + strconv.Itoa(ii) + "]"
			if err := s.Items.validateAt(item, ctxPathA); err != nil {
				return withPath(ctxPathA, fmt.Errorf("validate array item %s: %w", ctxPathA, err))
			}
		}
	}
	if s.UniqueItems != nil && s.UniqueItems.Value {
		if j, k := duplicateItems(i); j >= 0 {
			return fmt.Errorf("array contains duplicate items at indices %d and %d", j, k)
		}
	}

	return nil
}

// Inherit merges the source shape into the target shape.
func (s *ArrayShape) inherit(source Shape) (Shape, error) {
	ss, err := checkInheritType[*ArrayShape](s, source)
	if err != nil {
		return nil, err
	}
	if s.Items == nil {
		s.Items = ss.Items
	} else if ss.Items != nil {
		_, err := s.Items.Inherit(ss.Items)
		if err != nil {
			return nil, StacktraceNewWrapped("merge array items", err, s.Location,
				stacktrace.WithPosition(&s.Items.KeyPos))
		}
	}
	if s.MinItems == nil {
		s.MinItems = ss.MinItems
	} else if ss.MinItems != nil && s.MinItems.Value < ss.MinItems.Value {
		return nil, StacktraceNew("minItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.MinItems.ValuePos), stacktrace.WithInfo("source", ss.MinItems.Value),
			stacktrace.WithInfo("target", s.MinItems.Value))
	}
	if s.MaxItems == nil {
		s.MaxItems = ss.MaxItems
	} else if ss.MaxItems != nil && s.MaxItems.Value > ss.MaxItems.Value {
		return nil, StacktraceNew("maxItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.MaxItems.ValuePos), stacktrace.WithInfo("source", ss.MaxItems.Value),
			stacktrace.WithInfo("target", s.MaxItems.Value))
	}
	if s.UniqueItems == nil {
		s.UniqueItems = ss.UniqueItems
	} else if ss.UniqueItems != nil && ss.UniqueItems.Value && !s.UniqueItems.Value {
		return nil, StacktraceNew("uniqueItems constraint violation", s.Location,
			stacktrace.WithPosition(&s.UniqueItems.ValuePos), stacktrace.WithInfo("source", ss.UniqueItems.Value),
			stacktrace.WithInfo("target", s.UniqueItems.Value))
	}
	return s, nil
}

func (s *ArrayShape) check() error {
	if s.MinItems != nil && s.MaxItems != nil && s.MinItems.Value > s.MaxItems.Value {
		return StacktraceNew("minItems must be less than or equal to maxItems", s.Location,
			stacktrace.WithPosition(&s.MinItems.ValuePos))
	}
	if s.Items != nil {
		if err := s.Items.Check(); err != nil {
			return StacktraceNewWrapped("check items", err, s.Location,
				stacktrace.WithPosition(&s.Items.KeyPos))
		}
	}
	return nil
}

func (s *ArrayShape) String() string {
	var facets []string
	if s.Items != nil {
		facets = append(facets, fmt.Sprintf("items:%s", s.Items.Type))
	}
	if s.MinItems != nil {
		facets = append(facets, fmt.Sprintf("minItems:%d", s.MinItems.Value))
	}
	if s.MaxItems != nil {
		facets = append(facets, fmt.Sprintf("maxItems:%d", s.MaxItems.Value))
	}
	if s.UniqueItems != nil && s.UniqueItems.Value {
		facets = append(facets, "uniqueItems:true")
	}
	return fmt.Sprintf("ArrayShape{facets:[%s]}", strings.Join(facets, ","))
}

// UnmarshalYAMLNodes unmarshals the array shape from YAML nodes.
func (s *ArrayShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	for i := 0; i != len(v); i += 2 {
		keyNode := v[i]
		valueNode := v[i+1]
		switch keyNode.Value {
		case FacetMinItems:
			node, err := MakeScalarFacetYAML[uint64](s.raml, keyNode, valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
			}
			s.MinItems = node
		case FacetMaxItems:
			node, err := MakeScalarFacetYAML[uint64](s.raml, keyNode, valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
			}
			s.MaxItems = node
		case FacetItems:
			if valueNode.Kind != yaml.ScalarNode && valueNode.Kind != yaml.MappingNode {
				return StacktraceNew("items must be either a reference or an inline type", s.Location,
					WithNodePosition(valueNode))
			}
			shape, err := s.raml.makeNewShapeYAML(keyNode, valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make shape", err, s.Location,
					WithNodePosition(valueNode),
					stacktrace.WithInfo("facet", FacetItems))
			}
			s.Items = shape
		case FacetUniqueItems:
			node, err := MakeScalarFacetYAML[bool](s.raml, keyNode, valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
			}
			s.UniqueItems = node
		default:
			n, err := s.raml.makeRootNode(keyNode, valueNode, s.Location)
			if err != nil {
				return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
			}
			s.CustomShapeFacets.Set(keyNode.Value, n)
		}
	}
	return nil
}

// ObjectFacets contains constraints for object shapes.
type ObjectFacets struct {
	Discriminator        *ScalarFacet[string]
	DiscriminatorValue   *DataNode
	AdditionalProperties *ScalarFacet[bool]
	Properties           *orderedmap.OrderedMap[string, Property]
	PatternProperties    *orderedmap.OrderedMap[string, PatternProperty]
	MinProperties        *ScalarFacet[uint64]
	MaxProperties        *ScalarFacet[uint64]
}

// ObjectShape represents an object shape.
type ObjectShape struct {
	noScalarShape
	*BaseShape

	ObjectFacets
}

func (s *ObjectShape) unmarshalPatternProperties(propertyName string, keyNode, valueNode *yaml.Node, hasImplicitOptional bool) error {
	if s.PatternProperties == nil {
		s.PatternProperties = orderedmap.New[string, PatternProperty]()
	}
	property, err := s.raml.makePatternProperty(propertyName, keyNode, valueNode, s.Location,
		hasImplicitOptional)
	if err != nil {
		return StacktraceNewWrapped("make pattern property", err, s.Location,
			WithNodePosition(keyNode))
	}
	s.PatternProperties.Set(propertyName, property)
	return nil
}

func (s *ObjectShape) unmarshalProperty(keyNode, valueNode *yaml.Node) error {
	propertyName, hasImplicitOptional := chompImplicitOptional(keyNode.Value)
	if len(propertyName) > 1 && propertyName[0] == '/' && propertyName[len(propertyName)-1] == '/' {
		return s.unmarshalPatternProperties(propertyName, keyNode, valueNode, hasImplicitOptional)
	}

	if s.Properties == nil {
		s.Properties = orderedmap.New[string, Property]()
	}
	property, err := s.raml.makeProperty(keyNode, valueNode, propertyName, s.Location, hasImplicitOptional)
	if err != nil {
		return StacktraceNewWrapped("make property", err, s.Location, WithNodePosition(keyNode))
	}
	s.Properties.Set(property.Name, property)
	return nil
}

func (s *ObjectShape) unmarshalYAMLNode(keyNode, valueNode *yaml.Node) error {
	switch keyNode.Value {
	case FacetAdditionalProperties:
		node, err := MakeScalarFacetYAML[bool](s.raml, keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
		}
		s.AdditionalProperties = node
	case FacetDiscriminator:
		node, err := MakeScalarFacetYAML[string](s.raml, keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
		}
		s.Discriminator = node
	case FacetDiscriminatorValue:
		n, err := s.raml.makeRootNode(keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("decode", err, s.Location,
				WithNodePosition(valueNode),
				stacktrace.WithInfo("facet", FacetDiscriminatorValue))
		}
		s.DiscriminatorValue = n
	case FacetMinProperties:
		node, err := MakeScalarFacetYAML[uint64](s.raml, keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
		}
		s.MinProperties = node
	case FacetMaxProperties:
		node, err := MakeScalarFacetYAML[uint64](s.raml, keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make scalar node", err, s.Location, WithNodePosition(valueNode))
		}
		s.MaxProperties = node
	case FacetProperties:
		if valueNode.Kind != yaml.MappingNode && valueNode.Tag != TagNull {
			return StacktraceNew("properties must be a mapping", s.Location,
				WithNodePosition(valueNode), stacktrace.WithInfo("facet", FacetProperties))
		}
		for j := 0; j != len(valueNode.Content); j += 2 {
			keyNode := valueNode.Content[j]
			valueNode := valueNode.Content[j+1]

			if err := s.unmarshalProperty(keyNode, valueNode); err != nil {
				return fmt.Errorf("unmarshal property: %w", err)
			}
		}
	default:
		n, err := s.raml.makeRootNode(keyNode, valueNode, s.Location)
		if err != nil {
			return StacktraceNewWrapped("make node", err, s.Location, WithNodePosition(valueNode))
		}
		s.CustomShapeFacets.Set(keyNode.Value, n)
	}
	return nil
}

// UnmarshalYAMLNodes unmarshals the object shape from YAML nodes.
func (s *ObjectShape) unmarshalYAMLNodes(v []*yaml.Node) error {
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
	if l := s.Properties.Len(); l > 0 {
		c.Properties = orderedmap.New[string, Property](l)
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			c.Properties.Set(pair.Key, Property{
				Name:     pair.Value.Name,
				Base:     pair.Value.Base.clone(clonedMap),
				Required: pair.Value.Required,
			})
		}
	}
	if l := s.PatternProperties.Len(); l > 0 {
		c.PatternProperties = orderedmap.New[string, PatternProperty](l)
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			c.PatternProperties.Set(pair.Key, PatternProperty{
				Pattern: pair.Value.Pattern,
				Base:    pair.Value.Base.clone(clonedMap),
			})
		}
	}
	return &c
}

func (s *ObjectShape) alias(source Shape) (Shape, error) {
	ss, err := checkAliasType[*ObjectShape](s, source)
	if err != nil {
		return nil, err
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
	item any,
	ctxPathK string,
) (bool, error) {
	if s.PatternProperties == nil {
		return false, nil
	}
	for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		pp := &pair.Value
		if !pp.Pattern.MatchString(k) {
			continue
		}
		// NOTE: The first defined pattern property to validate prevails.
		err := pp.Base.validateAt(item, ctxPathK)
		if err != nil {
			return true, StacktraceNewWrapped("validate pattern property", err, s.Location,
				stacktrace.WithPosition(&pp.Base.KeyPos),
				stacktrace.WithInfo("property", pp.Pattern.String()))
		}
		return true, nil
	}
	return false, nil
}

func (s *ObjectShape) validateProperty(
	k string,
	item any,
	ctxPathK string,
) (bool, error) {
	if s.Properties == nil {
		return false, nil
	}
	p, present := s.Properties.Get(k)
	if !present {
		return false, nil
	}
	if err := p.Base.validateAt(item, ctxPathK); err != nil {
		return true, withPath(ctxPathK, fmt.Errorf("validate property %s: %w", ctxPathK, err))
	}
	return true, nil
}

func (s *ObjectShape) validateProperties(ctxPath string, props map[string]any) error {
	if missing := s.findMissingRequired(props); len(missing) > 0 {
		return withPath(ctxPath, fmt.Errorf("missing required properties: %s", strings.Join(missing, ", ")))
	}

	restrictedAdditionalProperties := s.AdditionalProperties != nil && !s.AdditionalProperties.Value
	// Validate schema-defined properties in schema order (deterministic, no allocation).
	for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
		item, present := props[pair.Key]
		if !present {
			continue
		}
		ctxPathK := ctxPath + "." + pair.Key
		if err := pair.Value.Base.validateAt(item, ctxPathK); err != nil {
			return withPath(ctxPathK, fmt.Errorf("validate property %s: %w", ctxPathK, err))
		}
	}
	// Check for additional or pattern properties not defined in the schema.
	for k, item := range props {
		if s.Properties != nil {
			if _, inSchema := s.Properties.Get(k); inSchema {
				continue
			}
		}
		// Explicitly defined properties have priority over pattern properties.
		ctxPathK := ctxPath + "." + k
		if restrictedAdditionalProperties {
			return withPath(ctxPathK, fmt.Errorf("unexpected additional property \"%s\"", k))
		}

		_, err := s.validatePatternProperty(k, item, ctxPathK)
		if err != nil {
			return withPath(ctxPathK, fmt.Errorf("validate pattern property %s: %w", ctxPathK, err))
		}
	}
	return nil
}

func (s *ObjectShape) findMissingRequired(props map[string]any) []string {
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

func (s *ObjectShape) validate(v any, ctxPath string) error {
	props, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid type, got %T, expected map[string]any", v)
	}

	mapLen := uint64(len(props))
	if s.MinProperties != nil && mapLen < s.MinProperties.Value {
		return fmt.Errorf("object must have at least %d properties", s.MinProperties.Value)
	}
	if s.MaxProperties != nil && mapLen > s.MaxProperties.Value {
		return fmt.Errorf("object must have not more than %d properties", s.MaxProperties.Value)
	}

	if err := s.validateProperties(ctxPath, props); err != nil {
		return fmt.Errorf("validate properties: %w", err)
	}

	return nil
}

func (s *ObjectShape) inheritMinProperties(source *ObjectShape) error {
	if s.MinProperties == nil {
		s.MinProperties = source.MinProperties
	} else if source.MinProperties != nil && s.MinProperties.Value < source.MinProperties.Value {
		return StacktraceNew("minProperties constraint violation", s.Location,
			stacktrace.WithPosition(&s.MinProperties.ValuePos),
			stacktrace.WithInfo("source", source.MinProperties.Value),
			stacktrace.WithInfo("target", s.MinProperties.Value))
	}
	return nil
}

func (s *ObjectShape) inheritMaxProperties(source *ObjectShape) error {
	if s.MaxProperties == nil {
		s.MaxProperties = source.MaxProperties
	} else if source.MaxProperties != nil && s.MaxProperties.Value > source.MaxProperties.Value {
		return StacktraceNew("maxProperties constraint violation", s.Location,
			stacktrace.WithPosition(&s.MaxProperties.ValuePos),
			stacktrace.WithInfo("source", source.MaxProperties.Value),
			stacktrace.WithInfo("target", s.MaxProperties.Value))
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
		k := pair.Key
		if targetProp, present := s.Properties.Get(k); present {
			sourceProp := &pair.Value
			if sourceProp.Required && !targetProp.Required {
				return StacktraceNew("cannot make required property optional", s.Location,
					stacktrace.WithPosition(&targetProp.Base.KeyPos),
					stacktrace.WithInfo("property", k),
					stacktrace.WithInfo("source", sourceProp.Required),
					stacktrace.WithInfo("target", targetProp.Required),
					stacktrace.WithType(StacktraceTypeUnwrapping))
			}
			_, err := targetProp.Base.Inherit(sourceProp.Base)
			if err != nil {
				return StacktraceNewWrapped("inherit property", err, s.Location,
					stacktrace.WithPosition(&targetProp.Base.KeyPos),
					stacktrace.WithInfo("property", k),
					stacktrace.WithType(StacktraceTypeUnwrapping))
			}
		} else {
			s.Properties.Set(k, pair.Value)
		}
	}
	return nil
}

func (s *ObjectShape) inheritPatternProperties(source *ObjectShape) error {
	if s.PatternProperties == nil {
		s.PatternProperties = source.PatternProperties
		return nil
	}
	for pair := source.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		k := pair.Key
		if targetProp, present := s.PatternProperties.Get(k); present {
			sourceProp := &pair.Value
			_, err := targetProp.Base.Inherit(sourceProp.Base)
			if err != nil {
				return StacktraceNewWrapped("inherit pattern property", err, s.Location,
					stacktrace.WithPosition(&targetProp.Base.KeyPos),
					stacktrace.WithInfo("property", k),
					stacktrace.WithType(StacktraceTypeUnwrapping))
			}
		} else {
			s.PatternProperties.Set(k, pair.Value)
		}
	}
	return nil
}

// Inherit merges the source shape into the target shape.
func (s *ObjectShape) inherit(source Shape) (Shape, error) {
	if ss, ok := source.(*RecursiveShape); ok {
		source = ss.Head.Shape
	}
	ss, err := checkInheritType[*ObjectShape](s, source)
	if err != nil {
		return nil, err
	}

	// Discriminator and AdditionalProperties are inherited as is
	if s.AdditionalProperties == nil {
		s.AdditionalProperties = ss.AdditionalProperties
	}
	if s.Discriminator == nil {
		s.Discriminator = ss.Discriminator
	}

	if err := s.inheritMinProperties(ss); err != nil {
		return nil, StacktraceNewWrapped("inherit minProperties", err, s.Location, stacktrace.WithPosition(&s.KeyPos))
	}

	if err := s.inheritMaxProperties(ss); err != nil {
		return nil, StacktraceNewWrapped("inherit maxProperties", err, s.Location, stacktrace.WithPosition(&s.KeyPos))
	}

	if err := s.inheritProperties(ss); err != nil {
		return nil, StacktraceNewWrapped("inherit properties", err, s.Location, stacktrace.WithPosition(&s.KeyPos))
	}

	if err := s.inheritPatternProperties(ss); err != nil {
		return nil, StacktraceNewWrapped("inherit pattern properties", err, s.Location, stacktrace.WithPosition(&s.KeyPos))
	}

	return s, nil
}

func (s *ObjectShape) checkPatternProperties() error {
	if s.PatternProperties == nil {
		return nil
	}
	if s.AdditionalProperties != nil && !s.AdditionalProperties.Value {
		// TODO: We actually can allow pattern properties with "additionalProperties: false" for stricter
		// 	validation.
		// This will contradict RAML 1.0 spec, but JSON Schema allows that.
		// https://json-schema.org/understanding-json-schema/reference/object#additionalproperties
		return StacktraceNew("pattern properties are not allowed with \"additionalProperties: false\"",
			s.Location, stacktrace.WithPosition(&s.AdditionalProperties.ValuePos))
	}
	for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
		prop := &pair.Value
		if err := prop.Base.Check(); err != nil {
			return StacktraceNewWrapped("check pattern property", err, s.Location,
				stacktrace.WithPosition(&prop.Base.KeyPos),
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
		prop := &pair.Value
		if err := prop.Base.Check(); err != nil {
			return StacktraceNewWrapped("check property", err, s.Location,
				stacktrace.WithPosition(&prop.Base.KeyPos),
				stacktrace.WithInfo("property", prop.Name))
		}
	}
	// FIXME: Need to validate on which level the discriminator is applied to avoid potential false positives.
	// Inline definitions with discriminator are not allowed.
	//nolint:nestif // Contains simple checks.
	if s.Discriminator != nil {
		prop, ok := s.Properties.Get(s.Discriminator.Value)
		if !ok {
			return StacktraceNew("discriminator property not found", s.Location,
				stacktrace.WithPosition(&s.Discriminator.ValuePos),
				stacktrace.WithInfo("discriminator", s.Discriminator.Value))
		}
		if !prop.Base.IsScalar() {
			return StacktraceNew("discriminator property type must be a scalar", s.Location,
				stacktrace.WithPosition(&prop.Base.KeyPos),
				stacktrace.WithInfo("discriminator", s.Discriminator.Value))
		}
		discriminatorValue := s.DiscriminatorValue
		// If discriminatorValue is set explicitly - validate it against the discriminator type
		if discriminatorValue != nil {
			if err := prop.Base.Validate(discriminatorValue.Value.Raw); err != nil {
				return StacktraceNewWrapped("validate discriminator value", err, s.Location,
					stacktrace.WithPosition(&s.Base().KeyPos),
					stacktrace.WithInfo("discriminator", s.Discriminator.Value))
			}
		}
	}

	return nil
}

func (s *ObjectShape) check() error {
	if s.MinProperties != nil && s.MaxProperties != nil && s.MinProperties.Value > s.MaxProperties.Value {
		return StacktraceNew("minProperties must be less than or equal to maxProperties",
			s.Location, stacktrace.WithPosition(&s.MinProperties.ValuePos))
	}
	if err := s.checkPatternProperties(); err != nil {
		return fmt.Errorf("check pattern properties: %w", err)
	}
	if err := s.checkProperties(); err != nil {
		return fmt.Errorf("check properties: %w", err)
	}
	if s.Discriminator != nil && s.Properties == nil {
		return StacktraceNew("discriminator without properties", s.Location,
			stacktrace.WithPosition(&s.Discriminator.KeyPos))
	}
	return nil
}

func (s *ObjectShape) String() string {
	var facets []string
	if l := s.Properties.Len(); l > 0 {
		facets = append(facets, fmt.Sprintf("properties:%d", l))
	}
	if l := s.PatternProperties.Len(); l > 0 {
		facets = append(facets, fmt.Sprintf("patternProperties:%d", l))
	}
	if s.AdditionalProperties != nil {
		facets = append(facets, fmt.Sprintf("additionalProperties:%v", s.AdditionalProperties.Value))
	}
	if s.MinProperties != nil {
		facets = append(facets, fmt.Sprintf("minProperties:%d", s.MinProperties.Value))
	}
	if s.MaxProperties != nil {
		facets = append(facets, fmt.Sprintf("maxProperties:%d", s.MaxProperties.Value))
	}
	return fmt.Sprintf("ObjectShape{facets:[%s]}", strings.Join(facets, ","))
}

// makeProperty creates a pattern property from a YAML node.
func (r *RAML) makePatternProperty(propertyName string, keyNode, valueNode *yaml.Node, location string,
	hasImplicitOptional bool) (PatternProperty, error) {
	shape, err := r.makeNewShapeYAML(keyNode, valueNode, location)
	if err != nil {
		return PatternProperty{}, StacktraceNewWrapped("make shape", err, location,
			WithNodePosition(keyNode))
	}
	// Pattern properties cannot be required
	if shape.Required != nil || hasImplicitOptional {
		return PatternProperty{}, StacktraceNew("'required' facet is not supported on pattern property",
			location, WithNodePosition(keyNode))
	}
	re, err := regexp.Compile(propertyName[1 : len(propertyName)-1])
	if err != nil {
		return PatternProperty{}, StacktraceNewWrapped("compile pattern", err, location, WithNodePosition(keyNode))
	}
	return PatternProperty{
		Pattern: re,
		Base:    shape,
	}, nil
}

// makeProperty creates a property from a YAML node.
func (r *RAML) makeProperty(
	keyNode, valueNode *yaml.Node,
	propertyName string,
	location string,
	hasImplicitOptional bool,
) (Property, error) {
	shape, err := r.makeNewShapeYAML(keyNode, valueNode, location)
	if err != nil {
		return Property{}, StacktraceNewWrapped("make shape", err, location, WithNodePosition(keyNode))
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
			finalName = keyNode.Value
		}
		required = shapeRequired.Value
	}
	return Property{
		Name:     finalName,
		Base:     shape,
		Required: required,
	}, nil
}

// Property represents a property of an object shape.
type Property struct {
	Name     string
	Base     *BaseShape
	Required bool
}

// Property represents a pattern property of an object shape.
type PatternProperty struct {
	Pattern *regexp.Regexp
	Base    *BaseShape
	// Pattern properties are always optional.
}

// UnionFacets contains constraints for union shapes.
type UnionFacets struct {
	AnyOf []*BaseShape
}

// UnionShape represents a union shape.
type UnionShape struct {
	noScalarShape
	*BaseShape

	UnionFacets
}

// UnmarshalYAMLNodes unmarshals the union shape from YAML nodes.
func (s *UnionShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	// RAML 1.0 spec § "Using Discriminator":
	// "Neither discriminator nor discriminatorValue can be defined for any
	// inline type declarations or union types."
	for i := 0; i+1 < len(v); i += 2 {
		k := v[i]
		switch k.Value {
		case FacetDiscriminator, FacetDiscriminatorValue:
			return StacktraceNew("discriminator is not allowed on union types",
				s.Location, WithNodePosition(k),
				stacktrace.WithInfo("facet", k.Value))
		}
	}
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
	ss, err := checkAliasType[*UnionShape](s, source)
	if err != nil {
		return nil, err
	}
	s.AnyOf = ss.AnyOf
	s.Enum = ss.Enum
	return s, nil
}

func (s *UnionShape) validate(v any, ctxPath string) error {
	var acc stacktrace.Accumulator
	for _, item := range s.AnyOf {
		err := item.validateAt(v, ctxPath)
		if err == nil {
			return nil
		}
		acc.Add(StacktraceNewWrapped(
			"validate union member",
			err,
			s.Location,
			stacktrace.WithPosition(&item.KeyPos),
			stacktrace.WithInfo("type", item.Type),
			stacktrace.WithInfo("name", item.Name),
			stacktrace.WithInfo("value", v),
		))
	}
	se := StacktraceNew("value does not match any type", s.Location,
		stacktrace.WithPosition(&s.KeyPos))
	if details := acc.Result(); details != nil {
		se = se.Append(details)
	}
	return se
}

// inherit merges the source shape into the target shape.
func (s *UnionShape) inherit(source Shape) (Shape, error) {
	ss, err := checkInheritType[*UnionShape](s, source)
	if err != nil {
		return nil, err
	}
	if len(s.AnyOf) == 0 {
		s.AnyOf = ss.AnyOf
		return s, nil
	}
	var finalFiltered []*BaseShape
	for _, sourceMember := range ss.AnyOf {
		var filtered []*BaseShape
		for _, targetMember := range s.AnyOf {
			if sourceMember.Type == targetMember.Type {
				// Clone is required to avoid modifying the original target member shape.
				cs := targetMember.CloneDetached()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.ID = s.raml.generateSequenceID()
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
				stacktrace.WithPosition(&s.ValuePos))
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
				stacktrace.WithPosition(&item.KeyPos))
		}
	}
	return nil
}

func (s *UnionShape) String() string {
	return fmt.Sprintf("UnionShape{anyOf:%d}", len(s.AnyOf))
}

type JSONShape struct {
	noScalarShape
	*BaseShape

	cachedShape Shape
	cachedDefs  *orderedmap.OrderedMap[string, Shape]
	Validator   *jsonschema.Schema
	Raw         string
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

func (s *JSONShape) validate(v any, _ string) error {
	if err := s.Validator.Validate(v); err != nil {
		return StacktraceNewWrapped("failed to validate against JSON schema", err,
			s.Location, stacktrace.WithPosition(&s.ValuePos))
	}
	return nil
}

func (s *JSONShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	if len(v) > 0 {
		return StacktraceNew("type-specific are not allowed for JSON external types",
			s.Location, WithNodePosition(v[0]))
	}
	return nil
}

func (s *JSONShape) inherit(source Shape) (Shape, error) {
	ss, err := checkInheritType[*JSONShape](s, source)
	if err != nil {
		return nil, err
	}
	// TODO: Check if the schemas are different more strictly
	if s.Raw != "" && ss.Raw != "" && s.Raw != ss.Raw {
		return nil, StacktraceNew("cannot inherit from different JSON schema", s.Location,
			stacktrace.WithPosition(&s.ValuePos))
	}
	s.cachedShape = nil
	s.cachedDefs = nil
	s.Raw = ss.Raw
	s.Validator = ss.Validator
	return s, nil
}

func (s *JSONShape) alias(source Shape) (Shape, error) {
	ss, err := checkAliasType[*JSONShape](s, source)
	if err != nil {
		return nil, err
	}
	s.cachedShape = nil
	s.cachedDefs = nil
	s.Raw = ss.Raw
	s.Validator = ss.Validator
	return s, nil
}

func (s *JSONShape) check() error {
	// TODO: JSON Schema check
	return nil
}

func (s *JSONShape) String() string {
	return "JSONShape{schema:present}"
}

// AsShape converts the JSON Schema embedded in this JSONShape into a native RAML
// Shape. Returns (nil, nil) when the validator is nil.
// Named $ref targets are accumulated as a side-effect; call AsShapeDefs() to
// retrieve them after AsShape() returns.
func (s *JSONShape) AsShape() (Shape, error) {
	if s.Validator == nil {
		return nil, nil
	}
	if s.cachedShape != nil {
		return s.cachedShape, nil
	}
	defs := orderedmap.New[string, Shape](0)
	shape, err := jsonSchemaToShape(s.BaseShape.raml, s.BaseShape, s.Validator,
		make(map[*jsonschema.Schema]*BaseShape), defs)
	if err != nil {
		return nil, err
	}
	s.cachedShape = shape
	s.cachedDefs = defs
	return shape, nil
}

// AsShapeDefs returns the named RAML types extracted from JSON Schema $ref targets
// during the most recent AsShape() call, in encounter order.
// Returns nil when AsShape() has not been called yet.
func (s *JSONShape) AsShapeDefs() *orderedmap.OrderedMap[string, Shape] {
	return s.cachedDefs
}

// jsonSchemaToShape recursively converts a compiled *jsonschema.Schema into the nearest
// RAML Shape. visited is a DFS-stack guard used to detect self-referential schemas.
// defs accumulates named shapes extracted from $ref targets.
// r.jsonShapeRegistry caches completed conversions so repeated $ref references to the
// same compiled schema share a single view Shape across the entire RAML instance.
func jsonSchemaToShape(
	r *RAML,
	parent *BaseShape,
	src *jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (result Shape, retErr error) {
	if src == nil {
		return makeCompiledAnyShape(r, parent), nil
	}

	// Reuse a previously completed conversion of the same compiled schema.
	if shape, found := r.jsonShapeRegistry[src]; found {
		return shape, nil
	}

	// Cycle: return a RecursiveShape pointing back to the in-progress head base.
	if head, seen := visited[src]; seen {
		recBase := newViewBase(r, head.Location)
		recBase.Name = head.Name
		recBase.Type = TypeRecursive
		rs := &RecursiveShape{BaseShape: recBase, Head: head}
		recBase.SetShape(rs)
		return rs, nil
	}

	// Boolean schema: true → any, false → unsupported.
	if src.Bool != nil {
		if !*src.Bool {
			return nil, fmt.Errorf("boolean-false JSON Schema has no RAML equivalent")
		}
		return makeCompiledAnyShape(r, parent), nil
	}

	// Follow $ref. Named targets are registered in defs so the RAML converter
	// can emit them as library type declarations. This block lives before the
	// base allocation because most $ref paths return early without needing a
	// local BaseShape.
	if src.Ref != nil {
		name := nameFromSchemaLocation(src.Ref.Location)
		if name != "" {
			existing, present := defs.Get(name)
			if present && existing != nil {
				// Already fully built — return it directly.
				return existing, nil
			}
			if !present {
				// Mark in-progress sentinel before recursing to handle cycles.
				defs.Set(name, nil)
				fullShape, err := jsonSchemaToShape(r, parent, src.Ref, visited, defs)
				if err != nil {
					return nil, err
				}
				fullShape.Base().Name = name
				defs.Set(name, fullShape)
				r.jsonShapeRegistry[src] = fullShape
				return fullShape, nil
			}
			// present && existing == nil: target is in-progress (cyclic $ref) —
			// fall through and convert the target inline to break the cycle.
		}
		return jsonSchemaToShape(r, parent, src.Ref, visited, defs)
	}

	// if/then/else has no RAML equivalent.
	if src.If != nil {
		return nil, fmt.Errorf("if/then/else is not representable as a RAML type")
	}

	// Register base early so recursive back-references via visited can resolve to it.
	// base.Shape is set below once the concrete type is determined.
	base := makeCompiledBase(r, parent, src)
	visited[src] = base
	defer delete(visited, src)
	// Cache the completed shape so repeated $ref references to the same compiled
	// schema reuse a single view Shape across the entire RAML instance.
	defer func() {
		if retErr == nil && result != nil {
			r.jsonShapeRegistry[src] = result
		}
	}()

	// allOf: merge all sub-schemas via sequential inheritance.
	if len(src.AllOf) > 0 {
		merged, err := mergeAllOf(r, base, src.AllOf, visited, defs)
		if err != nil {
			return nil, err
		}
		applyBaseAnnotations(base, merged.Base())
		return merged, nil
	}

	// oneOf → union (not semantically identical to anyOf, but the closest RAML
	// representation; the distinction between "exactly one" and "at least one" is lost).
	if len(src.OneOf) > 0 {
		return buildCompiledUnionFrom(r, base, src.OneOf, visited, defs)
	}

	// No type constraint.
	if src.Types == nil || src.Types.IsEmpty() {
		if len(src.AnyOf) > 0 {
			return buildCompiledUnionFrom(r, base, src.AnyOf, visited, defs)
		}
		base.Type = TypeAny
		s := &AnyShape{BaseShape: base}
		base.SetShape(s)
		return s, nil
	}

	typeStrs := src.Types.ToStrings()
	if len(typeStrs) > 1 {
		// JSON Schema allows an array of types (e.g. ["string","null"]).
		// Represent as a RAML union; each member carries only its type — no
		// constraints (those live on the union base itself via baseDecl).
		base.Type = TypeUnion
		s := &UnionShape{BaseShape: base}
		base.SetShape(s)
		s.AnyOf = make([]*BaseShape, 0, len(typeStrs))
		for _, t := range typeStrs {
			memberBase := newViewBase(r, base.Location)
			memberShape, err := buildCompiledShapeForType(r, memberBase, t, src, visited, defs)
			if err != nil {
				return nil, err
			}
			s.AnyOf = append(s.AnyOf, memberShape.Base())
		}
		return s, nil
	}

	return buildCompiledShapeForType(r, base, typeStrs[0], src, visited, defs)
}

// buildCompiledShapeForType dispatches on a single JSON Schema type string and
// builds the corresponding RAML shape.  src carries the full schema so that
// constraints (minLength, pattern, properties, …) are applied to the shape.
func buildCompiledShapeForType(
	r *RAML,
	base *BaseShape,
	t string,
	src *jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (Shape, error) {
	switch t {
	case TypeObject:
		return buildCompiledObject(r, base, src, visited, defs)
	case TypeArray:
		return buildCompiledArray(r, base, src, visited, defs)
	case TypeString:
		return buildCompiledString(base, src), nil
	case TypeInteger:
		return buildCompiledInteger(base, src), nil
	case TypeNumber:
		return buildCompiledNumber(base, src), nil
	case TypeBoolean:
		base.Type = TypeBoolean
		s := &BooleanShape{BaseShape: base}
		base.SetShape(s)
		return s, nil
	case TypeNull:
		base.Type = TypeNil
		s := &NilShape{BaseShape: base}
		base.SetShape(s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported JSON Schema type %q", t)
	}
}

// newViewBase allocates a minimal BaseShape for use in view (ephemeral) shapes
// produced by jsonSchemaToShape.  Unlike r.MakeBaseShape it does NOT register
// the shape in r.shapes, and it omits the three always-empty ordered-map
// allocations (CustomDomainProperties / CustomShapeFacets /
// CustomShapeFacetDefinitions), all of whose Len() methods are nil-safe.
func newViewBase(r *RAML, location string) *BaseShape {
	base := &BaseShape{
		ID:       r.generateSequenceID(),
		Location: location,
		KeyPos:   stacktrace.Position{},
		ValuePos: stacktrace.Position{},
		raml:     r,
	}
	// Mark as unwrapped so JSON-LD serialisation skips the Inherits/Alias link
	// blocks and emits the concrete type inline (mirrors behaviour of a fully
	// resolved RAML shape after UnwrapShapes).
	base.SetUnwrapped()
	return base
}

// makeCompiledAnyShape creates a minimal AnyShape with no constraints.
func makeCompiledAnyShape(r *RAML, parent *BaseShape) Shape {
	base := newViewBase(r, parent.Location)
	base.Type = TypeAny
	s := &AnyShape{BaseShape: base}
	base.SetShape(s)
	return s
}

// makeCompiledBase allocates a BaseShape and fills its annotation fields from the
// compiled schema's title, description, default, examples, and enum.
func makeCompiledBase(r *RAML, parent *BaseShape, src *jsonschema.Schema) *BaseShape {
	base := newViewBase(r, parent.Location)
	if src.Title != "" {
		base.DisplayName = &ScalarFacet[string]{Value: src.Title}
	}
	if src.Description != "" {
		base.Description = &ScalarFacet[string]{Value: src.Description}
	}
	if src.Default != nil {
		base.Default = &DataNode{Value: anyToNodeValue(*src.Default)}
	}
	if len(src.Examples) > 0 {
		m := orderedmap.New[string, *Example](len(src.Examples))
		for i, ex := range src.Examples {
			key := strconv.Itoa(i)
			m.Set(key, &Example{
				Name: key,
				Data: &DataNode{Value: anyToNodeValue(ex)},
			})
		}
		base.Examples = &Examples{Map: m}
	}
	if src.Enum != nil && len(src.Enum.Values) > 0 {
		base.Enum = make(Nodes, len(src.Enum.Values))
		for i, v := range src.Enum.Values {
			base.Enum[i] = &DataNode{Value: anyToNodeValue(v)}
		}
	}
	return base
}

// intPtrToScalarUint64 converts a *int JSON Schema length/count constraint into
// *ScalarFacet[uint64] as expected by RAML shape facets.
func intPtrToScalarUint64(v *int) *ScalarFacet[uint64] {
	if v == nil {
		return nil
	}
	return &ScalarFacet[uint64]{Value: uint64(*v)}
}

// ratPtrToScalarRat wraps a *big.Rat numeric JSON Schema constraint into
// *ScalarFacet[*big.Rat] as expected by RAML number/integer shape facets.
func ratPtrToScalarRat(v *big.Rat) *ScalarFacet[*big.Rat] {
	if v == nil {
		return nil
	}
	return &ScalarFacet[*big.Rat]{Value: v}
}

func buildCompiledObject(
	r *RAML,
	base *BaseShape,
	src *jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (Shape, error) {
	if _, isSchema := src.AdditionalProperties.(*jsonschema.Schema); isSchema {
		return nil, fmt.Errorf("schema-form additionalProperties is not representable as a RAML type")
	}
	base.Type = TypeObject
	s := &ObjectShape{BaseShape: base}
	base.SetShape(s)

	s.MinProperties = intPtrToScalarUint64(src.MinProperties)
	s.MaxProperties = intPtrToScalarUint64(src.MaxProperties)

	if ap, ok := src.AdditionalProperties.(bool); ok {
		s.AdditionalProperties = &ScalarFacet[bool]{Value: ap}
	}

	if l := len(src.Properties); l > 0 {
		requiredSet := make(map[string]struct{}, len(src.Required))
		for _, req := range src.Required {
			requiredSet[req] = struct{}{}
		}
		s.Properties = orderedmap.New[string, Property](l)
		for name, propSchema := range src.Properties {
			propShape, err := jsonSchemaToShape(r, base, propSchema, visited, defs)
			if err != nil {
				return nil, fmt.Errorf("property %q: %w", name, err)
			}
			_, required := requiredSet[name]
			s.Properties.Set(name, Property{
				Name:     name,
				Base:     propShape.Base(),
				Required: required,
			})
		}
	}

	if l := len(src.PatternProperties); l > 0 {
		s.PatternProperties = orderedmap.New[string, PatternProperty](l)
		for re, propSchema := range src.PatternProperties {
			propShape, err := jsonSchemaToShape(r, base, propSchema, visited, defs)
			if err != nil {
				return nil, fmt.Errorf("patternProperty %q: %w", re.String(), err)
			}
			key := "/" + re.String() + "/"
			compiled, compileErr := regexp.Compile(re.String())
			if compileErr != nil {
				return nil, fmt.Errorf("compile patternProperty regexp %q: %w", re.String(), compileErr)
			}
			s.PatternProperties.Set(key, PatternProperty{
				Pattern: compiled,
				Base:    propShape.Base(),
			})
		}
	}

	return s, nil
}

func buildCompiledArray(
	r *RAML,
	base *BaseShape,
	src *jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (Shape, error) {
	if _, isTuple := src.Items.([]*jsonschema.Schema); isTuple {
		return nil, fmt.Errorf("tuple-form items is not representable as a RAML type")
	}
	base.Type = TypeArray
	s := &ArrayShape{BaseShape: base}
	base.SetShape(s)

	s.MinItems = intPtrToScalarUint64(src.MinItems)
	s.MaxItems = intPtrToScalarUint64(src.MaxItems)
	if src.UniqueItems {
		s.UniqueItems = &ScalarFacet[bool]{Value: true}
	}

	if itemSchema, ok := src.Items.(*jsonschema.Schema); ok && itemSchema != nil {
		itemShape, err := jsonSchemaToShape(r, base, itemSchema, visited, defs)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}
		s.Items = itemShape.Base()
	}

	return s, nil
}

func buildCompiledString(base *BaseShape, src *jsonschema.Schema) Shape {
	base.Type = TypeString
	s := &StringShape{BaseShape: base}
	base.SetShape(s)
	s.MinLength = intPtrToScalarUint64(src.MinLength)
	s.MaxLength = intPtrToScalarUint64(src.MaxLength)
	if src.Pattern != nil {
		if compiled, err := regexp.Compile(src.Pattern.String()); err == nil {
			s.Pattern = &ScalarFacet[*regexp.Regexp]{Value: compiled}
		}
	}
	return s
}

func buildCompiledInteger(base *BaseShape, src *jsonschema.Schema) Shape {
	base.Type = TypeInteger
	s := &IntegerShape{BaseShape: base}
	base.SetShape(s)
	// JSON Schema integer constraints are *big.Rat; RAML IntegerShape uses *big.Int.
	// Taking the numerator is safe for whole-number bounds (denominator == 1).
	if src.Minimum != nil {
		s.Minimum = &ScalarFacet[*big.Int]{Value: src.Minimum.Num()}
	}
	if src.Maximum != nil {
		s.Maximum = &ScalarFacet[*big.Int]{Value: src.Maximum.Num()}
	}
	s.MultipleOf = ratPtrToScalarRat(src.MultipleOf)
	return s
}

func buildCompiledNumber(base *BaseShape, src *jsonschema.Schema) Shape {
	base.Type = TypeNumber
	s := &NumberShape{BaseShape: base}
	base.SetShape(s)
	s.Minimum = ratPtrToScalarRat(src.Minimum)
	s.Maximum = ratPtrToScalarRat(src.Maximum)
	s.MultipleOf = ratPtrToScalarRat(src.MultipleOf)
	return s
}

// buildCompiledUnionFrom builds a UnionShape from an explicit slice of member schemas.
// Used for both anyOf and oneOf (oneOf is translated approximately as a union).
func buildCompiledUnionFrom(
	r *RAML,
	base *BaseShape,
	members []*jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (Shape, error) {
	base.Type = TypeUnion
	s := &UnionShape{BaseShape: base}
	base.SetShape(s)
	s.AnyOf = make([]*BaseShape, 0, len(members))
	for i, memberSchema := range members {
		memberShape, err := jsonSchemaToShape(r, base, memberSchema, visited, defs)
		if err != nil {
			return nil, fmt.Errorf("member[%d]: %w", i, err)
		}
		s.AnyOf = append(s.AnyOf, memberShape.Base())
	}
	return s, nil
}

// mergeAllOf merges all allOf member schemas into a single shape via sequential
// inheritance. AnyShape members (no type constraint) are absorbed without
// modification. Returns an error when member types are incompatible.
func mergeAllOf(
	r *RAML,
	parent *BaseShape,
	members []*jsonschema.Schema,
	visited map[*jsonschema.Schema]*BaseShape,
	defs *orderedmap.OrderedMap[string, Shape],
) (Shape, error) {
	var result Shape
	for i, member := range members {
		s, err := jsonSchemaToShape(r, parent, member, visited, defs)
		if err != nil {
			return nil, fmt.Errorf("allOf[%d]: %w", i, err)
		}
		if result == nil {
			result = s
			continue
		}
		// AnyShape members add no constraints — skip them.
		if _, ok := s.(*AnyShape); ok {
			continue
		}
		// If the accumulated result is still AnyShape, the first typed member wins.
		if _, ok := result.(*AnyShape); ok {
			result = s
			continue
		}
		// Both typed: merge the new member into the accumulated result.
		merged, err := result.Base().Shape.inherit(s)
		if err != nil {
			return nil, fmt.Errorf("allOf merge at [%d]: %w", i, err)
		}
		result = merged
	}
	if result == nil {
		return makeCompiledAnyShape(r, parent), nil
	}
	return result, nil
}

// applyBaseAnnotations copies annotation fields (title, description, default,
// examples, enum) from src onto dst, but only when dst doesn't already have
// a value for each field. Used to propagate the outer JSON Schema's annotations
// onto the shape produced by an allOf merge.
func applyBaseAnnotations(src, dst *BaseShape) {
	if dst.DisplayName == nil {
		dst.DisplayName = src.DisplayName
	}
	if dst.Description == nil {
		dst.Description = src.Description
	}
	if dst.Default == nil {
		dst.Default = src.Default
	}
	if dst.Examples == nil {
		dst.Examples = src.Examples
	}
	if dst.Enum == nil {
		dst.Enum = src.Enum
	}
}

// nameFromSchemaLocation derives a RAML type name from a compiled JSON Schema
// Location URL. Returns empty string when no usable name can be extracted.
//
//   - #/definitions/Foo or #/$defs/Foo fragment  →  "Foo"
//   - plain file URL like file:///path/foo.json   →  "foo"
//
// Avoids allocations: no strings.Split; uses only IndexByte/HasPrefix.
func nameFromSchemaLocation(loc string) string {
	if loc == "" {
		return ""
	}

	fragIdx := strings.IndexByte(loc, '#')
	if fragIdx < 0 {
		// No fragment — use file stem.
		return schemaFileStem(loc)
	}

	frag := loc[fragIdx+1:] // e.g. "/definitions/Foo" or "/$defs/Foo"

	// Extract the top-level definition name without allocating a slice.
	var rest string
	switch {
	case strings.HasPrefix(frag, "/definitions/"):
		rest = frag[len("/definitions/"):]
	case strings.HasPrefix(frag, "/$defs/"):
		rest = frag[len("/$defs/"):]
	default:
		// Unknown fragment form — fall back to file stem.
		return schemaFileStem(loc[:fragIdx])
	}

	// Take only the first path segment: for /definitions/Foo/properties/bar
	// we want "Foo" (the top-level definition), not "bar".
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// schemaFileStem returns the last path segment of a URL/path without its extension.
// Returns empty string when no usable stem can be extracted.
func schemaFileStem(filePart string) string {
	seg := filePart
	if i := strings.LastIndexAny(seg, "/\\"); i >= 0 {
		seg = seg[i+1:]
	}
	if i := strings.LastIndexByte(seg, '.'); i >= 0 {
		seg = seg[:i]
	}
	return seg
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

func (s *UnknownShape) validate(_ any, _ string) error {
	return StacktraceNew("cannot validate against unknown shape", s.Location, stacktrace.WithPosition(&s.KeyPos))
}

func (s *UnknownShape) unmarshalYAMLNodes(v []*yaml.Node) error {
	s.facets = v
	return nil
}

func (s *UnknownShape) inherit(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot inherit from unknown shape", s.Location, stacktrace.WithPosition(&s.KeyPos))
}

func (s *UnknownShape) alias(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot alias from unknown shape", s.Location, stacktrace.WithPosition(&s.KeyPos))
}

func (s *UnknownShape) check() error {
	return StacktraceNew("cannot check unknown shape", s.Location, stacktrace.WithPosition(&s.KeyPos))
}

func (s *UnknownShape) String() string {
	return fmt.Sprintf("UnknownShape{facets:%d}", len(s.facets))
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
	ss, err := checkAliasType[*RecursiveShape](s, source)
	if err != nil {
		return nil, err
	}
	s.Head = ss.Head
	return s, nil
}

func (s *RecursiveShape) validate(v any, ctxPath string) error {
	if err := s.Head.validateAt(v, ctxPath); err != nil {
		return fmt.Errorf("validate recursive shape: %w", err)
	}
	return nil
}

// Inherit merges the source shape into the target shape.
func (s *RecursiveShape) inherit(_ Shape) (Shape, error) {
	return nil, StacktraceNew("cannot inherit from recursive shape", s.Location, stacktrace.WithPosition(&s.KeyPos))
}

func (s *RecursiveShape) check() error {
	return nil
}

func (s *RecursiveShape) String() string {
	if s.Head != nil {
		return fmt.Sprintf("RecursiveShape{head:%s}", s.Head.Name)
	}
	return "RecursiveShape{head:nil}"
}
