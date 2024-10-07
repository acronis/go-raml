package raml

import (
	"encoding/json"
	"fmt"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type JSONSchemaConverterOpt interface {
	Apply(*JSONSchemaConverterOptions)
}

type optOmitRefs struct {
	omitRefs bool
}

func (o optOmitRefs) Apply(e *JSONSchemaConverterOptions) {
	e.omitRefs = o.omitRefs
}

func WithOmitRefs(omitRefs bool) JSONSchemaConverterOpt {
	return optOmitRefs{omitRefs: omitRefs}
}

type JSONSchemaConverterOptions struct {
	omitRefs bool
}

type JSONSchemaConverter struct {
	ShapeVisitor[JSONSchema]

	definitions    Definitions
	complexSchemas map[int64]*JSONSchema

	opts JSONSchemaConverterOptions
}

func NewJSONSchemaConverter(opts ...JSONSchemaConverterOpt) *JSONSchemaConverter {
	c := &JSONSchemaConverter{}
	for _, opt := range opts {
		opt.Apply(&c.opts)
	}
	return c
}

func (c *JSONSchemaConverter) Convert(s Shape) (*JSONSchema, error) {
	// TODO: Need to pass *BaseShape
	// TODO: JSONSchema converter should also work with non-unwrapped shapes.
	if !s.Base().IsUnwrapped() {
		return nil, fmt.Errorf("entrypoint shape must be unwrapped")
	}

	entrypointName := s.Base().Name
	c.complexSchemas = make(map[int64]*JSONSchema)
	c.definitions = make(Definitions)
	schema := &JSONSchema{}
	// NOTE: Assign empty schema before traversing to definitions to occupy the name.
	// TODO: Probably can be refactored in a better way.
	c.definitions[entrypointName] = schema
	*schema = *c.Visit(s)

	return &JSONSchema{
		Version:     JSONSchemaVersion,
		Ref:         "#/definitions/" + entrypointName,
		Definitions: c.definitions,
	}, nil
}

func (c *JSONSchemaConverter) Visit(s Shape) *JSONSchema {
	switch s := s.(type) {
	case *ObjectShape:
		return c.VisitObjectShape(s)
	case *ArrayShape:
		return c.VisitArrayShape(s)
	case *StringShape:
		return c.VisitStringShape(s)
	case *NumberShape:
		return c.VisitNumberShape(s)
	case *IntegerShape:
		return c.VisitIntegerShape(s)
	case *BooleanShape:
		return c.VisitBooleanShape(s)
	case *FileShape:
		return c.VisitFileShape(s)
	case *UnionShape:
		return c.VisitUnionShape(s)
	case *NilShape:
		return c.VisitNilShape(s)
	case *AnyShape:
		return c.VisitAnyShape(s)
	case *DateTimeShape:
		return c.VisitDateTimeShape(s)
	case *DateTimeOnlyShape:
		return c.VisitDateTimeOnlyShape(s)
	case *DateOnlyShape:
		return c.VisitDateOnlyShape(s)
	case *TimeOnlyShape:
		return c.VisitTimeOnlyShape(s)
	case *JSONShape:
		return c.VisitJSONShape(s)
	case *RecursiveShape:
		return c.VisitRecursiveShape(s)
	default:
		return nil
	}
}

func (c *JSONSchemaConverter) VisitObjectShape(s *ObjectShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	c.complexSchemas[s.Base().ID] = schema

	schema.Type = "object"
	schema.MinProperties = s.MinProperties
	schema.MaxProperties = s.MaxProperties
	schema.AdditionalProperties = s.AdditionalProperties

	if s.Properties != nil {
		schema.Properties = orderedmap.New[string, *JSONSchema](s.Properties.Len())
		for pair := s.Properties.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			schema.Properties.Set(k, c.Visit(v.Shape.Shape))
			if v.Required {
				schema.Required = append(schema.Required, k)
			}
		}
	}
	if s.PatternProperties != nil {
		schema.PatternProperties = orderedmap.New[string, *JSONSchema](s.PatternProperties.Len())
		for pair := s.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			k = k[1 : len(k)-1]
			schema.PatternProperties.Set(k, c.Visit(v.Shape.Shape))
		}
	}
	return schema
}

func (c *JSONSchemaConverter) VisitArrayShape(s *ArrayShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	c.complexSchemas[s.Base().ID] = schema

	schema.Type = "array"
	schema.MinItems = s.MinItems
	schema.MaxItems = s.MaxItems
	schema.UniqueItems = s.UniqueItems

	if s.Items != nil {
		schema.Items = c.Visit(s.Items.Shape)
	}
	return schema
}

func (c *JSONSchemaConverter) VisitUnionShape(s *UnionShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	c.complexSchemas[s.Base().ID] = schema

	schema.AnyOf = make([]*JSONSchema, len(s.AnyOf))
	for i, item := range s.AnyOf {
		schema.AnyOf[i] = c.Visit(item.Shape)
	}
	return schema
}

func (c *JSONSchemaConverter) VisitStringShape(s *StringShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
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
	return schema
}

func (c *JSONSchemaConverter) VisitIntegerShape(s *IntegerShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
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
	return schema
}

