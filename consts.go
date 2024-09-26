package raml

// SetOfScalarTypes contains a set of scalar types
var SetOfScalarTypes = map[string]struct{}{
	TypeString: {}, TypeInteger: {}, TypeNumber: {}, TypeBoolean: {}, TypeDatetime: {}, TypeDatetimeOnly: {},
	TypeDateOnly: {}, TypeTimeOnly: {}, TypeFile: {},
}

// SetOfStringFacets contains a set of string facets
var SetOfStringFacets = map[string]struct{}{
	FacetMinLength: {}, FacetMaxLength: {}, FacetPattern: {},
}

// SetOfNumberFacets contains a set of number facets
var SetOfNumberFacets = map[string]struct{}{
	FacetMinimum: {}, FacetMaximum: {}, FacetMultipleOf: {},
}

// SetOfFileFacets contains a set of file facets
var SetOfFileFacets = map[string]struct{}{
	FacetFileTypes: {},
}

// SetOfObjectFacets contains a set of object facets
var SetOfObjectFacets = map[string]struct{}{
	FacetProperties: {}, FacetAdditionalProperties: {}, FacetMinProperties: {},
	FacetMaxProperties: {}, FacetDiscriminator: {}, FacetDiscriminatorValue: {},
}

// SetOfArrayFacets contains a set of array facets
var SetOfArrayFacets = map[string]struct{}{
	FacetItems: {}, FacetMinItems: {}, FacetMaxItems: {}, FacetUniqueItems: {},
}

// SetOfNumberFormats contains a set of number formats
var SetOfNumberFormats = map[string]struct{}{
	"float": {}, "double": {},
}

// SetOfIntegerFormats contains a set of integer formats
var SetOfIntegerFormats = map[string]int8{
	// int is an alias for int32
	// long is an alias for int64
	"int8": 0, "int16": 1, "int32": 2, "int": 2, "int64": 3, "long": 3,
}

// SetOfDateTimeFormats contains a set of date-time formats
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
