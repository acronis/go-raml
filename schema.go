package raml

import (
	"encoding/json"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Adapted from https://github.com/invopop/jsonschema/blob/main/schema.go

// Version is the JSON Schema version.
const JSONSchemaVersion = "http://json-schema.org/draft-07/schema"

// Schema represents a JSON Schema object type.
//
// https://json-schema.org/draft-07/draft-handrews-json-schema-00.pdf
type JSONSchemaGeneric[T any] struct {
	Version     string         `json:"$schema,omitempty"`
	ID          string         `json:"$id,omitempty"`
	Ref         string         `json:"$ref,omitempty"`
	Definitions Definitions[T] `json:"definitions,omitempty"`
	Comment     string         `json:"$comment,omitempty"`

	AllOf []*T `json:"allOf,omitempty"`
	AnyOf []*T `json:"anyOf,omitempty"`
	OneOf []*T `json:"oneOf,omitempty"`
	Not   *T   `json:"not,omitempty"`

	If   *T `json:"if,omitempty"`
	Then *T `json:"then,omitempty"`
	Else *T `json:"else,omitempty"`

	Items *T `json:"items,omitempty"`

	Properties           *orderedmap.OrderedMap[string, *T] `json:"properties,omitempty"`
	PatternProperties    *orderedmap.OrderedMap[string, *T] `json:"patternProperties,omitempty"`
	AdditionalProperties *bool                              `json:"additionalProperties,omitempty"`
	PropertyNames        *T                                 `json:"propertyNames,omitempty"`

	Type             string      `json:"type,omitempty"`
	Enum             []any       `json:"enum,omitempty"`
	Const            any         `json:"const,omitempty"`
	MultipleOf       json.Number `json:"multipleOf,omitempty"`
	Maximum          json.Number `json:"maximum,omitempty"`
	Minimum          json.Number `json:"minimum,omitempty"`
	MaxLength        *uint64     `json:"maxLength,omitempty"`
	MinLength        *uint64     `json:"minLength,omitempty"`
	Pattern          string      `json:"pattern,omitempty"`
	MaxItems         *uint64     `json:"maxItems,omitempty"`
	MinItems         *uint64     `json:"minItems,omitempty"`
	UniqueItems      *bool       `json:"uniqueItems,omitempty"`
	MaxContains      *uint64     `json:"maxContains,omitempty"`
	MinContains      *uint64     `json:"minContains,omitempty"`
	MaxProperties    *uint64     `json:"maxProperties,omitempty"`
	MinProperties    *uint64     `json:"minProperties,omitempty"`
	Required         []string    `json:"required,omitempty"`
	ContentEncoding  string      `json:"contentEncoding,omitempty"`
	ContentMediaType string      `json:"contentMediaType,omitempty"`

	Format string `json:"format,omitempty"`

	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`
	Examples    []any  `json:"examples,omitempty"`
}

// Definitions hold schema definitions.
// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
// RFC draft-wright-json-schema-validation-00, section 5.26
type Definitions[T any] map[string]*T

type JSONSchema struct {
	JSONSchemaGeneric[JSONSchema]
	Annotations      *orderedmap.OrderedMap[string, any]         `json:"x-annotations,omitempty"`
	FacetDefinitions *orderedmap.OrderedMap[string, *JSONSchema] `json:"x-facet-definitions,omitempty"`
	FacetData        *orderedmap.OrderedMap[string, any]         `json:"x-facet-data,omitempty"`
}