func (c *JSONSchemaConverter) VisitNumberShape(s *NumberShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
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
	return schema
}

func (c *JSONSchemaConverter) VisitFileShape(s *FileShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
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
	return schema
}

func (c *JSONSchemaConverter) VisitBooleanShape(s *BooleanShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeBoolean

	if s.Enum != nil {
		schema.Enum = make([]interface{}, len(s.Enum))
		for i, v := range s.Enum {
			schema.Enum[i] = v.Value
		}
	}
	return schema
}

func (c *JSONSchemaConverter) VisitDateTimeShape(s *DateTimeShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeString

	if s.Format != nil {
		switch *s.Format {
		case DateTimeFormatRFC3339:
			schema.Format = "date-time"
		case DateTimeFormatRFC2616:
			schema.Pattern = "^(Mon|Tue|Wed|Thu|Fri|Sat|Sun), ([0-3][0-9]) " +
				"(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) ([0-9]{4})" +
				" ([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9] GMT$"
		}
	} else {
		schema.Format = "date-time"
	}
	return schema
}

func (c *JSONSchemaConverter) VisitDateTimeOnlyShape(s *DateTimeOnlyShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeString
	schema.Pattern = "^[0-9]{4}-(?:0[0-9]|1[0-2])-(?:[0-2][0-9]|3[01])T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]$"
	return schema
}

func (c *JSONSchemaConverter) VisitDateOnlyShape(s *DateOnlyShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeString
	schema.Format = "date"
	return schema
}

func (c *JSONSchemaConverter) VisitTimeOnlyShape(s *TimeOnlyShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeString
	schema.Format = "time"
	return schema
}

func (c *JSONSchemaConverter) VisitAnyShape(s *AnyShape) *JSONSchema {
	return c.makeSchemaFromBaseShape(s.Base())
}

func (c *JSONSchemaConverter) VisitNilShape(s *NilShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	schema.Type = TypeNull
	return schema
}

func (c *JSONSchemaConverter) VisitRecursiveShape(s *RecursiveShape) *JSONSchema {
	// NOTE: Recursive schema will always produce ref.
	// However, ref ignores all other keywords defined within the schema per JSON Schema spec.
	// We keep the keywords just in case the schema is not used as a ref.
	schema := c.makeSchemaFromBaseShape(s.Base())

	head := s.Head.Shape
	baseHead := head.Base()
	// TODO: Type name is not unique, need pretty naming to avoid collisions.
	definition := baseHead.Name
	if c.definitions[definition] == nil {
		defSchema := &JSONSchema{}
		// NOTE: Assign empty defSchema before traversing to definitions to occupy the name.
		c.definitions[definition] = defSchema
		*defSchema = *c.Visit(head)
	}
	schema.Ref = "#/definitions/" + definition

	return schema
}

func (c *JSONSchemaConverter) VisitJSONShape(s *JSONShape) *JSONSchema {
	schema := c.makeSchemaFromBaseShape(s.Base())
	// NOTE: RAML type may override common properties like title, description, etc.
	schema = c.overrideCommonProperties(schema, s.Schema)
	// NOTE: Nested JSON Schema may not have $schema keyword.
	schema.Version = ""
	return schema
}

func (c *JSONSchemaConverter) overrideCommonProperties(parent *JSONSchema, child *JSONSchema) *JSONSchema {
	cs := *child
	if parent.Title != "" {
		cs.Title = parent.Title
	}
	if parent.Description == "" {
		cs.Description = parent.Description
	}
	if parent.Default != nil {
		cs.Default = parent.Default
	}
	if parent.Examples != nil {
		cs.Examples = parent.Examples
	}
	if parent.Extras != nil {
		if cs.Extras == nil {
			cs.Extras = parent.Extras
		} else {
			for k, v := range parent.Extras {
				cs.Extras[k] = v
			}
		}
	}
	return &cs
}

func (c *JSONSchemaConverter) makeSchemaFromBaseShape(base *BaseShape) *JSONSchema {
	schema := &JSONSchema{
		Extras: make(map[string]interface{}),
	}
	if base.DisplayName != nil {
		schema.Title = *base.DisplayName
	}
	if base.Description != nil {
		schema.Description = *base.Description
	}
	if base.Default != nil {
		schema.Default = base.Default.Value
	}
	if base.Examples != nil {
		for pair := base.Examples.Map.Oldest(); pair != nil; pair = pair.Next() {
			ex := pair.Value
			schema.Examples = append(schema.Examples, ex.Data.Value)
		}
	}
	if base.Example != nil {
		schema.Examples = []any{base.Example.Data.Value}
	}
	for pair := base.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		schema.Extras["x-domainExt-"+k] = v.Extension.Value
	}
	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		m := schema.Extras["x-shapeExt-definitions"]
		if m == nil {
			m = make(map[string]interface{})
			schema.Extras["x-shapeExt-definitions"] = m
		}
		shouldBeMap, ok := m.(map[string]interface{})
		if !ok {
			panic("invalid shape extension definitions")
		}
		shapeExtDefs := shouldBeMap
		shapeExtDefs[k] = c.Visit(v.Shape.Shape)
	}
	for pair := base.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		schema.Extras["x-shapeExt-data-"+k] = v.Value
	}
	return schema
}
