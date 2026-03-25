package raml

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type WrapperFunc[T jsonSchemaWrapper[T]] func(conv *JSONSchemaConverter[T], core *JSONSchemaGeneric[T], src *BaseShape) T

type JSONSchemaConverterOptions[T jsonSchemaWrapper[T]] struct {
	// omitRefs bool
	wrap WrapperFunc[T]
}

type JSONSchemaConverterOpt[T jsonSchemaWrapper[T]] interface {
	apply(*JSONSchemaConverterOptions[T])
}

type optWrapper[T jsonSchemaWrapper[T]] struct{ f WrapperFunc[T] }

//nolint:unused // Actually used in JSONSchemaConverter constructor.
func (o optWrapper[T]) apply(c *JSONSchemaConverterOptions[T]) { c.wrap = o.f }

// WithWrapper lets the caller provide a dialect wrapper.
func WithWrapper[T jsonSchemaWrapper[T]](f WrapperFunc[T]) JSONSchemaConverterOpt[T] {
	return optWrapper[T]{f}
}

// type optOmitRefs[T jsonSchemaWrapper[T]] struct{ omitRefs bool }

// func (o optOmitRefs[T]) apply(c *JSONSchemaConverterOptions[T]) { c.omitRefs = o.omitRefs }
// func WithOmitRefs[T jsonSchemaWrapper[T]](b bool) JSONSchemaConverterOpt[T] {
// 	return optOmitRefs[T]{b}
// }

type JSONSchemaConverter[T jsonSchemaWrapper[T]] struct {
	ShapeVisitor[T]

	definitions map[string]T

	opts JSONSchemaConverterOptions[T]
}

func NewJSONSchemaConverter[T jsonSchemaWrapper[T]](opt ...JSONSchemaConverterOpt[T]) (*JSONSchemaConverter[T], error) {
	c := &JSONSchemaConverter[T]{definitions: make(map[string]T)}
	for _, o := range opt {
		o.apply(&c.opts)
	}
	if c.opts.wrap == nil {
		if _, ok := any((*JSONSchema)(nil)).(T); !ok {
			return nil, errors.New("NewJSONSchemaConverter requires WithWrapper for customized schemas")
		}
	}
	return c, nil
}

func (c *JSONSchemaConverter[T]) Convert(s Shape) (T, error) {
	// TODO: Need to pass *BaseShape
	var zero T
	if !s.Base().IsUnwrapped() {
		return zero, fmt.Errorf("entrypoint shape must be unwrapped")
	}

	entrypointName := s.Base().Name
	c.definitions = make(map[string]T)
	// NOTE: Assign empty schema before traversing to definitions to occupy the name.
	c.definitions[entrypointName] = zero
	c.definitions[entrypointName] = c.Visit(s)

	core := &JSONSchemaGeneric[T]{
		Version:     JSONSchemaVersion,
		Ref:         "#/definitions/" + entrypointName,
		Definitions: c.definitions,
	}

	if c.opts.wrap != nil {
		return c.opts.wrap(c, core, nil), nil
	}

	return any(core).(T), nil
}

func (c *JSONSchemaConverter[T]) Visit(s Shape) T {
	switch shapeType := s.(type) {
	case *ObjectShape:
		return c.VisitObjectShape(shapeType)
	case *ArrayShape:
		return c.VisitArrayShape(shapeType)
	case *StringShape:
		return c.VisitStringShape(shapeType)
	case *NumberShape:
		return c.VisitNumberShape(shapeType)
	case *IntegerShape:
		return c.VisitIntegerShape(shapeType)
	case *BooleanShape:
		return c.VisitBooleanShape(shapeType)
	case *FileShape:
		return c.VisitFileShape(shapeType)
	case *UnionShape:
		return c.VisitUnionShape(shapeType)
	case *NilShape:
		return c.VisitNilShape(shapeType)
	case *AnyShape:
		return c.VisitAnyShape(shapeType)
	case *DateTimeShape:
		return c.VisitDateTimeShape(shapeType)
	case *DateTimeOnlyShape:
		return c.VisitDateTimeOnlyShape(shapeType)
	case *DateOnlyShape:
		return c.VisitDateOnlyShape(shapeType)
	case *TimeOnlyShape:
		return c.VisitTimeOnlyShape(shapeType)
	case *JSONShape:
		return c.VisitJSONShape(shapeType)
	case *RecursiveShape:
		return c.VisitRecursiveShape(shapeType)
	default:
		var zero T
		return zero
	}
}

