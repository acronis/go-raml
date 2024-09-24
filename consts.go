package raml

var SetOfScalarTypes = map[string]struct{}{
	TypeString: {}, TypeInteger: {}, TypeNumber: {}, TypeBoolean: {}, TypeDatetime: {}, TypeDatetimeOnly: {},
	TypeDateOnly: {}, TypeTimeOnly: {}, TypeFile: {},
}

var SetOfStringFacets = map[string]struct{}{
	FacetMinLength: {}, FacetMaxLength: {}, FacetPattern: {},
}

var SetOfNumberFacets = map[string]struct{}{
	FacetMinimum: {}, FacetMaximum: {}, FacetMultipleOf: {},
}

var SetOfFileFacets = map[string]struct{}{
	FacetFileTypes: {},
}

var SetOfObjectFacets = map[string]struct{}{
	FacetProperties: {}, FacetAdditionalProperties: {}, FacetMinProperties: {},
	FacetMaxProperties: {}, FacetDiscriminator: {}, FacetDiscriminatorValue: {},
}

var SetOfArrayFacets = map[string]struct{}{
	FacetItems: {}, FacetMinItems: {}, FacetMaxItems: {}, FacetUniqueItems: {},
}

var SetOfNumberFormats = map[string]struct{}{
	"float": {}, "double": {},
}

var SetOfIntegerFormats = map[string]int8{
	// int is an alias for int32
	// long is an alias for int64
	"int8": 0, "int16": 1, "int32": 2, "int": 2, "int64": 3, "long": 3,
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
	TypeNull         = "null"
)

// Special non-standard types
const (
	TypeUnion     = "union"     // Can be used in RAML
	TypeJSON      = "json"      // Cannot be used in RAML
	TypeComposite = "composite" // Cannot be used in RAML
)

const (
	TagNull      = "!!null"
	TagInclude   = "!include"
	TagStr       = "!!str"
	TagTimestamp = "!!timestamp"
	TagInt       = "!!int"
)

const (
	FacetFormat               = "format"
	FacetEnum                 = "enum"
	FacetMinimum              = "minimum"
	FacetMaximum              = "maximum"
	FacetMultipleOf           = "multipleOf"
	FacetMinLength            = "minLength"
	FacetMaxLength            = "maxLength"
	FacetPattern              = "pattern"
	FacetFileTypes            = "fileTypes"
	FacetAdditionalProperties = "additionalProperties"
	FacetProperties           = "properties"
	FacetMinProperties        = "minProperties"
	FacetMaxProperties        = "maxProperties"
	FacetItems                = "items"
	FacetMinItems             = "minItems"
	FacetMaxItems             = "maxItems"
	FacetUniqueItems          = "uniqueItems"
	FacetDiscriminator        = "discriminator"
	FacetDiscriminatorValue   = "discriminatorValue"
)
