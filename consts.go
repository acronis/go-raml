package raml

var SetOfScalarTypes = map[string]struct{}{
	TypeString: {}, TypeInteger: {}, TypeNumber: {}, TypeBoolean: {}, TypeDatetime: {}, TypeDatetimeOnly: {},
	TypeDateOnly: {}, TypeTimeOnly: {}, TypeFile: {},
}

var SetOfStringFacets = map[string]struct{}{
	"minLength": {}, "maxLength": {}, "pattern": {},
}

var SetOfNumberFacets = map[string]struct{}{
	"minimum": {}, "maximum": {}, "multipleOf": {},
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

var SetOfNumberFormats = map[string]struct{}{
	"float": {}, "double": {},
}

var SetOfIntegerFormats = map[string]struct{}{
	// int is an alias for int32
	// long is an alias for int64
	"int8": {}, "int16": {}, "int32": {}, "int": {}, "int64": {}, "long": {},
}

var SetOfDateTimeFormats = map[string]struct{}{
	"rfc3339": {}, "rfc2616": {},
}

// Standard types according to specification
const (
	TypeAny          = "any"
	TypeString       = "string"
	TypeInteger      = "integer"
	TypeNumber       = "number"
	TypeBoolean      = "boolean"
	TypeDatetime     = "datetime"
	TypeDatetimeOnly = "datetime-only"
	TypeDateOnly     = "date-only"
	TypeTimeOnly     = "time-only"
	TypeArray        = "array"
	TypeObject       = "object"
	TypeFile         = "file"
	TypeNil          = "nil"
)

// Special non-standard types
const (
	TypeUnion     = "union"     // Can be used in RAML
	TypeJSON      = "json"      // Cannot be used in RAML
	TypeComposite = "composite" // Cannot be used in RAML
)
