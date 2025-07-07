package raml

import (
	"encoding/json"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Adapted from https://github.com/invopop/jsonschema/blob/main/schema.go

// Version is the JSON Schema version.
const JSONSchemaVersion = "http://json-schema.org/draft-07/schema"

type Copyable[T any] interface {
	// DeepCopy creates a deep copy of the JSON Schema object.
	DeepCopy() T

	// Generic returns the generic JSON Schema object.
	ShallowCopy() T

	// Map returns the JSON Schema object as a map[string]any.
	// Implementers should return a map that is compatible for use in gojsonschema.NewRawLoader.
	Map() map[string]any
}

// Schema represents a JSON Schema object type.
//
// https://json-schema.org/draft-07/draft-handrews-json-schema-00.pdf
type JSONSchemaGeneric[T Copyable[T]] struct {
	// TODO: Need to collect unknown "x-" annotations into a dict
	Version string `json:"$schema,omitempty" yaml:"$schema,omitempty"`
	ID      string `json:"$id,omitempty" yaml:"$id,omitempty"`
	Ref     string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Comment string `json:"$comment,omitempty" yaml:"$comment,omitempty"`

	// Definitions hold schema definitions.
	// http://json-schema.org/latest/json-schema-validation.html#rfc.section.5.26
	// RFC draft-wright-json-schema-validation-00, section 5.26
	Definitions map[string]T `json:"definitions,omitempty" yaml:"definitions,omitempty"`

	AllOf []T `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	AnyOf []T `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	OneOf []T `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	Not   T   `json:"not,omitempty" yaml:"not,omitempty"`

	If   T `json:"if,omitempty" yaml:"if,omitempty"`
	Then T `json:"then,omitempty" yaml:"then,omitempty"`
	Else T `json:"else,omitempty" yaml:"else,omitempty"`

	Items T `json:"items,omitempty" yaml:"items,omitempty"`

	Properties           *orderedmap.OrderedMap[string, T] `json:"properties,omitempty" yaml:"properties,omitempty"`
	PatternProperties    *orderedmap.OrderedMap[string, T] `json:"patternProperties,omitempty" yaml:"patternProperties,omitempty"`
	AdditionalProperties *bool                             `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	PropertyNames        T                                 `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`

	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Enum             []any       `json:"enum,omitempty" yaml:"enum,omitempty"`
	Const            any         `json:"const,omitempty" yaml:"const,omitempty"`
	MultipleOf       json.Number `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Maximum          json.Number `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Minimum          json.Number `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	MaxLength        *uint64     `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength        *uint64     `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern          string      `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MaxItems         *uint64     `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems         *uint64     `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems      *bool       `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxContains      *uint64     `json:"maxContains,omitempty" yaml:"maxContains,omitempty"`
	MinContains      *uint64     `json:"minContains,omitempty" yaml:"minContains,omitempty"`
	MaxProperties    *uint64     `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties    *uint64     `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Required         []string    `json:"required,omitempty" yaml:"required,omitempty"`
	ContentEncoding  string      `json:"contentEncoding,omitempty" yaml:"contentEncoding,omitempty"`
	ContentMediaType string      `json:"contentMediaType,omitempty" yaml:"contentMediaType,omitempty"`

	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	Title       string `json:"title,omitempty" yaml:"title,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Default     any    `json:"default,omitempty" yaml:"default,omitempty"`
	Examples    []any  `json:"examples,omitempty" yaml:"examples,omitempty"`
}

func (js *JSONSchemaGeneric[T]) ShallowCopy() *JSONSchemaGeneric[T] {
	if js == nil {
		return nil
	}
	newJs := &JSONSchemaGeneric[T]{}
	*newJs = *js
	return newJs
}

func (js *JSONSchemaGeneric[T]) DeepCopy() *JSONSchemaGeneric[T] {
	if js == nil {
		return nil
	}
	newJs := js.ShallowCopy()

	if len(js.Required) > 0 {
		newJs.Required = make([]string, len(js.Required))
		copy(newJs.Required, js.Required)
	}

	if len(js.Examples) > 0 {
		newJs.Examples = make([]any, len(js.Examples))
		copy(newJs.Examples, js.Examples)
	}

	if len(js.Enum) > 0 {
		newJs.Enum = make([]any, len(js.Enum))
		copy(newJs.Enum, js.Enum)
	}

	if len(js.AnyOf) > 0 {
		newJs.AnyOf = make([]T, len(js.AnyOf))
		for i, v := range js.AnyOf {
			newJs.AnyOf[i] = v.DeepCopy()
		}
	}

	if len(js.AllOf) > 0 {
		newJs.AllOf = make([]T, len(js.AllOf))
		for i, v := range js.AllOf {
			newJs.AllOf[i] = v.DeepCopy()
		}
	}

	if (len(js.OneOf)) > 0 {
		newJs.OneOf = make([]T, len(js.OneOf))
		for i, v := range js.OneOf {
			newJs.OneOf[i] = v.DeepCopy()
		}
	}

	newJs.Not = js.Not.DeepCopy()
	newJs.If = js.If.DeepCopy()
	newJs.Then = js.Then.DeepCopy()
	newJs.Else = js.Else.DeepCopy()
	newJs.Items = js.Items.DeepCopy()
	newJs.PropertyNames = js.PropertyNames.DeepCopy()

	if js.Properties.Len() > 0 {
		newJs.Properties = orderedmap.New[string, T](js.Properties.Len())
		for p := js.Properties.Oldest(); p != nil; p = p.Next() {
			newJs.Properties.Set(p.Key, p.Value.DeepCopy())
		}
	}

	if js.PatternProperties.Len() > 0 {
		newJs.PatternProperties = orderedmap.New[string, T](js.PatternProperties.Len())
		for p := js.PatternProperties.Oldest(); p != nil; p = p.Next() {
			newJs.PatternProperties.Set(p.Key, p.Value.DeepCopy())
		}
	}

	if len(js.Definitions) > 0 {
		newJs.Definitions = make(map[string]T, len(js.Definitions))
		for k, v := range js.Definitions {
			newJs.Definitions[k] = v.DeepCopy()
		}
	}

	return newJs
}

func (js *JSONSchemaGeneric[T]) Map() map[string]any {
	if js == nil {
		return nil
	}
	out := make(map[string]any)

	if js.Version != "" {
		out["$schema"] = js.Version
	}
	if js.ID != "" {
		out["$id"] = js.ID
	}
	if js.Ref != "" {
		out["$ref"] = js.Ref
	}
	if js.Comment != "" {
		out["$comment"] = js.Comment
	}

	if js.Type != "" {
		out["type"] = js.Type
	}
	if len(js.Enum) > 0 {
		out["enum"] = js.Enum
	}
	if js.Const != nil {
		out["const"] = js.Const
	}

	if js.MultipleOf != "" {
		out["multipleOf"] = js.MultipleOf
	}
	if js.Maximum != "" {
		out["maximum"] = js.Maximum
	}
	if js.Minimum != "" {
		out["minimum"] = js.Minimum
	}

	if v := uint64PtrToNumber(js.MaxLength); v != "" {
		out["maxLength"] = v
	}
	if v := uint64PtrToNumber(js.MinLength); v != "" {
		out["minLength"] = v
	}
	if js.Pattern != "" {
		out["pattern"] = js.Pattern
	}
	if v := uint64PtrToNumber(js.MaxItems); v != "" {
		out["maxItems"] = v
	}
	if v := uint64PtrToNumber(js.MinItems); v != "" {
		out["minItems"] = v
	}
	if b := boolPtr(js.UniqueItems); b != nil {
		out["uniqueItems"] = b
	}
	if v := uint64PtrToNumber(js.MaxContains); v != "" {
		out["maxContains"] = v
	}
	if v := uint64PtrToNumber(js.MinContains); v != "" {
		out["minContains"] = v
	}
	if v := uint64PtrToNumber(js.MaxProperties); v != "" {
		out["maxProperties"] = v
	}
	if v := uint64PtrToNumber(js.MinProperties); v != "" {
		out["minProperties"] = v
	}
	if len(js.Required) > 0 {
		arr := make([]any, len(js.Required))
		for i, v := range js.Required {
			arr[i] = v
		}
		out["required"] = arr
	}

	if js.ContentEncoding != "" {
		out["contentEncoding"] = js.ContentEncoding
	}
	if js.ContentMediaType != "" {
		out["contentMediaType"] = js.ContentMediaType
	}

	if js.Format != "" {
		out["format"] = js.Format
	}

	if js.Title != "" {
		out["title"] = js.Title
	}
	if js.Description != "" {
		out["description"] = js.Description
	}
	if js.Default != nil {
		out["default"] = js.Default
	}
	if len(js.Examples) > 0 {
		out["examples"] = js.Examples
	}

	if len(js.Definitions) > 0 {
		defs := make(map[string]any, len(js.Definitions))
		for k, v := range js.Definitions {
			defs[k] = v.Map()
		}
		out["definitions"] = defs
	}

	if len(js.AllOf) > 0 {
		arr := make([]any, len(js.AllOf))
		for i, v := range js.AllOf {
			arr[i] = v.Map()
		}
		out["allOf"] = arr
	}
	if len(js.AnyOf) > 0 {
		arr := make([]any, len(js.AnyOf))
		for i, v := range js.AnyOf {
			arr[i] = v.Map()
		}
		out["anyOf"] = arr
	}
	if len(js.OneOf) > 0 {
		arr := make([]any, len(js.OneOf))
		for i, v := range js.OneOf {
			arr[i] = v.Map()
		}
		out["oneOf"] = arr
	}

	v := js.Not.Map()
	if v != nil {
		out["not"] = v
	}
	v = js.If.Map()
	if v != nil {
		out["if"] = v
	}
	v = js.Then.Map()
	if v != nil {
		out["then"] = v
	}
	v = js.Else.Map()
	if v != nil {
		out["else"] = v
	}
	v = js.Items.Map()
	if v != nil {
		out["items"] = v
	}

	if js.Properties.Len() > 0 {
		o := make(map[string]any, js.Properties.Len())
		for el := js.Properties.Oldest(); el != nil; el = el.Next() {
			o[el.Key] = el.Value.Map()
		}
		out["properties"] = o
	}
	if js.PatternProperties.Len() > 0 {
		o := make(map[string]any, js.PatternProperties.Len())
		for el := js.PatternProperties.Oldest(); el != nil; el = el.Next() {
			o[el.Key] = el.Value.Map()
		}
		out["patternProperties"] = o
	}
	if b := boolPtr(js.AdditionalProperties); b != nil {
		out["additionalProperties"] = b
	}
	v = js.PropertyNames.Map()
	if v != nil {
		out["propertyNames"] = v
	}

	return out
}

// jsonSchemaWrapper – every dialect node exposes the embedded canonical struct.
type jsonSchemaWrapper[T Copyable[T]] interface {
	Generic() *JSONSchemaGeneric[T]
	DeepCopy() T
	ShallowCopy() T
	Map() map[string]any
}

// Plain, spec‑only flavour – just an alias.
// It does not use pointer to embedded structure since it does not use wrapper function that would require a newJs.
type JSONSchema struct {
	JSONSchemaGeneric[*JSONSchema] `yaml:",inline"`
}

func (js *JSONSchema) Generic() *JSONSchemaGeneric[*JSONSchema] { return &js.JSONSchemaGeneric }

func (js *JSONSchema) ShallowCopy() *JSONSchema {
	if js == nil {
		return nil
	}
	return &JSONSchema{JSONSchemaGeneric: *js.JSONSchemaGeneric.ShallowCopy()}
}

func (js *JSONSchema) DeepCopy() *JSONSchema {
	if js == nil {
		return nil
	}
	return &JSONSchema{JSONSchemaGeneric: *js.JSONSchemaGeneric.DeepCopy()}
}

func (js *JSONSchema) Map() map[string]any {
	if js == nil {
		return nil
	}
	return js.JSONSchemaGeneric.Map()
}

// RAML‑extended node (x‑annotations, x‑facet‑*)

type JSONSchemaRAML struct {
	JSONSchemaGeneric[*JSONSchemaRAML] `yaml:",inline"`
	Annotations                        *orderedmap.OrderedMap[string, any]             `json:"x-annotations,omitempty" yaml:"x-annotations,omitempty"`
	FacetDefinitions                   *orderedmap.OrderedMap[string, *JSONSchemaRAML] `json:"x-facet-definitions,omitempty" yaml:"x-facet-definitions,omitempty"`
	FacetData                          *orderedmap.OrderedMap[string, any]             `json:"x-facet-data,omitempty" yaml:"x-facet-data,omitempty"`
}

func (js *JSONSchemaRAML) Generic() *JSONSchemaGeneric[*JSONSchemaRAML] { return &js.JSONSchemaGeneric }

func (js *JSONSchemaRAML) ShallowCopy() *JSONSchemaRAML {
	if js == nil {
		return nil
	}
	newJs := &JSONSchemaRAML{JSONSchemaGeneric: *js.JSONSchemaGeneric.ShallowCopy()}
	if js.Annotations != nil {
		newJs.Annotations = orderedmap.New[string, any](js.Annotations.Len())
		for p := js.Annotations.Oldest(); p != nil; p = p.Next() {
			newJs.Annotations.Set(p.Key, p.Value)
		}
	}
	if js.FacetDefinitions != nil {
		newJs.FacetDefinitions = orderedmap.New[string, *JSONSchemaRAML](js.FacetDefinitions.Len())
		for p := js.FacetDefinitions.Oldest(); p != nil; p = p.Next() {
			newJs.FacetDefinitions.Set(p.Key, p.Value)
		}
	}
	if js.FacetData != nil {
		newJs.FacetData = orderedmap.New[string, any](js.FacetData.Len())
		for p := js.FacetData.Oldest(); p != nil; p = p.Next() {
			newJs.FacetData.Set(p.Key, p.Value)
		}
	}
	return newJs
}

func (js *JSONSchemaRAML) DeepCopy() *JSONSchemaRAML {
	if js == nil {
		return nil
	}
	newJs := &JSONSchemaRAML{JSONSchemaGeneric: *js.JSONSchemaGeneric.DeepCopy()}
	if js.Annotations != nil {
		newJs.Annotations = orderedmap.New[string, any](js.Annotations.Len())
		for p := js.Annotations.Oldest(); p != nil; p = p.Next() {
			newJs.Annotations.Set(p.Key, p.Value)
		}
	}
	if js.FacetDefinitions != nil {
		newJs.FacetDefinitions = orderedmap.New[string, *JSONSchemaRAML](js.FacetDefinitions.Len())
		for p := js.FacetDefinitions.Oldest(); p != nil; p = p.Next() {
			newJs.FacetDefinitions.Set(p.Key, p.Value.DeepCopy())
		}
	}
	if js.FacetData != nil {
		newJs.FacetData = orderedmap.New[string, any](js.FacetData.Len())
		for p := js.FacetData.Oldest(); p != nil; p = p.Next() {
			newJs.FacetData.Set(p.Key, p.Value)
		}
	}
	return newJs
}

func (js *JSONSchemaRAML) Map() map[string]any {
	if js == nil {
		return nil
	}
	m := js.JSONSchemaGeneric.Map()
	if js.Annotations != nil && js.Annotations.Len() > 0 {
		annotations := make(map[string]any, js.Annotations.Len())
		for p := js.Annotations.Oldest(); p != nil; p = p.Next() {
			annotations[p.Key] = p.Value
		}
		m["x-annotations"] = annotations
	}
	if js.FacetDefinitions != nil && js.FacetDefinitions.Len() > 0 {
		facetDefs := make(map[string]any, js.FacetDefinitions.Len())
		for p := js.FacetDefinitions.Oldest(); p != nil; p = p.Next() {
			facetDefs[p.Key] = p.Value.Map()
		}
		m["x-facet-definitions"] = facetDefs
	}
	if js.FacetData != nil && js.FacetData.Len() > 0 {
		facetData := make(map[string]any, js.FacetData.Len())
		for p := js.FacetData.Oldest(); p != nil; p = p.Next() {
			facetData[p.Key] = p.Value
		}
		m["x-facet-data"] = facetData
	}
	return m
}

func JSONSchemaWrapper(c *JSONSchemaConverter[*JSONSchemaRAML], core *JSONSchemaGeneric[*JSONSchemaRAML], b *BaseShape) *JSONSchemaRAML {
	if core == nil {
		return nil
	}
	w := &JSONSchemaRAML{JSONSchemaGeneric: *core}
	if b == nil {
		return w
	}
	if n := b.CustomDomainProperties.Len(); n > 0 {
		m := orderedmap.New[string, any](n)
		for p := b.CustomDomainProperties.Oldest(); p != nil; p = p.Next() {
			m.Set(p.Key, p.Value.Extension.Value)
		}
		w.Annotations = m
	}
	if n := b.CustomShapeFacetDefinitions.Len(); n > 0 {
		m := orderedmap.New[string, *JSONSchemaRAML](n)
		for p := b.CustomShapeFacetDefinitions.Oldest(); p != nil; p = p.Next() {
			m.Set(p.Key, c.Visit(p.Value.Base.Shape))
		}
		w.FacetDefinitions = m
	}
	if n := b.CustomShapeFacets.Len(); n > 0 {
		m := orderedmap.New[string, any](n)
		for p := b.CustomShapeFacets.Oldest(); p != nil; p = p.Next() {
			m.Set(p.Key, p.Value.Value)
		}
		w.FacetData = m
	}
	return w
}

func uint64PtrToNumber(p *uint64) json.Number {
	if p == nil {
		return ""
	}
	return json.Number(strconv.FormatUint(*p, 10))
}

func boolPtr(p *bool) any {
	if p == nil {
		return nil
	}
	return *p
}