func (c *JSONSchemaConverter[T]) VisitObjectShape(s *ObjectShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()

	schema.Type = TypeObject
	schema.MinProperties = s.MinProperties
	schema.MaxProperties = s.MaxProperties
	schema.AdditionalProperties = s.AdditionalProperties

	if s.Properties != nil {
		schema.Properties = orderedmap.New[string, T](s.Properties.Len())
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			schema.Properties.Set(k, c.Visit(v.Base.Shape))
			if v.Required {
				schema.Required = append(schema.Required, k)
			}
		}
	}
	if s.PatternProperties != nil {
		schema.PatternProperties = orderedmap.New[string, T](s.PatternProperties.Len())
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			k = k[1 : len(k)-1]
			schema.PatternProperties.Set(k, c.Visit(v.Base.Shape))
		}
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitArrayShape(s *ArrayShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()

	schema.Type = "array"
	schema.MinItems = s.MinItems
	schema.MaxItems = s.MaxItems
	schema.UniqueItems = s.UniqueItems

	if s.Items != nil {
		schema.Items = c.Visit(s.Items.Shape)
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitUnionShape(s *UnionShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()

	schema.AnyOf = make([]T, len(s.AnyOf))
	for i, item := range s.AnyOf {
		schema.AnyOf[i] = c.Visit(item.Shape)
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitStringShape(s *StringShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString
	schema.MinLength = s.MinLength
	schema.MaxLength = s.MaxLength
	if s.Pattern != nil {
		schema.Pattern = s.Pattern.String()
	}
	if s.Enum != nil {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = v.Value
		}
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitIntegerShape(s *IntegerShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeInteger
	if s.Minimum != nil {
		schema.Minimum = json.Number(s.Minimum.String())
	}
	if s.Maximum != nil {
		schema.Maximum = json.Number(s.Maximum.String())
	}
	if s.MultipleOf != nil {
		schema.MultipleOf = json.Number(strconv.FormatFloat(*s.MultipleOf, 'f', -1, 64))
	}
	if s.Enum != nil {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = v.Value
		}
	}
	// TODO: JSON Schema does not have a format for numbers
	return node
}

func (c *JSONSchemaConverter[T]) VisitNumberShape(s *NumberShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeNumber
	if s.Minimum != nil {
		schema.Minimum = json.Number(strconv.FormatFloat(*s.Minimum, 'f', -1, 64))
	}
	if s.Maximum != nil {
		schema.Maximum = json.Number(strconv.FormatFloat(*s.Maximum, 'f', -1, 64))
	}
	if s.MultipleOf != nil {
		schema.MultipleOf = json.Number(strconv.FormatFloat(*s.MultipleOf, 'f', -1, 64))
	}
	if s.Enum != nil {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = v.Value
		}
	}
	// TODO: JSON Schema does not have a format for numbers
	return node
}

func (c *JSONSchemaConverter[T]) VisitFileShape(s *FileShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString
	schema.MinLength = s.MinLength
	schema.MaxLength = s.MaxLength
	schema.ContentEncoding = "base64"

	// TODO: JSON Schema allows for only one content media type
	if s.FileTypes != nil {
		maybeStr, ok := s.FileTypes[0].Value.(string)
		if !ok {
			panic("file type must be a string")
		}
		schema.ContentMediaType = maybeStr
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitBooleanShape(s *BooleanShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeBoolean

	if s.Enum != nil {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = v.Value
		}
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitDateTimeShape(s *DateTimeShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString

	if s.Format != nil {
		switch *s.Format {
		case DateTimeFormatRFC3339:
			schema.Format = FormatDateTime
		case DateTimeFormatRFC2616:
			schema.Pattern = "^(Mon|Tue|Wed|Thu|Fri|Sat|Sun), ([0-3][0-9]) " +
				"(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ([0-9]{4})" +
				" ([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9] GMT$"
		}
	} else {
		schema.Format = FormatDateTime
	}
	return node
}

func (c *JSONSchemaConverter[T]) VisitDateTimeOnlyShape(s *DateTimeOnlyShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString
	schema.Pattern = "^[0-9]{4}-(?:0[0-9]|1[0-2])-(?:[0-2][0-9]|3[01])T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$"
	return node
}

func (c *JSONSchemaConverter[T]) VisitDateOnlyShape(s *DateOnlyShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString
	schema.Format = FormatDate
	return node
}

func (c *JSONSchemaConverter[T]) VisitTimeOnlyShape(s *TimeOnlyShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeString
	schema.Format = FormatTime
	return node
}

func (c *JSONSchemaConverter[T]) VisitAnyShape(s *AnyShape) T {
	return c.makeSchemaFromBaseShape(s.Base())
}

func (c *JSONSchemaConverter[T]) VisitNilShape(s *NilShape) T {
	node := c.makeSchemaFromBaseShape(s.Base())
	schema := node.Generic()
	schema.Type = TypeNull
	return node
}

func (c *JSONSchemaConverter[T]) VisitRecursiveShape(s *RecursiveShape) T {
	// NOTE: Recursive schema will always produce ref.
	// Ref ignores all other keywords defined within the schema per JSON Schema spec.

	// NOTE: We create empty schema because all base RAML types are allowed to have
	// custom facets which can be recursive.
	// The use of `makeSchemaFromBaseShape` will lead to infinite recursion.
	node := c.makeEmptySchema()
	schema := node.Generic()

	head := s.Head.Shape
	baseHead := head.Base()
	// TODO: Type name is not unique, need pretty naming to avoid collisions.
	definition := baseHead.Name
	if _, ok := c.definitions[definition]; !ok {
		// NOTE: Assign empty schema to definitions to occupy the name before traversing.
		var placeholder T
		c.definitions[definition] = placeholder
		c.definitions[definition] = c.Visit(head)
	}
	schema.Ref = "#/definitions/" + definition

	return node
}

func (c *JSONSchemaConverter[T]) VisitJSONShape(s *JSONShape) T {
	// Base RAML meta (annotations, facets) first …
	node := c.makeSchemaFromBaseShape(s.Base())
	dst := node.Generic()

	// Recast the plain schema to T
	recasted := c.recast(s.Schema)
	src := recasted.Generic()

	// NOTE: RAML type may override common properties like title, description, etc.
	// We can safely modify src since recast produces a copy.
	if dst.Title != "" {
		src.Title = dst.Title
	}
	if dst.Description != "" {
		src.Description = dst.Description
	}
	if dst.Default != nil {
		src.Default = dst.Default
	}
	if dst.Examples != nil {
		src.Examples = dst.Examples
	}

	// Copy every other keyword from recast schema
	*dst = *src

	// NOTE: Nested JSON Schema may not have $schema keyword.
	dst.Version = ""
	return node
}

func (c *JSONSchemaConverter[T]) recast(src *JSONSchema) T {
	var zero T
	if src == nil {
		return zero
	}

	core := &JSONSchemaGeneric[T]{
		Version: src.Version,
		ID:      src.ID,
		Ref:     src.Ref,
		Comment: src.Comment,

		If:   c.recast(src.If),
		Then: c.recast(src.Then),
		Else: c.recast(src.Else),
		Not:  c.recast(src.Not),

		Items: c.recast(src.Items),

		AdditionalProperties: src.AdditionalProperties,
		PropertyNames:        c.recast(src.PropertyNames),
		Type:                 src.Type,
		Enum:                 src.Enum,
		Const:                src.Const,
		MultipleOf:           src.MultipleOf,
		Maximum:              src.Maximum,
		Minimum:              src.Minimum,
		MaxLength:            src.MaxLength,
		MinLength:            src.MinLength,
		Pattern:              src.Pattern,
		MaxItems:             src.MaxItems,
		MinItems:             src.MinItems,
		UniqueItems:          src.UniqueItems,
		MaxContains:          src.MaxContains,
		MinContains:          src.MinContains,
		MaxProperties:        src.MaxProperties,
		MinProperties:        src.MinProperties,
		Required:             src.Required,
		ContentEncoding:      src.ContentEncoding,
		ContentMediaType:     src.ContentMediaType,
		Format:               src.Format,

		Title:       src.Title,
		Description: src.Description,
		Default:     src.Default,
		Examples:    src.Examples,
	}

	if len(src.AnyOf) > 0 {
		core.AnyOf = make([]T, len(src.AnyOf))
		for i, it := range src.AnyOf {
			core.AnyOf[i] = c.recast(it)
		}
	}
	if len(src.AllOf) > 0 {
		core.AllOf = make([]T, len(src.AllOf))
		for i, it := range src.AllOf {
			core.AllOf[i] = c.recast(it)
		}
	}
	if len(src.OneOf) > 0 {
		core.OneOf = make([]T, len(src.OneOf))
		for i, it := range src.OneOf {
			core.OneOf[i] = c.recast(it)
		}
	}
	if src.Properties.Len() > 0 {
		core.Properties = orderedmap.New[string, T](src.Properties.Len())
		for p := src.Properties.Oldest(); p != nil; p = p.Next() {
			core.Properties.Set(p.Key, c.recast(p.Value))
		}
	}
	if src.PatternProperties.Len() > 0 {
		core.PatternProperties = orderedmap.New[string, T](src.PatternProperties.Len())
		for p := src.PatternProperties.Oldest(); p != nil; p = p.Next() {
			core.PatternProperties.Set(p.Key, c.recast(p.Value))
		}
	}
	if len(src.Definitions) > 0 {
		core.Definitions = make(map[string]T, len(src.Definitions))
		for k, v := range src.Definitions {
			core.Definitions[k] = c.recast(v)
		}
	}
	if c.opts.wrap != nil {
		return c.opts.wrap(c, core, nil)
	}
	return any(core).(T)
}

func (c *JSONSchemaConverter[T]) makeEmptySchema() T {
	core := &JSONSchemaGeneric[T]{}
	if c.opts.wrap != nil {
		return c.opts.wrap(c, core, nil)
	}
	return any(core).(T)
}

func (c *JSONSchemaConverter[T]) makeSchemaFromBaseShape(base *BaseShape) T {
	core := &JSONSchemaGeneric[T]{}
	if base.DisplayName != nil {
		core.Title = *base.DisplayName
	}
	if base.Description != nil {
		core.Description = *base.Description
	}
	if base.Default != nil {
		core.Default = base.Default.Value
	}
	if base.Examples != nil {
		for pair := base.Examples.Map.Oldest(); pair != nil; pair = pair.Next() {
			ex := pair.Value
			core.Examples = append(core.Examples, ex.Data.Value)
		}
	}
	if base.Example != nil {
		core.Examples = []any{base.Example.Data.Value}
	}
	if c.opts.wrap != nil {
		return c.opts.wrap(c, core, base)
	}
	return any(core).(T)
}
