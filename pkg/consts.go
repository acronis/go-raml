package goraml

var SCALAR_TYPES = map[string]struct{}{
	STRING: {}, INTEGER: {}, NUMBER: {}, BOOLEAN: {}, DATETIME: {}, DATETIME_ONLY: {},
	DATE_ONLY: {}, TIME_ONLY: {}, FILE: {},
}

var STRING_FACETS = map[string]struct{}{
	"minLength": {}, "maxLength": {}, "pattern": {},
}

var INTEGER_FACETS = map[string]struct{}{
	"minimum": {}, "maximum": {},
}

var NUMBER_FACETS = map[string]struct{}{
	"multipleOf": {},
}

var FILE_FACETS = map[string]struct{}{
	"fileTypes": {},
}

var OBJECT_FACETS = map[string]struct{}{
	"properties": {}, "additionalProperties": {}, "minProperties": {},
	"maxProperties": {}, "discriminator": {}, "discriminatorValue": {},
}

var ARRAY_FACETS = map[string]struct{}{
	"items": {}, "minItems": {}, "maxItems": {}, "uniqueItems": {},
}

const (
	// Standard types according to specification
	ANY           = "any"
	STRING        = "string"
	INTEGER       = "integer"
	NUMBER        = "number"
	BOOLEAN       = "boolean"
	DATETIME      = "datetime"
	DATETIME_ONLY = "datetime-only"
	DATE_ONLY     = "date-only"
	TIME_ONLY     = "time-only"
	ARRAY         = "array"
	OBJECT        = "object"
	FILE          = "file"
	NIL           = "nil"

	// Special non-standard types
	UNION     = "union"     // Can be used in RAML
	JSON      = "json"      // Cannot be used in RAML
	COMPOSITE = "composite" // Cannot be used in RAML
)
