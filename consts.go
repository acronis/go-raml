package raml

var SetOfScalarTypes = map[string]struct{}{
	String: {}, Integer: {}, Number: {}, Boolean: {}, Datetime: {}, DatetimeOnly: {},
	DateOnly: {}, TimeOnly: {}, File: {},
}

var SetOfStringFacets = map[string]struct{}{
	"minLength": {}, "maxLength": {}, "pattern": {},
}

var SetOfIntegerFacets = map[string]struct{}{
	"minimum": {}, "maximum": {},
}

var SetOfNumberFacets = map[string]struct{}{
	"multipleOf": {},
}

var SetOfFileFacets = map[string]struct{}{
	"fileTypes": {},
}

var SetOfObjectFacets = map[string]struct{}{
	"properties": {}, "additionalProperties": {}, "minProperties": {},
	"maxProperties": {}, "discriminator": {}, "discriminatorValue": {},
}

var SetOfArrayFacets = map[string]struct{}{
	"items": {}, "minItems": {}, "maxItems": {}, "uniqueItems": {},
}

// Standard types according to specification
const (
	Any          = "any"
	String       = "string"
	Integer      = "integer"
	Number       = "number"
	Boolean      = "boolean"
	Datetime     = "datetime"
	DatetimeOnly = "datetime-only"
	DateOnly     = "date-only"
	TimeOnly     = "time-only"
	Array        = "array"
	Object       = "object"
	File         = "file"
	Nil          = "nil"
)

// Special non-standard types
const (
	Union     = "union"     // Can be used in RAML
	JSON      = "json"      // Cannot be used in RAML
	Composite = "composite" // Cannot be used in RAML
)
