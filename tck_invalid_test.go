package raml

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/acronis/go-stacktrace"
)

// tckErrSpec describes what must hold for an error returned by an invalid TCK fixture.
type tckErrSpec struct {
	// fixture is the path relative to the raml-tck directory (forward slashes).
	fixture string
	// msgContains is a substring that must appear in at least one node's Message
	// in the StackTrace tree.
	msgContains string
	// line/col are the expected 1-based start position (-1 = don't check).
	line, col int
	// endLine/endCol are the expected 1-based end position (-1 = don't check).
	endLine, endCol int
	// extraOpts are additional ParseOpt values merged with the default
	// (OptWithUnwrap, OptWithValidate) for this fixture only.
	extraOpts []ParseOpt
}

// failingRoundTripper is an http.RoundTripper that always returns a 404 response.
// Used in TCK tests to verify error handling for remote !include references
// without making real network requests.
type failingRoundTripper struct{}

func (failingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusNotFound,
		Status:     "404 Not Found",
		Header:     make(http.Header),
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

// Test_TCKInvalidErrors verifies that specific invalid TCK fixtures yield errors
// with the expected messages and source positions.
func Test_TCKInvalidErrors(t *testing.T) {
	tckDir := "./raml-tck"
	if _, err := os.Stat(tckDir); err != nil {
		t.Skipf("raml-tck directory not found at %s: %v", tckDir, err)
	}

	// Count all invalid*.raml files in the TCK directory for coverage tracking.
	var totalInvalidFiles int
	_ = filepath.WalkDir(tckDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasPrefix(filepath.Base(path), "invalid") && strings.HasSuffix(path, ".raml") {
			totalInvalidFiles++
		}
		return nil
	})

	// Auto-generated code with internal\cmd\tck-audit\main.go
	// Fixtures marked TODO currently produce no error and need implementation fixes.
	cases := []tckErrSpec{
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/complex-01/invalid-wrong-target.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected nil"
		{
			fixture:     "Annotations/complex-02/invalid-nil-annotation.raml",
			msgContains: "invalid type, got string, expected nil",
			line:        20, col: 19, endLine: 20, endCol: 28,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/complex-03/invalid-target.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.childProperty1: value must be one of (enum_defaultValue, enum_value1, enum_value2)"
		{
			fixture:     "Annotations/complex-05/invalid-enum.raml",
			msgContains: "value must be one of (enum_defaultValue, enum_value1, enum_value2)",
			line:        36, col: 7, endLine: 36, endCol: 33,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "missing required parameter"
		{
			fixture:     "Annotations/complex-06/invalid-param-missing.raml",
			msgContains: "missing required parameter",
			line:        7, col: 5, endLine: 13, endCol: 39,
		},
		{
			fixture:     "Annotations/complex-08/invalid-undefined-property.raml",
			msgContains: "validate properties: unexpected additional property \"hi\"",
			line:        17, col: 15, endLine: 17, endCol: 20,
		},
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected []any"
		{
			fixture:     "Annotations/complex-10/invalid-wrong-type.raml",
			msgContains: "invalid type, got string, expected []any",
			line:        9, col: 28, endLine: 9, endCol: 33,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/complex-11/invalid-multiple-annots.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Annotations/other-02/invalid-prop-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        6, col: 8, endLine: 6, endCol: 10,
		},
		// error chain: "validate shapes" -> "check domain extension: value must be greater than or equal to 2"
		{
			fixture:     "Annotations/other-03/invalid-min.raml",
			msgContains: "value must be greater than or equal to 2",
			line:        8, col: 8, endLine: 8, endCol: 9,
		},
		// error chain: "validate shapes" -> "check domain extension: value must be less than or equal to 10"
		{
			fixture:     "Annotations/other-04/invalid-max.raml",
			msgContains: "value must be less than or equal to 10",
			line:        8, col: 8, endLine: 8, endCol: 10,
		},
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected bool"
		{
			fixture:     "Annotations/other-05/invalid-prop-type.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        14, col: 11, endLine: 14, endCol: 17,
		},
		// error chain: "resolve domain extensions" -> "resolve domain extension" -> "get referenced shape: reference \"foobar\" not found"
		{
			fixture:     "Annotations/other-06/invalid-undefined-annotation.raml",
			msgContains: "reference \"foobar\" not found",
			line:        9, col: 5, endLine: 9, endCol: 13,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.point: validate properties: missing required properties: y"
		{
			fixture:     "Annotations/resource-01/invalid-missing-required-prop.raml",
			msgContains: "missing required properties",
			line:        17, col: 7, endLine: 18, endCol: 13,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.point: validate properties: missing required properties: x"
		{
			fixture:     "Annotations/resource-02/invalid-missing-required-prop.raml",
			msgContains: "missing required properties",
			line:        17, col: 7, endLine: 18, endCol: 13,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.point: validate array item $.point[1]: validate properties: unexpected additional property \"z\""
		{
			fixture:     "Annotations/resource-03/invalid-not-allowed-prop.raml",
			msgContains: "unexpected additional property \"z\"",
			line:        25, col: 11, endLine: 25, endCol: 15,
		},
		// error chain: "validate shapes" -> "check domain extension" -> "validate properties" -> "validate pattern property $.role: C:\\Sources\\go-raml\\raml-tck\\Annotations\\resource-04\\invalid-pattern-prop.raml:8:7" -> "validate pattern property: invalid type, got string, expected map[string]any"
		{
			fixture:     "Annotations/resource-04/invalid-pattern-prop.raml",
			msgContains: "invalid type, got string, expected map[string]any",
			line:        8, col: 7, endLine: 8, endCol: 11,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: missing required properties: prop1"
		{
			fixture:     "Annotations/resource-05/invalid-pattern-prop.raml",
			msgContains: "missing required properties",
			line:        11, col: 9, endLine: 14, endCol: 20,
		},
		// error chain: "resolve domain extensions" -> "resolve domain extension" -> "get referenced shape: reference \"suborg1\" not found"
		{
			fixture:     "Annotations/resource-06/invalid-undefined-annotation.raml",
			msgContains: "reference \"suborg1\" not found",
			line:        14, col: 3, endLine: 14, endCol: 12,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.level: value must be one of (low, medium, high)"
		{
			fixture:     "Annotations/resource-07/invalid-enum-val.raml",
			msgContains: "value must be one of (low, medium, high)",
			line:        15, col: 5, endLine: 15, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Annotations/resource-08/invalid-duplicate-prop.raml",
			msgContains: "unknown facet",
			line:        11, col: 5, endLine: 11, endCol: 10,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.items: value must be one of (W, A)"
		{
			fixture:     "Annotations/root-01/invalid-enum-val.raml",
			msgContains: "value must be one of (W, A)",
			line:        12, col: 7, endLine: 12, endCol: 15,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.q: invalid type, got string, expected bool"
		{
			fixture:     "Annotations/root-02/invalid-boolean-val.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        11, col: 7, endLine: 11, endCol: 15,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: missing required properties: items"
		{
			fixture:     "Annotations/root-03/invalid-missing-required-prop.raml",
			msgContains: "missing required properties",
			line:        11, col: 7, endLine: 11, endCol: 14,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.v: validate array item $.v[0]: invalid type, got string, expected bool"
		{
			fixture:     "Annotations/root-04/invalid-bools-array.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        11, col: 8, endLine: 11, endCol: 13,
		},
		// error chain: "validate shapes" -> "check domain extension" -> "validate pattern property" -> "validate properties" -> "validate pattern property $.persons.156798654.Alice: C:\\Sources\\go-raml\\raml-tck\\Annotations\\root-05\\invalid-inheritance.raml:10:7" -> "validate pattern property: validate properties: missing required properties: name"
		{
			fixture:     "Annotations/root-05/invalid-inheritance.raml",
			msgContains: "validate shapes",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "check domain extension" -> "validate properties" -> "validate property $.prop1" -> "validate pattern property $.prop1.prop1: C:\\Sources\\go-raml\\raml-tck\\Annotations\\root-06\\invalid-num-pattern-prop.raml:8:11" -> "validate pattern property: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Annotations/root-06/invalid-num-pattern-prop.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        8, col: 11, endLine: 8, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet" -> "check domain extension: length must be greater than 10"
		{
			fixture:     "Annotations/root-07/invalid-min-length.raml",
			msgContains: "length must be greater than 10",
			line:        17, col: 11, endLine: 17, endCol: 20,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/root-08/invalid-target.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Annotations/root-09/invalid-inheritance.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        12, col: 27, endLine: 12, endCol: 29,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.numberProp: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Annotations/root-10/invalid-inherit-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        11, col: 3, endLine: 11, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Annotations/scalar-values-annotated/invalid-missing-value.raml",
			msgContains: "unknown field in annotated scalar",
			line:        9, col: 3, endLine: 9, endCol: 20,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-annotation-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-api-used-in-annotation.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-doc-item-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-example-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-method-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-request-body-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-resource-type-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-resource-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-response-body-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-response-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-security-scheme-settings.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-security-scheme-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-trait-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Annotations/target-locations/invalid-type-declaration-used-in-api.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "parse api" -> "decode api: yaml: unknown anchor 'second' referenced"
		{
			fixture:     "EdgeCases/anchor-refs-tmpl-params/invalid-inexisting-anchor.raml",
			msgContains: "unknown anchor 'second' referenced",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got string, expected []any"
		{
			fixture:     "EdgeCases/array-example/invalid.raml",
			msgContains: "invalid type, got string, expected []any",
			line:        13, col: 14, endLine: 13, endCol: 19,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown authorization grant"
		{
			fixture:     "EdgeCases/authgrants-element/invalid-authgrants-element.raml",
			msgContains: "unknown authorization grant",
			line:        16, col: 50, endLine: 16, endCol: 63,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "value does not match any type" -> "validate union member: value must match format 2006-01-02" -> "validate union member: value must match format 15:04:05"
		{
			fixture:     "EdgeCases/dates-union/invalid-wrong-date-format.raml",
			msgContains: "value must match format 15:04:05",
			line:        8, col: 3, endLine: 8, endCol: 9,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\datetime-type\\invalid-datetime-type.raml:7:3" -> "unwrap shape" -> "unwrap parents" -> "multiple parents unwrap" -> "merge shapes" -> "cannot inherit from different type"
		{
			fixture:     "EdgeCases/datetime-type/invalid-datetime-type.raml",
			msgContains: "cannot inherit from different type",
			line:        7, col: 12, endLine: 7, endCol: 20,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "EdgeCases/discriminator-inline/invalid-discriminator-inline.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression" -> "make concrete shape yaml: C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\discriminator-union\\invalid-discriminator-union.raml:21:3" -> "unmarshal yaml nodes" -> "discriminator is not allowed on union types"
		{
			fixture:     "EdgeCases/discriminator-union/invalid-discriminator-union.raml",
			msgContains: "discriminator is not allowed on union types",
			line:        23, col: 4, endLine: 23, endCol: 17,
		},
		// error chain: "apply security schemes" -> "apply security scheme" -> "get security scheme definition: library \"oauth2\" not found"
		{
			fixture:     "EdgeCases/dot-in-securityscheme-name/invalid.raml",
			msgContains: "library \"oauth2\" not found",
			line:        20, col: 17, endLine: 20, endCol: 25,
		},
		// error chain: "parse api" -> "decode api" -> "title must not be empty"
		{
			fixture:     "EdgeCases/empty-title/invalid-empty-title.raml",
			msgContains: "title must not be empty",
			line:        2, col: 1, endLine: 2, endCol: 6,
		},
		// error chain: "parse library" -> "decode library" -> "unmarshal trait definitions" -> "make traits" -> "decode trait definition" -> "trait definition must be a mapping node"
		{
			fixture:     "EdgeCases/empty-trait/invalid-trait-structure.raml",
			msgContains: "trait definition must be a mapping node",
			line:        3, col: 10, endLine: 3, endCol: 13,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got int, expected bool"
		{
			fixture:     "EdgeCases/enum-booleans/invalid-wrong-enum-item-type.raml",
			msgContains: "invalid type, got int, expected bool",
			line:        9, col: 12, endLine: 9, endCol: 13,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got bool, expected string"
		{
			fixture:     "EdgeCases/enum-dates/invalid-wrong-enum-item-type.raml",
			msgContains: "invalid type, got bool, expected string",
			line:        9, col: 12, endLine: 9, endCol: 17,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got bool, expected a numeric type or json.Number"
		{
			fixture:     "EdgeCases/enum-integers/invalid-wrong-enum-item-type.raml",
			msgContains: "invalid type, got bool, expected a numeric type or json.Number",
			line:        9, col: 18, endLine: 9, endCol: 23,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got bool, expected a numeric type or json.Number"
		{
			fixture:     "EdgeCases/enum-numbers/invalid-wrong-enum-item-type.raml",
			msgContains: "invalid type, got bool, expected a numeric type or json.Number",
			line:        9, col: 18, endLine: 9, endCol: 23,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got int, expected string"
		{
			fixture:     "EdgeCases/enum-strings/invalid-wrong-enum-item-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        9, col: 17, endLine: 9, endCol: 18,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "EdgeCases/identifying-discriminator/invalid-inexisting-descriminator.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "parse data" -> "decode data type" -> "must be map"
		{
			fixture:     "EdgeCases/include-empty-file/invalid-include-invalid-raml.raml",
			msgContains: "must be map",
			line:        2, col: 1, endLine: 2, endCol: 27,
		},
		// error chain: "parse data type" -> "decode data type" -> "must be map"
		{
			fixture:     "EdgeCases/include-empty-file/invalid-user.raml",
			msgContains: "must be map",
			line:        2, col: 1, endLine: 2, endCol: 27,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "EdgeCases/include-no-whitespace/invalid-include-no-whitespace.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "parse resource type fragment" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\includes-resolution\\resourceTypes\\idontexist.raml: The system cannot find the file specified."
		{
			fixture:     "EdgeCases/includes-resolution/invalid-include-inexisting-file.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"date\" not found" -> "resolve multiple inheritance" -> "resolve inherit"
		{
			fixture:     "EdgeCases/inherit-multiple-scalars/invalid-inherit-multiple-scalars.raml",
			msgContains: "reference \"date\" not found",
			line:        10, col: 11, endLine: 10, endCol: 15,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: get reference type: query.idontexist: reference \"idontexist\" not found"
		{
			fixture:     "EdgeCases/inheriting-unknown-type/invalid-inherit-unknown-type.raml",
			msgContains: "reference \"idontexist\" not found",
			line:        10, col: 5, endLine: 10, endCol: 16,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "unmarshal yaml node" -> "multipleOf must not be zero"
		{
			fixture:     "EdgeCases/integer-zero-division/invalid-integer-zero-division.raml",
			msgContains: "multipleOf must not be zero",
			line:        9, col: 17, endLine: 9, endCol: 18,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "EdgeCases/invalid-usage-node/invalid.raml",
			msgContains: "unknown facet",
			line:        5, col: 5, endLine: 5, endCol: 10,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "maxLength" -> "decode: cannot unmarshal !!int `-1` into uint64"
		{
			fixture:     "EdgeCases/maxlength-negative-value/invalid.raml",
			msgContains: "cannot unmarshal !!int `-1` into uint64",
			line:        6, col: 16, endLine: 6, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal media type" -> "media type: invalid media type someStringvalue"
		{
			fixture:     "EdgeCases/media-type/invalid-media-type.raml",
			msgContains: "invalid media type someStringvalue",
			line:        6, col: 12, endLine: 6, endCol: 27,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "make parametrized endpoint" -> "decode parametrized endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "explicit media type is required"
		{
			fixture:     "EdgeCases/mediaType-param/invalid.raml",
			msgContains: "explicit media type is required",
			line:        8, col: 11, endLine: 8, endCol: 15,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "minLength" -> "decode: cannot unmarshal !!int `-2` into uint64"
		{
			fixture:     "EdgeCases/minlength-negative-value/invalid.raml",
			msgContains: "cannot unmarshal !!int `-2` into uint64",
			line:        6, col: 16, endLine: 6, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "requestTokenUri is required for OAuth 1.0"
		{
			fixture:     "EdgeCases/missing-oauth1-settings/invalid-missing-oauth1-settings.raml",
			msgContains: "requestTokenUri is required for OAuth 1.0",
			line:        12, col: 11, endLine: 12, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "accessTokenUri is required for OAuth 2.0"
		{
			fixture:     "EdgeCases/missing-oauth2-settings/invalid-missing-oauth2-settings.raml",
			msgContains: "accessTokenUri is required for OAuth 2.0",
			line:        12, col: 11, endLine: 12, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: value must be a multiple of 3"
		{
			fixture:     "EdgeCases/multipleof-example/invalid-example.raml",
			msgContains: "value must be a multiple of 3",
			line:        8, col: 14, endLine: 8, endCol: 19,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "multipleOf must not be zero"
		{
			fixture:     "EdgeCases/multipleof-integer/invalid-multipleof-zero.raml",
			msgContains: "multipleOf must not be zero",
			line:        6, col: 17, endLine: 6, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "invalid multipleOf value"
		{
			fixture:     "EdgeCases/multipleof-string/invalid-multipleof-string.raml",
			msgContains: "invalid multipleOf value",
			line:        7, col: 17, endLine: 7, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make shape type" -> "parse data" -> "make json data type" -> "decode fragment" -> "parse types: make shape" -> "make json shape" -> "compile schema: failing loading \"file:///C:/Sources/go-raml/raml-tck/EdgeCases/nested-json-schema/ref/inexisting-h45h45h-partner-schema.json\": open C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\nested-json-schema\\ref\\inexisting-h45h45h-partner-schema.json: The system cannot find the file specified."
		{
			fixture:     "EdgeCases/nested-json-schema/invalid-refer-inexisting-nested-schema.raml",
			msgContains: "file does not exist",
			line:        1, col: 0, endLine: 0, endCol: 27,
		},
		// error chain: "parse data type" -> "parse uses library" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\nested-lib-uses\\inexisting-library.raml: The system cannot find the file specified."
		{
			fixture:     "EdgeCases/nested-lib-uses/invalid-data-type.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "parse data" -> "parse uses library" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\nested-lib-uses\\inexisting-library.raml: The system cannot find the file specified."
		{
			fixture:     "EdgeCases/nested-lib-uses/invalid-refer-nested-inexisting-lib.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "multipleOf must not be zero"
		{
			fixture:     "EdgeCases/number-zero-division/invalid-number-zero-division.raml",
			msgContains: "multipleOf must not be zero",
			line:        9, col: 17, endLine: 9, endCol: 18,
		},
		// error chain: "validate shapes" -> "check type" -> "invalid format"
		{
			fixture:     "EdgeCases/numeric-formats/invalid-integer-format.raml",
			msgContains: "invalid format",
			line:        12, col: 13, endLine: 12, endCol: 21,
		},
		// error chain: "validate shapes" -> "check type" -> "invalid format"
		{
			fixture:     "EdgeCases/numeric-formats/invalid-number-format.raml",
			msgContains: "invalid format",
			line:        12, col: 13, endLine: 12, endCol: 21,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown authorization grant"
		{
			fixture:     "EdgeCases/oauth2-grants-url/invalid-not-absolute-uri.raml",
			msgContains: "unknown authorization grant",
			line:        16, col: 30, endLine: 16, endCol: 41,
		},
		{
			fixture:     "EdgeCases/overlay-overrides-resources/invalid-adds-resources.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "duplicate custom facet"
		{
			fixture:     "EdgeCases/override-parent-facet/invalid-override-parent-facet.raml",
			msgContains: "duplicate custom facet",
			line:        9, col: 7, endLine: 9, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal trait definitions" -> "make traits" -> "compile trait definition" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "responses must be a mapping node"
		{
			fixture:     "EdgeCases/param-resp-definitions/invalid-node-is-scalar.raml",
			msgContains: "responses must be a mapping node",
			line:        9, col: 12, endLine: 9, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "decode" -> "decode value node" -> "decode facets" -> "cannot redefine built-in facet"
		{
			fixture:     "EdgeCases/parsing-facets/invalid-override-builtin-facet.raml",
			msgContains: "cannot redefine built-in facet",
			line:        12, col: 7, endLine: 12, endCol: 18,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\EdgeCases\\property-type-conflict\\invalid-property-type-conflict.raml:20:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "inherit property" -> "cannot inherit from different type"
		{
			fixture:     "EdgeCases/property-type-conflict/invalid-property-type-conflict.raml",
			msgContains: "cannot inherit from different type",
			line:        16, col: 7, endLine: 16, endCol: 11,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-any/invalid-redefine-any.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 6,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-array/invalid-redefine-array.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 8,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-boolean/invalid-redefine-boolean.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 10,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-date-only/invalid-redefine-date-only.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 12,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-datetime-only/invalid-redefine-datetime-only.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 16,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-datetime/invalid-redefine-datetime.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 11,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-file/invalid-redefine-file.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 7,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-integer/invalid-redefine-integer.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 10,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-nil/invalid-redefine-nil.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 6,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-number/invalid-redefine-number.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 9,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-object/invalid-redefine-object.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 9,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-string/invalid-redefine-string.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 9,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "EdgeCases/redefine-time-only/invalid-redefine-time-only.raml",
			msgContains: "cannot redefine built-in type",
			line:        6, col: 3, endLine: 6, endCol: 12,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "make parametrized endpoint" -> "decode parametrized endpoint" -> "make operation" -> "decode operation" -> "unknown field"
		{
			fixture:     "EdgeCases/resourcetype-application/invalid-resourcetype-application.raml",
			msgContains: "unknown field",
			line:        14, col: 7, endLine: 14, endCol: 10,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "make parametrized endpoint" -> "decode parametrized endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make shape type" -> "parse data" -> "make json data type" -> "decode fragment" -> "parse types: make shape" -> "make json shape" -> "compile schema: json-pointer in \"...invalid-list.json#/definitions/BusinessModelObject\" not found"
		{
			fixture:     "EdgeCases/schemas-inner-definitions/invalid-references-invalid-json-schema.raml",
			msgContains: "invalid-list.json#/definitions/BusinessModelObject\" not found",
			line:        1, col: 0, endLine: 0, endCol: 17,
		},
		// error chain: "parse api" -> "decode api" -> "types and schemas are mutually exclusive"
		{
			fixture:     "EdgeCases/schemas-types-exclusive/invalid.raml",
			msgContains: "types and schemas are mutually exclusive",
			line:        5, col: 7, endLine: 5, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme description" -> "decode security scheme description" -> "security scheme describedBy must be a mapping node"
		{
			fixture:     "EdgeCases/security-describedby/invalid.raml",
			msgContains: "security scheme describedBy must be a mapping node",
			line:        9, col: 18, endLine: 9, endCol: 24,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate default: invalid type, got int, expected string"
		{
			fixture:     "EdgeCases/string-in-angle-brackets/invalid.raml",
			msgContains: "invalid type, got int, expected string",
			line:        10, col: 18, endLine: 10, endCol: 24,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "parse security scheme fragment" -> "decode security scheme fragment" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "EdgeCases/unrecognized-security-scheme/invalid-scheme-type.raml",
			msgContains: "unknown security scheme type",
			line:        2, col: 7, endLine: 2, endCol: 22,
		},
		// error chain: "parse security scheme fragment" -> "decode security scheme fragment" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "EdgeCases/unrecognized-security-scheme/securitySchemes/invalid-scheme.raml",
			msgContains: "unknown security scheme type",
			line:        2, col: 7, endLine: 2, endCol: 22,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "validate uri parameters" -> "uri parameter value contains slash" -> "default value must not contain '/'"
		{
			fixture:     "EdgeCases/uriparam-default-slash/invalid-uriparam-default-slash.raml",
			msgContains: "default value must not contain '/'",
			line:        9, col: 16, endLine: 9, endCol: 29,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "validate uri parameters" -> "uri parameter value contains slash" -> "enum value must not contain '/'"
		{
			fixture:     "EdgeCases/uriparam-enum-slash/invalid-uriparam-enum-slash.raml",
			msgContains: "enum value must not contain '/'",
			line:        9, col: 15, endLine: 9, endCol: 28,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "validate uri parameters" -> "uri parameter value contains slash" -> "example value must not contain '/'"
		{
			fixture:     "EdgeCases/uriparam-example-slash/invalid-uriparam-example-slash.raml",
			msgContains: "example value must not contain '/'",
			line:        9, col: 16, endLine: 9, endCol: 29,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "validate uri parameters" -> "uri parameter value contains slash" -> "example value must not contain '/'"
		{
			fixture:     "EdgeCases/uriparam-examples-slash/invalid-uriparam-examples-slash.raml",
			msgContains: "example value must not contain '/'",
			line:        10, col: 19, endLine: 10, endCol: 32,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Fragments/annotation/includes/invalid-wrong-structure.raml",
			msgContains: "unknown facet",
			line:        3, col: 1, endLine: 3, endCol: 5,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet" -> "check domain extension: invalid type, got map[string]interface {}, expected string"
		{
			fixture:     "Fragments/annotation/invalid-annotation-included.raml",
			msgContains: "invalid type, got map[string]interface {}, expected string",
			line:        10, col: 5, endLine: 11, endCol: 32,
		},
		// error chain: "parse data type" -> "decode data type: mapping values are not allowed in this context"
		{
			fixture:     "Fragments/datatype-decode/invalid-dtype-decode.raml",
			msgContains: "mapping values are not allowed in this context",
			line:        5, col: 1, endLine: 0, endCol: 0,
		},
		{
			fixture:     "Fragments/datatype-decode/invalid-dtype-header.raml",
			msgContains: "#%RAML 1.0 Data",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse named example" -> "decode named example: mapping values are not allowed in this context"
		{
			fixture:     "Fragments/datatype-decode/invalid-named-example-decode.raml",
			msgContains: "mapping values are not allowed in this context",
			line:        7, col: 1, endLine: 0, endCol: 0,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Fragments/datatype/includes/invalid-nodes.raml",
			msgContains: "unknown facet",
			line:        10, col: 1, endLine: 10, endCol: 3,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Fragments/datatype/invalid-datatype-included.raml",
			msgContains: "unknown facet",
			line:        10, col: 1, endLine: 10, endCol: 3,
		},
		// error chain: "parse documentation item fragment" -> "decode documentation item" -> "unknown field"
		{
			fixture:     "Fragments/documentationitem/includes/invalid-wrong-nodes.raml",
			msgContains: "unknown field",
			line:        7, col: 1, endLine: 7, endCol: 6,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "parse documentation item fragment" -> "decode documentation item fragment" -> "decode documentation item" -> "unknown field"
		{
			fixture:     "Fragments/documentationitem/invalid-docitem-included.raml",
			msgContains: "unknown field",
			line:        7, col: 1, endLine: 7, endCol: 6,
		},
		{
			fixture:     "Fragments/extend-with-new-method/invalid-inexisting-base.raml",
			msgContains: "#%RAML 1.0 Extension",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Fragments/extension/invalid-nodes.raml",
			msgContains: "#%RAML 1.0 Extension",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Fragments/namedexample-01/examples/invalid-one-example.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Fragments/namedexample-01/invalid-includes-incorrect-named-example.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "parse named example" -> "decode named example" -> "must be map"
		{
			fixture:     "Fragments/namedexample-02/examples/invalid-meaningless-content.raml",
			msgContains: "must be map",
			line:        3, col: 1, endLine: 3, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unmarshal query parameters" -> "make property" -> "make shape" -> "decode" -> "decode value node" -> "decode example" -> "parse named example" -> "decode named example" -> "must be map"
		{
			fixture:     "Fragments/namedexample-02/invalid-meaningless-examples-content.raml",
			msgContains: "must be map",
			line:        3, col: 1, endLine: 3, endCol: 7,
		},
		// error chain: "parse resource type fragment" -> "decode resource type fragment" -> "make resource type definition" -> "decode resource type definition" -> "resource type method must be an HTTP method"
		{
			fixture:     "Fragments/resourcetype/includes/invalid-nodes.raml",
			msgContains: "resource type method must be an HTTP method",
			line:        13, col: 1, endLine: 13, endCol: 3,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "parse resource type fragment" -> "decode resource type fragment" -> "resource type method must be an HTTP method"
		{
			fixture:     "Fragments/resourcetype/invalid-nodes-in-resourcetype.raml",
			msgContains: "resource type method must be an HTTP method",
			line:        13, col: 1, endLine: 13, endCol: 3,
		},
		// error chain: "parse security scheme fragment" -> "decode security scheme fragment" -> "make security scheme definition" -> "decode security scheme definition" -> "unknown key in security scheme definition"
		{
			fixture:     "Fragments/securityscheme/includes/invalid-nodes.raml",
			msgContains: "unknown key in security scheme definition",
			line:        14, col: 1, endLine: 14, endCol: 3,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "parse security scheme fragment" -> "decode security scheme fragment" -> "unknown key in security scheme definition"
		{
			fixture:     "Fragments/securityscheme/invalid-nodes-in-security-scheme.raml",
			msgContains: "unknown key in security scheme definition",
			line:        14, col: 1, endLine: 14, endCol: 3,
		},
		// error chain: "parse library" -> "decode library" -> "unknown field"
		{
			fixture:     "Fragments/simple-library/invalid-nodes.raml",
			msgContains: "unknown field",
			line:        20, col: 1, endLine: 20, endCol: 3,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "unknown field"
		{
			fixture:     "Fragments/using-libraries/invalid-chaining.raml",
			msgContains: "unknown field",
			line:        10, col: 3, endLine: 10, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "parse resource type fragment" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Libraries\\include-01\\<<version>>.raml: The filename, directory name, or volume label syntax is incorrect."
		{
			fixture:     "Libraries/include-01/invalid-dynamic-inclusion.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "parse resource type fragment" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Libraries\\include-01\\f31f23f23f23f23f.raml: The system cannot find the file specified."
		{
			fixture:     "Libraries/include-01/invalid-include-inexisting.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "apply resource types" -> "get resource type definition: library \"files-resource\" not found"
		{
			fixture:     "Libraries/include-02/invalid-include-in-wrong-place.raml",
			msgContains: "library \"files-resource\" not found",
			line:        5, col: 9, endLine: 5, endCol: 37,
		},
		// error chain: "parse library" -> "decode library: did not find expected key"
		{
			fixture:     "Libraries/library-validation/invalid-decode.raml",
			msgContains: "did not find expected key",
			line:        2, col: 1, endLine: 0, endCol: 0,
		},
		// error chain: "parse library" -> "parse uses library" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Libraries\\library-validation\\library.raml: The system cannot find the file specified."
		{
			fixture:     "Libraries/library-validation/invalid-dot-import.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"B\" not found"
		{
			fixture:     "Libraries/library-validation/invalid-inheritance.raml",
			msgContains: "reference \"B\" not found",
			line:        4, col: 3, endLine: 4, endCol: 4,
		},
		// error chain: "validate shapes" -> "check type" -> "minProperties must be less than or equal to maxProperties"
		{
			fixture:     "Libraries/library-validation/invalid-library.raml",
			msgContains: "minProperties must be less than or equal to maxProperties",
			line:        6, col: 20, endLine: 6, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate custom facet definition" -> "required custom facet is missing"
		{
			fixture:     "Libraries/library-validation/invalid-recursive-trait.raml",
			msgContains: "required custom facet is missing",
			line:        7, col: 7, endLine: 7, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate custom facet definition" -> "validate example: invalid type, got int, expected string"
		{
			fixture:     "Libraries/library-validation/invalid-trait-example.raml",
			msgContains: "invalid type, got int, expected string",
			line:        9, col: 18, endLine: 9, endCol: 19,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Libraries\\library-validation\\invalid-unwrap.raml:8:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "inherit property" -> "cannot inherit from different type"
		{
			fixture:     "Libraries/library-validation/invalid-unwrap.raml",
			msgContains: "cannot inherit from different type",
			line:        11, col: 9, endLine: 11, endCol: 10,
		},
		// error chain: "parse library" -> "decode library" -> "unknown field"
		{
			fixture:     "Libraries/standalone/invalid-resource-defined.raml",
			msgContains: "unknown field",
			line:        32, col: 1, endLine: 32, endCol: 7,
		},
		// error chain: "parse api" -> "parse uses library" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Libraries\\uses-01\\lib123.raml: The system cannot find the file specified."
		{
			fixture:     "Libraries/uses-01/invalid-uses-inexisting-lib.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "parse uses library" -> "check fragment kind: unexpected fragment frag != kind: API != Library"
		{
			fixture:     "Libraries/uses-02/invalid-uses-non-lib.raml",
			msgContains: "API != Library",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "failed to validate against JSON schema: jsonschema validation failed with 'file:///C:/Sources/go-raml/raml-tck/MethodResponses/body-schema-json-01/invalid-conform-schema.raml#'\n- at '': additional properties 'r' not allowed"
		{
			fixture:     "MethodResponses/body-schema-json-01/invalid-conform-schema.raml",
			msgContains: "additional properties 'r' not allowed",
			line:        13, col: 13, endLine: 25, endCol: 47,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "failed to validate against JSON schema: jsonschema validation failed with 'file:///C:/Sources/go-raml/raml-tck/MethodResponses/body-schema-json-02/invalid-conform-schema.raml#'\n- at '/message': got number, want string"
		{
			fixture:     "MethodResponses/body-schema-json-02/invalid-conform-schema.raml",
			msgContains: "got number, want string",
			line:        10, col: 13, endLine: 23, endCol: 25,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make shape type" -> "make json shape" -> "validate json schema: jsonschema validation failed with 'http://json-schema.org/draft-07/schema#'\n- at '/properties/message/required': got boolean, want array\n- at '/required': got boolean, want array"
		{
			fixture:     "MethodResponses/complex-json-schemes/invalid-schema-json.raml",
			msgContains: "got boolean, want array",
			line:        36, col: 13, endLine: 36, endCol: 29,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: x"
		{
			fixture:     "MethodResponses/example-json/invalid-json.raml",
			msgContains: "missing required properties",
			line:        12, col: 22, endLine: 12, endCol: 39,
		},
		// error chain: "validate shapes" -> "check type" -> "check properties: C:\\Sources\\go-raml\\raml-tck\\MethodResponses\\inline-schema-01\\invalid-missing-req-params.raml:18:15" -> "check property" -> "invalid format"
		{
			fixture:     "MethodResponses/inline-schema-01/invalid-missing-req-params.raml",
			msgContains: "invalid format",
			line:        22, col: 25, endLine: 22, endCol: 28,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: field"
		{
			fixture:     "MethodResponses/inline-using-datatype-01/invalid-missing-property.raml",
			msgContains: "missing required properties",
			line:        24, col: 19, endLine: 25, endCol: 36,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: field"
		{
			fixture:     "MethodResponses/inline-using-datatype-02/invalid-missing-property.raml",
			msgContains: "missing required properties",
			line:        22, col: 15, endLine: 23, endCol: 32,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: c"
		{
			fixture:     "MethodResponses/inline-using-datatype-03/invalid-missing-req-property.raml",
			msgContains: "missing required properties",
			line:        22, col: 13, endLine: 22, endCol: 17,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: c"
		{
			fixture:     "MethodResponses/inline-using-datatype-04/invalid-missing-req-property.raml",
			msgContains: "missing required properties",
			line:        25, col: 13, endLine: 25, endCol: 17,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.items: validate array item $.items[1]: validate properties: missing required properties: c"
		{
			fixture:     "MethodResponses/inline-using-datatype-05/invalid-missing-req-property.raml",
			msgContains: "missing required properties",
			line:        27, col: 13, endLine: 33, endCol: 22,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "value does not match any type" -> "validate union member: validate properties: missing required properties: name, color" -> "validate union member: validate properties: missing required properties: name"
		{
			fixture:     "MethodResponses/inline-using-datatype-06/invalid-missing-req-property.raml",
			msgContains: "missing required properties",
			line:        7, col: 3, endLine: 7, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "decode" -> "decode value node" -> "decode example" -> "make example" -> "make node" -> "include: open C:\\Sources\\go-raml\\raml-tck\\MethodResponses\\inline-using-datatype-lib\\examples\\fewfefwefwefwef.yaml: The system cannot find the file specified."
		{
			fixture:     "MethodResponses/inline-using-datatype-lib/invalid-inexisting-example-file.raml",
			msgContains: "file does not exist",
			line:        15, col: 20, endLine: 15, endCol: 58,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "value does not match any type" -> "validate union member: validate properties: missing required properties: color" -> "validate union member: validate properties: missing required properties: fangs"
		{
			fixture:     "MethodResponses/inline-using-datatype-union/invalid-wrong-union.raml",
			msgContains: "missing required properties",
			line:        7, col: 3, endLine: 7, endCol: 11,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"TestType2\" not found"
		{
			fixture:     "MethodResponses/not-used-type/invalid-not-defined-type-used.raml",
			msgContains: "reference \"TestType2\" not found",
			line:        10, col: 3, endLine: 10, endCol: 12,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"TestType2\" not found"
		{
			fixture:     "MethodResponses/response-body-type/invalid-reference-not-defined-type.raml",
			msgContains: "reference \"TestType2\" not found",
			line:        15, col: 11, endLine: 15, endCol: 27,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "status code must be a 3-digit number"
		{
			fixture:     "MethodResponses/response-code/invalid.raml",
			msgContains: "status code must be a 3-digit number",
			line:        6, col: 7, endLine: 6, endCol: 10,
		},
		// error chain: "parse api" -> "decode api" -> "parse schemas (types)" -> "unmarshal types: make shape" -> "make shape type" -> "make json shape" -> "unmarshal json schema doc: invalid character '$' looking for beginning of object key string"
		{
			fixture:     "MethodResponses/root-schemas/invalid-json.raml",
			msgContains: "invalid character '$' looking for beginning of object key string",
			line:        4, col: 4, endLine: 4, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "unknown field"
		{
			fixture:     "Methods/available-methods/invalid-unknown-method.raml",
			msgContains: "unknown field",
			line:        11, col: 3, endLine: 11, endCol: 6,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unmarshal headers" -> "headers must be a mapping node"
		{
			fixture:     "Methods/custom-request-header/invalid-headers-node-type.raml",
			msgContains: "headers must be a mapping node",
			line:        8, col: 14, endLine: 8, endCol: 17,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "unmarshal headers" -> "headers must be a mapping node"
		{
			fixture:     "Methods/custom-response-header/invalid-headers-node-type.raml",
			msgContains: "headers must be a mapping node",
			line:        21, col: 18, endLine: 21, endCol: 21,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "decode" -> "decode value node" -> "decode example" -> "parse named example" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Methods\\include-example-raml\\exa1d212dmple.raml: The system cannot find the file specified."
		{
			fixture:     "Methods/include-example-raml/invalid-inexisting-file.raml",
			msgContains: "load resource",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unknown protocol"
		{
			fixture:     "Methods/protocols-array/invalid-element.raml",
			msgContains: "unknown protocol",
			line:        5, col: 23, endLine: 5, endCol: 26,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got string, expected bool"
		{
			fixture:     "Methods/query-params-boolean/invalid-example-type.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        11, col: 18, endLine: 11, endCol: 29,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "queryParameters and queryString are mutually exclusive"
		{
			fixture:     "Methods/query-params-enum/invalid-along-with-qs.raml",
			msgContains: "queryParameters and queryString are mutually exclusive",
			line:        7, col: 5, endLine: 7, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got []interface {}, expected a numeric type or json.Number"
		{
			fixture:     "Methods/query-params-number-01/invalid-example-type.raml",
			msgContains: "invalid type, got []interface {}, expected a numeric type or json.Number",
			line:        9, col: 18, endLine: 9, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: value must be one of (FOO, BAR, BAZ)"
		{
			fixture:     "Methods/query-params-ref-named-enum/invalid-example-type.raml",
			msgContains: "value must be one of (FOO, BAR, BAZ)",
			line:        15, col: 26, endLine: 15, endCol: 35,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unknown field"
		{
			fixture:     "Methods/querystring-queryparams/invalid-mutual-exclusive.raml",
			msgContains: "unknown field",
			line:        7, col: 5, endLine: 7, endCol: 19,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "explicit media type is required"
		{
			fixture:     "Methods/request-body-01/invalid-missing-root-media-type.raml",
			msgContains: "explicit media type is required",
			line:        16, col: 5, endLine: 16, endCol: 9,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"Admin\" not found"
		{
			fixture:     "Methods/request-body-02/invalid-inexisting-type.raml",
			msgContains: "reference \"Admin\" not found",
			line:        12, col: 7, endLine: 12, endCol: 23,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make concrete shape" -> "unmarshal yaml nodes" -> "unmarshal object facet: C:\\Sources\\go-raml\\raml-tck\\Methods\\request-body-03\\invalid-structure.raml:8:21" -> "properties must be a mapping"
		{
			fixture:     "Methods/request-body-03/invalid-structure.raml",
			msgContains: "properties must be a mapping",
			line:        8, col: 21, endLine: 8, endCol: 24,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "decode" -> "`type` and `schema` are mutually exclusive"
		{
			fixture:     "Methods/typed-request-body/invalid-type-with-schema.raml",
			msgContains: "`type` and `schema` are mutually exclusive",
			line:        19, col: 9, endLine: 19, endCol: 13,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Methods/typed-resp-and-req-body/invalid-type-with-scheme.raml",
			msgContains: "unknown facet",
			line:        15, col: 9, endLine: 15, endCol: 15,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "decode" -> "`type` and `schema` are mutually exclusive"
		{
			fixture:     "Methods/typed-response-body/invalid-scheme-and-type.raml",
			msgContains: "`type` and `schema` are mutually exclusive",
			line:        21, col: 13, endLine: 21, endCol: 17,
		},
		{
			fixture:     "Overlays/define-new-annotations/invalid-extends-inexisting-file.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/define-new-params/invalid-defineds-resource.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/define-new-schemas/invalid-defineds-trait.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/define-new-types/invalid-defines-resourcetype.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/double-displayname-override/invalid-add-trait-headers.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/double-overlay-with-lib/invalid-define-subresource.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/double-overlay/invalid-define-new-resource.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/empty-base/invalid-define-security-schemes.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/ext-override-deep-param/invalid-overide-body-content-type.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/extend-deep-param/invalid-resp-code.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/lib-extend-method/invalid-adds-protocols.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/overlay-with-metadata/invalid-defines-mediatype.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/override-deep-param/invalid-overrides-method.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/override-default/invalid.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/override-not-existing-method/invalid.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/override-not-existing-nested-resource/invalid.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/override-version/invalid.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		{
			fixture:     "Overlays/with-lib/invalid-inexisting-lib.raml",
			msgContains: "#%RAML 1.0 Overlay",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "collect variables index" -> "find variable" -> "parse template variables" -> "parse variable content" -> "unknown action"
		{
			fixture:     "ResourceTypes/chaining-functions/invalid-inexisting-func.raml",
			msgContains: "unknown action",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromType1"
		{
			fixture:     "ResourceTypes/datatype-properties-01/invalid-unknown-property.raml",
			msgContains: "propertyFromType1",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType2"
		{
			fixture:     "ResourceTypes/datatype-properties-02/invalid-unknown-property.raml",
			msgContains: "propertyFromResourceType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromType2"
		{
			fixture:     "ResourceTypes/datatype-properties-03/invalid-req-property-missing.raml",
			msgContains: "propertyFromType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType2"
		{
			fixture:     "ResourceTypes/datatype-properties-04/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: extraProperty, propertyFromResourceType1, propertyFromType1"
		{
			fixture:     "ResourceTypes/datatype-properties-05/invalid-req-property-missing.raml",
			msgContains: "extraProperty, propertyFromResourceType1, propertyFromType1",
			line:        52, col: 11, endLine: 52, endCol: 39,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: extraProperty, propertyFromResourceType1, propertyFromType1"
		{
			fixture:     "ResourceTypes/datatype-properties-06/invalid-req-property-missing.raml",
			msgContains: "extraProperty, propertyFromResourceType1, propertyFromType1",
			line:        51, col: 11, endLine: 51, endCol: 42,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: extraProperty, propertyFromResourceType1, propertyFromType1"
		{
			fixture:     "ResourceTypes/datatype-properties-07/invalid-req-property-missing.raml",
			msgContains: "extraProperty, propertyFromResourceType1, propertyFromType1",
			line:        51, col: 11, endLine: 51, endCol: 50,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: extraProperty, propertyFromType1"
		{
			fixture:     "ResourceTypes/datatype-properties-08/invalid-req-property-missing.raml",
			msgContains: "extraProperty, propertyFromType1",
			line:        50, col: 11, endLine: 51, endCol: 42,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType1"
		{
			fixture:     "ResourceTypes/datatype-properties-09/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType1",
			line:        45, col: 11, endLine: 47, endCol: 38,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "make parametrized endpoint" -> "decode parametrized endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "status code must be a 3-digit number"
		{
			fixture:     "ResourceTypes/datatype-properties-11/invalid-status-code.raml",
			msgContains: "status code must be a 3-digit number",
			line:        8, col: 9, endLine: 8, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "resource type method must be an HTTP method"
		{
			fixture:     "ResourceTypes/inherit-and-used/invalid-defines-resources.raml",
			msgContains: "resource type method must be an HTTP method",
			line:        23, col: 5, endLine: 23, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "resourceTypes must be a mapping node"
		{
			fixture:     "ResourceTypes/invalid-type/invalid.raml",
			msgContains: "resourceTypes must be a mapping node",
			line:        4, col: 3, endLine: 5, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "resource type method must be an HTTP method"
		{
			fixture:     "ResourceTypes/not-required-methods/invalid-not-supported-method.raml",
			msgContains: "resource type method must be an HTTP method",
			line:        5, col: 5, endLine: 5, endCol: 11,
		},
		// error chain: "apply resource types" -> "get resource type definition: reference \"Idontexist\" not found"
		{
			fixture:     "ResourceTypes/used-in-resource/invalid-inexisting-resourcetype.raml",
			msgContains: "reference \"Idontexist\" not found",
			line:        20, col: 9, endLine: 20, endCol: 19,
		},
		// error chain: "apply traits" -> "apply trait" -> "get trait definition: reference \"trait989898\" not found"
		{
			fixture:     "ResourceTypes/used-with-traits/invalid-not-defined-trait.raml",
			msgContains: "reference \"trait989898\" not found",
			line:        18, col: 12, endLine: 18, endCol: 23,
		},
		// error chain: "apply resource types" -> "compile resource type" -> "missing required parameter"
		{
			fixture:     "ResourceTypes/with-params/invalid-missing-param.raml",
			msgContains: "missing required parameter",
			line:        6, col: 4, endLine: 11, endCol: 116,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Resources/complex-description/invalid-structure.raml",
			msgContains: "unknown field in annotated scalar",
			line:        5, col: 5, endLine: 5, endCol: 8,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "unknown field"
		{
			fixture:     "Resources/description-only/invalid-not-supported-node.raml",
			msgContains: "unknown field",
			line:        5, col: 3, endLine: 5, endCol: 8,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "duplicated endpoint"
		{
			fixture:     "Resources/duplicate-uris/invalid-duplicate-uris.raml",
			msgContains: "duplicated endpoint",
			line:        12, col: 12, endLine: 12, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "duplicated endpoint"
		{
			fixture:     "Resources/nesting/invalid-share-same-uri.raml",
			msgContains: "duplicated endpoint",
			line:        19, col: 28, endLine: 19, endCol: 28,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromType1"
		{
			fixture:     "Resources/request-datatype-property/invalid-unknown-property.raml",
			msgContains: "propertyFromType1",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Resources/request-datatype/invalid-property-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        22, col: 15, endLine: 23, endCol: 23,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Resources/response-datatype/invalid-property-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        24, col: 15, endLine: 25, endCol: 23,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Resources/response-inline-type/invalid-property-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        20, col: 15, endLine: 21, endCol: 23,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: extraProperty"
		{
			fixture:     "Resources/restype-datatype-property-01/invalid-req-property-missing.raml",
			msgContains: "extraProperty",
			line:        44, col: 13, endLine: 45, endCol: 52,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromType1"
		{
			fixture:     "Resources/restype-datatype-property-02/invalid-req-property-missing.raml",
			msgContains: "propertyFromType1",
			line:        44, col: 13, endLine: 45, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType1"
		{
			fixture:     "Resources/restype-datatype-property-03/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType1",
			line:        44, col: 13, endLine: 45, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType1, propertyFromType1"
		{
			fixture:     "Resources/restype-datatype-property-04/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType1, propertyFromType1",
			line:        45, col: 13, endLine: 45, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType1"
		{
			fixture:     "Resources/restype-datatype-property-05/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType1",
			line:        22, col: 13, endLine: 24, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType2"
		{
			fixture:     "Resources/restype-datatype-property-06/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromType2"
		{
			fixture:     "Resources/restype-datatype-property-07/invalid-req-property-missing.raml",
			msgContains: "propertyFromType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: propertyFromResourceType2"
		{
			fixture:     "Resources/restype-datatype-property-08/invalid-req-property-missing.raml",
			msgContains: "propertyFromResourceType2",
			line:        42, col: 13, endLine: 44, endCol: 40,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "validate uri parameters" -> "uri parameter is not used"
		{
			fixture:     "Resources/uri-parameters-01/invalid-param-not-used.raml",
			msgContains: "uri parameter is not used",
			line:        8, col: 5, endLine: 8, endCol: 9,
		},
		// error chain: "build endpoints" -> "validate uri parameters" -> "unclosed '{'"
		// Position is the precise byte offset of the offending '{' (column =
		// uri start column + offset).
		{
			fixture:     "Resources/uri-parameters-02/invalid-unmatched-bracket.raml",
			msgContains: "unclosed '{'",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "status code must be a 3-digit number"
		{
			fixture:     "Responses/body-without-schema/invalid-resp-code.raml",
			msgContains: "status code must be a 3-digit number",
			line:        6, col: 7, endLine: 6, endCol: 13,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "duplicate status code"
		{
			fixture:     "Responses/code-without-body/invalid-duplicate-codes.raml",
			msgContains: "duplicate status code",
			line:        12, col: 7, endLine: 12, endCol: 10,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: username"
		{
			fixture:     "Responses/complex-body-type/invalid-missing-req-property.raml",
			msgContains: "username",
			line:        40, col: 15, endLine: 40, endCol: 27,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"TestType123123\" not found"
		{
			fixture:     "Responses/datatype-body-type/invalid-not-defined-type.raml",
			msgContains: "reference \"TestType123123\" not found",
			line:        15, col: 11, endLine: 15, endCol: 27,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make shape type" -> "make json shape" -> "unmarshal json schema doc: invalid character 'n' looking for beginning of object key string"
		{
			fixture:     "Responses/inline-json-schema/invalid-schema.raml",
			msgContains: "invalid character 'n' looking for beginning of object key string",
			line:        9, col: 11, endLine: 9, endCol: 27,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make responses" -> "make response" -> "decode response" -> "unmarshal headers" -> "headers must be a mapping node"
		{
			fixture:     "Responses/response-headers/invalid-headers-node-type.raml",
			msgContains: "headers must be a mapping node",
			line:        8, col: 18, endLine: 8, endCol: 21,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Root/baseuri-with-value/invalid.raml",
			msgContains: "unknown field in annotated scalar",
			line:        4, col: 3, endLine: 4, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "parse base uri" -> "unclosed '{'"
		{
			fixture:     "Root/baseuri/invalid-wrong-param.raml",
			msgContains: "unclosed '{'",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal base uri parameters" -> "make new shape yaml" -> "make shape type" -> "node kind must be scalar"
		{
			fixture:     "Root/baseuriparameters-01/invalid-val-sequence.raml",
			msgContains: "node kind must be scalar",
			line:        6, col: 7, endLine: 8, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Root/baseuriparameters-03/invalid-unknown-facet.raml",
			msgContains: "unknown facet",
			line:        8, col: 5, endLine: 8, endCol: 22,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"X\" not found"
		{
			fixture:     "Root/baseuriparameters-04/invalid-wrong-inherit.raml",
			msgContains: "reference \"X\" not found",
			line:        5, col: 3, endLine: 5, endCol: 4,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Root/baseuriparameters-05/invalid-example-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        9, col: 14, endLine: 9, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Root/baseuriparameters-06/invalid-unknown-node.raml",
			msgContains: "unknown facet",
			line:        9, col: 5, endLine: 9, endCol: 8,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal base uri parameters" -> "make new shape yaml" -> "make shape type" -> "mapping node is not allowed"
		{
			fixture:     "Root/baseuriparameters-07/invalid-type-structure.raml",
			msgContains: "mapping node is not allowed",
			line:        9, col: 7, endLine: 9, endCol: 17,
		},
		{
			fixture:     "Root/decode-error/invalid-decode.raml",
			msgContains: "types:",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "decode documentation item" -> "title must not be empty"
		{
			fixture:     "Root/documentation/invalid-empty-content-and-title.raml",
			msgContains: "title must not be empty",
			line:        4, col: 4, endLine: 4, endCol: 9,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "decode documentation item" -> "content must not be empty"
		{
			fixture:     "Root/documentation/invalid-empty-content.raml",
			msgContains: "content must not be empty",
			line:        5, col: 4, endLine: 5, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "decode documentation item" -> "title must not be empty"
		{
			fixture:     "Root/documentation/invalid-empty-title.raml",
			msgContains: "title must not be empty",
			line:        4, col: 4, endLine: 4, endCol: 9,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "decode documentation item" -> "content is required"
		{
			fixture:     "Root/documentation/invalid-no-content-node.raml",
			msgContains: "content is required",
			line:        4, col: 4, endLine: 4, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "documentation item must be a mapping node"
		{
			fixture:     "Root/documentation/invalid-no-items.raml",
			msgContains: "documentation item must be a mapping node",
			line:        3, col: 1, endLine: 3, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "unmarshal documentation item" -> "decode documentation item" -> "title is required"
		{
			fixture:     "Root/documentation/invalid-no-title-node.raml",
			msgContains: "title is required",
			line:        4, col: 4, endLine: 4, endCol: 57,
		},
		// error chain: "parse api" -> "decode api" -> "parse documentation items" -> "documentation item must be a mapping node"
		{
			fixture:     "Root/documentation/invalid-wrong-format.raml",
			msgContains: "documentation item must be a mapping node",
			line:        3, col: 1, endLine: 3, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "title is required"
		{
			fixture:     "Root/empty-01/invalid-empty.raml",
			msgContains: "title is required",
			line:        0, col: 0, endLine: 0, endCol: 0,
		},
		// error chain: "parse api" -> "decode api" -> "title is required"
		{
			fixture:     "Root/empty-02/invalid-empty-newline.raml",
			msgContains: "title is required",
			line:        0, col: 0, endLine: 0, endCol: 0,
		},
		// error chain: "parse api" -> "decode api" -> "title is required"
		{
			fixture:     "Root/empty-03/invalid-empty-2newline.raml",
			msgContains: "title is required",
			line:        0, col: 0, endLine: 0, endCol: 0,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve include" -> "include: open C:\\Sources\\go-raml\\raml-tck\\Root\\include-01\\relative.md: The system cannot find the file specified."
		{
			fixture:     "Root/include-01/invalid-missing-include.raml",
			msgContains: "file does not exist",
			line:        2, col: 8, endLine: 2, endCol: 28,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "parse resource type fragment" -> "load resource: http get https://fefwdflf3232f23f.fsd: status 404"
		// The !include node is at line 5, col 6 in the fixture.
		// We assert on "parse resource type fragment" because that is the stacktrace
		// node that carries the include position; the leaf "load resource" node has no position.
		{
			fixture:     "Root/include-02/invalid-https.raml",
			msgContains: "parse resource type fragment",
			line:        5, col: 6, endLine: 5, endCol: 43,
			extraOpts: []ParseOpt{OptWithHTTPClient(&http.Client{Transport: failingRoundTripper{}})},
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal media type" -> "media type must be a string"
		{
			fixture:     "Root/mediatype-01/invalid-missing-value.raml",
			msgContains: "media type must be a string",
			line:        7, col: 1, endLine: 7, endCol: 10,
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal media type" -> "media type: invalid media type someStringvalue"
		{
			fixture:     "Root/mediatype-02/invalid-not-supported.raml",
			msgContains: "invalid media type someStringvalue",
			line:        3, col: 12, endLine: 3, endCol: 27,
		},
		// error chain: "parse api" -> "decode api" -> "unknown field"
		{
			fixture:     "Root/other-01/invalid-unknown-node.raml",
			msgContains: "unknown field",
			line:        4, col: 1, endLine: 4, endCol: 18,
		},
		// error chain: "parse api" -> "decode api" -> "unknown field"
		{
			fixture:     "Root/other-02/invalid-unknown-node.raml",
			msgContains: "unknown field",
			line:        4, col: 1, endLine: 4, endCol: 5,
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal protocols" -> "protocols must not be empty"
		{
			fixture:     "Root/protocols/invalid-empty-array.raml",
			msgContains: "protocols must not be empty",
			line:        4, col: 12, endLine: 4, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal protocols" -> "protocols must be an array"
		{
			fixture:     "Root/protocols/invalid-not-array.raml",
			msgContains: "protocols must be an array",
			line:        4, col: 12, endLine: 4, endCol: 16,
		},
		// error chain: "parse api" -> "decode api" -> "preprocess" -> "unmarshal protocols" -> "unknown protocol"
		{
			fixture:     "Root/protocols/invalid-unknown-protocol.raml",
			msgContains: "unknown protocol",
			line:        5, col: 5, endLine: 5, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "title is required"
		{
			fixture:     "Root/title-01/invalid-missing.raml",
			msgContains: "title is required",
			line:        2, col: 1, endLine: 6, endCol: 28,
		},
		{
			fixture:     "Root/title-01/invalid-no-raml-version-whitespace.raml",
			msgContains: "#%RAML1.0",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve value node" -> "expected scalar or mapping node"
		{
			fixture:     "Root/title-02/invalid-not-string.raml",
			msgContains: "expected scalar or mapping node",
			line:        2, col: 8, endLine: 2, endCol: 46,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Root/title-03/invalid-not-string.raml",
			msgContains: "unknown field in annotated scalar",
			line:        2, col: 10, endLine: 2, endCol: 16,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve include" -> "include: open C:\\Sources\\go-raml\\raml-tck\\Root\\title-04\\adsrelative.md: The system cannot find the file specified."
		{
			fixture:     "Root/title-04/invalid-included.raml",
			msgContains: "file does not exist",
			line:        2, col: 8, endLine: 2, endCol: 31,
		},
		// error chain: "parse api" -> "decode api" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Root/version/invalid-version-structure.raml",
			msgContains: "unknown field in annotated scalar",
			line:        5, col: 3, endLine: 5, endCol: 8,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "SecuritySchemes/basic-authentication/invalid-unknown-type.raml",
			msgContains: "unknown security scheme type",
			line:        9, col: 11, endLine: 9, endCol: 30,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "SecuritySchemes/custom-scheme-prefix/invalid-prefix.raml",
			msgContains: "unknown security scheme type",
			line:        7, col: 11, endLine: 7, endCol: 16,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "SecuritySchemes/digest-authentication/invalid-unknown-type.raml",
			msgContains: "unknown security scheme type",
			line:        9, col: 11, endLine: 9, endCol: 30,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown signature"
		{
			fixture:     "SecuritySchemes/oauth1/invalid-not-supported-signature.raml",
			msgContains: "unknown signature",
			line:        14, col: 21, endLine: 14, endCol: 23,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "requestTokenUri is required for OAuth 1.0"
		{
			fixture:     "SecuritySchemes/oauth1/invalid-req-property-missing.raml",
			msgContains: "requestTokenUri is required for OAuth 1.0",
			line:        9, col: 11, endLine: 9, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme description" -> "decode security scheme description" -> "unknown key in describedBy"
		{
			fixture:     "SecuritySchemes/oauth2-01/invalid-unknown-node.raml",
			msgContains: "unknown key in describedBy",
			line:        10, col: 7, endLine: 10, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "accessTokenUri is required for OAuth 2.0"
		{
			fixture:     "SecuritySchemes/oauth2-02/invalid-req-property-missing.raml",
			msgContains: "accessTokenUri is required for OAuth 2.0",
			line:        9, col: 11, endLine: 9, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown authorization grant"
		{
			fixture:     "SecuritySchemes/oauth2-03/invalid-property-value.raml",
			msgContains: "unknown authorization grant",
			line:        13, col: 50, endLine: 13, endCol: 63,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "SecuritySchemes/oauth2-used/invalid-unknown-type.raml",
			msgContains: "unknown security scheme type",
			line:        7, col: 11, endLine: 7, endCol: 16,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal security scheme definitions" -> "make security scheme definition" -> "decode security scheme definition" -> "make security scheme settings" -> "unknown security scheme type"
		{
			fixture:     "SecuritySchemes/pass-through/invalid-unknown-type.raml",
			msgContains: "unknown security scheme type",
			line:        9, col: 11, endLine: 9, endCol: 20,
		},
		// error chain: "apply security schemes" -> "apply security scheme" -> "validate security scope overrides" -> "scope is not declared by the security scheme"
		{
			fixture:     "SecuritySchemes/scopes/invalid-scope.raml",
			msgContains: "scope is not declared by the security scheme",
			line:        17, col: 23, endLine: 17, endCol: 51,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/lowercamelcase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 64,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/lowercase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 59,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/lowerhyphencase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 65,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/lowerunderscorecase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 69,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal trait definitions" -> "make traits" -> "collect variables index" -> "find variable" -> "parse template variables" -> "parse variable content" -> "unknown action"
		{
			fixture:     "TemplateFunctions/multiple/invalid-used-without-pipe.raml",
			msgContains: "unknown action",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/pluralize/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 59,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/singularize/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 61,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/uppercamelcase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 64,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/uppercase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 59,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/upperhyphencase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 65,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "TemplateFunctions/upperunderscorecase/invalid-used-without-pipe.raml",
			msgContains: "unexpected parameter",
			line:        7, col: 5, endLine: 9, endCol: 69,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate array item $[1]: must match pattern ^\\d+\\-\\w+$"
		{
			fixture:     "Traits/applied-to-method/invalid-pattern.raml",
			msgContains: "must match pattern ^\\d+\\-\\w+$",
			line:        14, col: 13, endLine: 15, endCol: 26,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: rtTypeProperty1"
		{
			fixture:     "Traits/datatype-properties-01/invalid-req-property-missing.raml",
			msgContains: "rtTypeProperty1",
			line:        62, col: 11, endLine: 62, endCol: 39,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: rtTypeProperty2"
		{
			fixture:     "Traits/datatype-properties-02/invalid-req-property-missing.raml",
			msgContains: "rtTypeProperty2",
			line:        62, col: 11, endLine: 62, endCol: 39,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.traitProperty1: invalid type, got int, expected bool"
		{
			fixture:     "Traits/datatype-properties-03/invalid-wrong-value-type.raml",
			msgContains: "invalid type, got int, expected bool",
			line:        69, col: 15, endLine: 72, endCol: 45,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.traitProperty1: invalid type, got bool, expected a numeric type or json.Number"
		{
			fixture:     "Traits/datatype-properties-04/invalid-wrong-value-type.raml",
			msgContains: "invalid type, got bool, expected a numeric type or json.Number",
			line:        69, col: 15, endLine: 72, endCol: 45,
		},
		// error chain: "validate shapes" -> "check type" -> "enum value is invalid: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Traits/merge-array-values/invalid-types-conflict.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        19, col: 13, endLine: 19, endCol: 16,
		},
		// error chain: "apply traits" -> "apply trait" -> "compile trait" -> "unexpected parameter"
		{
			fixture:     "Traits/params-collision-resolution/invalid-unknown-param.raml",
			msgContains: "unexpected parameter",
			line:        10, col: 5, endLine: 12, endCol: 54,
		},
		// error chain: "apply traits" -> "apply trait" -> "get trait definition: reference \"idontexist\" not found"
		{
			fixture:     "Traits/with-params/invalid-inexisting-trait.raml",
			msgContains: "reference \"idontexist\" not found",
			line:        13, col: 10, endLine: 13, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "parse data" -> "check fragment kind: identify fragment: unknown fragment kind: head: Raw text."
		{
			fixture:     "Types/External Types/include-txt/invalid-unknown-type.raml",
			msgContains: "Raw text.",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "parse data" -> "make json data type" -> "decode fragment" -> "parse types: make shape" -> "make json shape" -> "unmarshal json schema doc: invalid character 'p' looking for beginning of object key string"
		{
			fixture:     "Types/External Types/include-type-json-01/invalid-included-json.raml",
			msgContains: "invalid character 'p' looking for beginning of object key string",
			line:        1, col: 0, endLine: 0, endCol: 17,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "resolve link" -> "make concrete shape" -> "unmarshal yaml nodes" -> "type-specific are not allowed for JSON external types"
		{
			fixture:     "Types/External Types/include-type-json-02/invalid-add-more-properties.raml",
			msgContains: "type-specific are not allowed for JSON external types",
			line:        6, col: 5, endLine: 6, endCol: 15,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "parse data" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Types\\External Types\\include-type-json-02\\account.json: The system cannot find the file specified."
		{
			fixture:     "Types/External Types/include-type-json-02/invalid-use-in-other-types.raml",
			msgContains: "file does not exist",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unmarshal headers" -> "make property" -> "make shape" -> "make shape type" -> "parse data" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Types\\External Types\\include-type-json-02\\account.json: The system cannot find the file specified."
		{
			fixture:     "Types/External Types/include-type-json-02/invalid-used-in-headers.raml",
			msgContains: "file does not exist",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "unmarshal query parameters" -> "make property" -> "make shape" -> "make shape type" -> "parse data" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Types\\External Types\\include-type-json-02\\account.json: The system cannot find the file specified."
		{
			fixture:     "Types/External Types/include-type-json-02/invalid-used-in-queryParameters.raml",
			msgContains: "file does not exist",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "unmarshal uri parameters" -> "make new shape yaml" -> "make shape type" -> "parse data" -> "load resource: open C:\\Sources\\go-raml\\raml-tck\\Types\\External Types\\include-type-json-02\\account.json: The system cannot find the file specified."
		{
			fixture:     "Types/External Types/include-type-json-02/invalid-used-in-uriParameters.raml",
			msgContains: "file does not exist",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "failed to validate against JSON schema: jsonschema validation failed with 'file:///C:/Sources/go-raml/raml-tck/Types/External Types/json-schema-examples-01/invalid-examples.raml#'\n- at '': missing property 'id'"
		{
			fixture:     "Types/External Types/json-schema-examples-01/invalid-examples.raml",
			msgContains: "missing property 'id'",
			line:        19, col: 5, endLine: 21, endCol: 17,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate property $.c: C:\\Sources\\go-raml\\raml-tck\\Types\\External Types\\json-schema-examples-02\\invalid-external-prop-definition.raml:21:10" -> "failed to validate against JSON schema: jsonschema validation failed with 'file:///C:/Sources/go-raml/raml-tck/Types/External Types/json-schema-examples-02/invalid-external-prop-definition.raml#'\n- at '': missing property 'id'"
		{
			fixture:     "Types/External Types/json-schema-examples-02/invalid-external-prop-definition.raml",
			msgContains: "missing property 'id'",
			line:        21, col: 10, endLine: 21, endCol: 18,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate custom facet: invalid type, got int, expected string"
		{
			fixture:     "Types/Facets/inheritance-01/invalid-wrong-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        13, col: 15, endLine: 13, endCol: 19,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "explicit media type is required"
		{
			fixture:     "Types/Facets/inheritance-02/invalid-inherit-unknown-type.raml",
			msgContains: "explicit media type is required",
			line:        16, col: 5, endLine: 16, endCol: 9,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "required custom facet is missing"
		{
			fixture:     "Types/Facets/naming-constraints/invalid-ancestor-facet.raml",
			msgContains: "required custom facet is missing",
			line:        11, col: 3, endLine: 11, endCol: 14,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "required custom facet is missing"
		{
			fixture:     "Types/Facets/naming-constraints/invalid-matches-built-in.raml",
			msgContains: "required custom facet is missing",
			line:        9, col: 3, endLine: 9, endCol: 7,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "required custom facet is missing"
		{
			fixture:     "Types/Facets/naming-constraints/invalid-missing-required-facet.raml",
			msgContains: "required custom facet is missing",
			line:        9, col: 3, endLine: 9, endCol: 7,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "decode" -> "decode value node" -> "decode facets" -> "facet name must not begin with '('"
		{
			fixture:     "Types/Facets/naming-constraints/invalid-paren-in-name.raml",
			msgContains: "facet name must not begin with '('",
			line:        8, col: 7, endLine: 8, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "cannot redefine built-in type"
		{
			fixture:     "Types/Facets/redefine-built-in/invalid-redefine-datetime.raml",
			msgContains: "cannot redefine built-in type",
			line:        4, col: 3, endLine: 4, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "required custom facet is missing"
		{
			fixture:     "Types/Facets/simple-facet/invalid-wrong-facet-used.raml",
			msgContains: "required custom facet is missing",
			line:        14, col: 3, endLine: 14, endCol: 7,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Types/ObjectTypes/discriminator/invalid-inline-discriminator.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression" -> "make concrete shape yaml: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\discriminator\\invalid-union-type.raml:14:3" -> "unmarshal yaml nodes" -> "discriminator is not allowed on union types"
		{
			fixture:     "Types/ObjectTypes/discriminator/invalid-union-type.raml",
			msgContains: "discriminator is not allowed on union types",
			line:        16, col: 6, endLine: 16, endCol: 19,
		},
		// error chain: "validate shapes" -> "check type" -> "check properties: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\discriminator\\invalid-wrong-prop-pointed.raml:6:20" -> "discriminator property not found"
		{
			fixture:     "Types/ObjectTypes/discriminator/invalid-wrong-prop-pointed.raml",
			msgContains: "discriminator property not found",
			line:        6, col: 20, endLine: 6, endCol: 30,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: title??"
		{
			fixture:     "Types/ObjectTypes/double-trailing-question-mark-explicit-optional/invalid-explicitly-required.raml",
			msgContains: "title??",
			line:        13, col: 7, endLine: 13, endCol: 17,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: title??"
		{
			fixture:     "Types/ObjectTypes/double-trailing-question-mark-val-provided/invalid-missing-required-value.raml",
			msgContains: "title??",
			line:        13, col: 7, endLine: 13, endCol: 17,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: title??"
		{
			fixture:     "Types/ObjectTypes/double-trailing-question-mark/invalid-explicitly-required.raml",
			msgContains: "title??",
			line:        14, col: 7, endLine: 15, endCol: 24,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "maxLength" -> "decode: cannot unmarshal !!int `-14` into uint64"
		{
			fixture:     "Types/ObjectTypes/inherit-string/invalid-wrong-constraint.raml",
			msgContains: "cannot unmarshal !!int `-14` into uint64",
			line:        6, col: 16, endLine: 6, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: object must have not more than 4 properties"
		{
			fixture:     "Types/ObjectTypes/max-properties/invalid-max-violated.raml",
			msgContains: "object must have not more than 4 properties",
			line:        15, col: 7, endLine: 19, endCol: 26,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: object must have at least 2 properties"
		{
			fixture:     "Types/ObjectTypes/min-properties/invalid-min-violated.raml",
			msgContains: "object must have at least 2 properties",
			line:        11, col: 7, endLine: 11, endCol: 27,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"EmailAdmin\" not found" -> "resolve multiple inheritance" -> "resolve inherit"
		{
			fixture:     "Types/ObjectTypes/multiple-inheritance/invalid-inherit-inexisting-type.raml",
			msgContains: "reference \"EmailAdmin\" not found",
			line:        13, col: 21, endLine: 13, endCol: 31,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate object shape: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\not-required-with-default\\invalid-wrong-default-type.raml:10:7" -> "validate property" -> "validate default: invalid type, got int, expected string"
		{
			fixture:     "Types/ObjectTypes/not-required-with-default/invalid-wrong-default-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        12, col: 18, endLine: 12, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.post: invalid type, got int, expected map[string]any"
		{
			fixture:     "Types/ObjectTypes/pattern-property-and-explicit/invalid-expected-pattern-prevail.raml",
			msgContains: "invalid type, got int, expected map[string]any",
			line:        20, col: 7, endLine: 21, endCol: 13,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate pattern property $.foo: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\pattern-property-asterisk\\invalid-wrong-type.raml:7:7" -> "validate pattern property: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/ObjectTypes/pattern-property-asterisk/invalid-wrong-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        7, col: 7, endLine: 7, endCol: 11,
		},
		// error chain: "validate shapes" -> "check type" -> "check pattern properties: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\pattern-property-or\\invalid-no-additionalProperties.raml:16:27" -> "pattern properties are not allowed with \"additionalProperties: false\""
		{
			fixture:     "Types/ObjectTypes/pattern-property-or/invalid-no-additionalProperties.raml",
			msgContains: "false\"",
			line:        16, col: 27, endLine: 16, endCol: 32,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate pattern property $.put: C:\\Sources\\go-raml\\raml-tck\\Types\\ObjectTypes\\pattern-property-two\\invalid-wrong-type.raml:16:7" -> "validate pattern property: invalid type, got int, expected map[string]any"
		{
			fixture:     "Types/ObjectTypes/pattern-property-two/invalid-wrong-type.raml",
			msgContains: "invalid type, got int, expected map[string]any",
			line:        16, col: 7, endLine: 16, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Types/ObjectTypes/properties-property/invalid-wrong-parent-type.raml",
			msgContains: "unknown facet",
			line:        6, col: 5, endLine: 6, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: comment_id"
		{
			fixture:     "Types/ObjectTypes/required-property/invalid-missing.raml",
			msgContains: "comment_id",
			line:        17, col: 7, endLine: 19, endCol: 29,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: name"
		{
			fixture:     "Types/ObjectTypes/simple-inheritance/invalid-missing-required-prop.raml",
			msgContains: "missing required properties",
			line:        15, col: 7, endLine: 15, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.initial_comments: invalid type, got int, expected string"
		{
			fixture:     "Types/ObjectTypes/simple-type/invalid-wrong-value-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        10, col: 7, endLine: 11, endCol: 28,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: title?"
		{
			fixture:     "Types/ObjectTypes/single-trailing-question-mark/invalid-explicitly-required.raml",
			msgContains: "title?",
			line:        13, col: 7, endLine: 13, endCol: 17,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\PropertyOverride\\define-restrictions\\invalid-restrictions-conflict.raml:19:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "inherit property" -> "cannot inherit from different type"
		{
			fixture:     "Types/PropertyOverride/define-restrictions/invalid-restrictions-conflict.raml",
			msgContains: "cannot inherit from different type",
			line:        15, col: 7, endLine: 15, endCol: 11,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\PropertyOverride\\multiple-override\\invalid-make-property-not-required.raml:19:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "cannot make required property optional"
		{
			fixture:     "Types/PropertyOverride/multiple-override/invalid-make-property-not-required.raml",
			msgContains: "cannot make required property optional",
			line:        22, col: 7, endLine: 22, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "duplicate custom facet"
		{
			fixture:     "Types/PropertyOverride/override-facet/invalid-cannot-be-overriden.raml",
			msgContains: "duplicate custom facet",
			line:        6, col: 7, endLine: 6, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got <nil>, expected map[string]any"
		{
			fixture:     "Types/PropertyOverride/override-optional-property/invalid-blank-example.raml",
			msgContains: "invalid type, got <nil>, expected map[string]any",
			line:        19, col: 13, endLine: 19, endCol: 13,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\PropertyOverride\\override-string-with-type-01\\invalid-make-property-not-required.raml:11:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "cannot make required property optional"
		{
			fixture:     "Types/PropertyOverride/override-string-with-type-01/invalid-make-property-not-required.raml",
			msgContains: "cannot make required property optional",
			line:        14, col: 7, endLine: 14, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate object shape: C:\\Sources\\go-raml\\raml-tck\\Types\\PropertyOverride\\override-type-with-type-01\\invalid-violate-maxlength.raml:18:7" -> "validate property" -> "validate example: length must be less than 5"
		{
			fixture:     "Types/PropertyOverride/override-type-with-type-01/invalid-violate-maxlength.raml",
			msgContains: "length must be less than 5",
			line:        20, col: 18, endLine: 20, endCol: 39,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: visit array items: get referenced shape: reference \"Admin\" not found"
		{
			fixture:     "Types/Type Expressions/inherit-datatype-array/invalid-inherit-inexisting-type.raml",
			msgContains: "reference \"Admin\" not found",
			line:        6, col: 3, endLine: 6, endCol: 10,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "antlr error" -> "token recognition error at: '[ '" -> "token recognition error at: ','" -> "token recognition error at: ']'" -> "extraneous input 'integer' expecting <EOF>"
		{
			fixture:     "Types/Type Expressions/inherit-datatype-scalar-union/invalid-inherit-two-scalars.raml",
			msgContains: "extraneous input 'integer' expecting <EOF>",
			line:        6, col: 3, endLine: 0, endCol: 0,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: visit array items: visit children: visit children: get referenced shape: reference \"Admin\" not found"
		{
			fixture:     "Types/Type Expressions/inherit-datatype-union-array-01/invalid-use-inexisting-type.raml",
			msgContains: "reference \"Admin\" not found",
			line:        22, col: 7, endLine: 22, endCol: 14,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: visit array items: visit children: visit children: get referenced shape: reference \"Wall\" not found"
		{
			fixture:     "Types/Type Expressions/inherit-datatype-union-array-02/invalid-inherit-inexisting-type.raml",
			msgContains: "reference \"Wall\" not found",
			line:        18, col: 3, endLine: 18, endCol: 10,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"Admin\" not found"
		{
			fixture:     "Types/Type Expressions/inherit-datatype/invalid-inherit-inexisting-datatype.raml",
			msgContains: "reference \"Admin\" not found",
			line:        6, col: 3, endLine: 6, endCol: 11,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "antlr error" -> "token recognition error at: '[['" -> "token recognition error at: ']'"
		{
			fixture:     "Types/Type Expressions/inherit-scalar-nested-array/invalid-nesting-syntax.raml",
			msgContains: "token recognition error at",
			line:        4, col: 3, endLine: 0, endCol: 0,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "unmarshal object facet: C:\\Sources\\go-raml\\raml-tck\\Types\\additional-properties\\invalid-property-value.raml:8:7" -> "make scalar node" -> "resolve value node" -> "unknown field in annotated scalar"
		{
			fixture:     "Types/additional-properties/invalid-property-value.raml",
			msgContains: "unknown field in annotated scalar",
			line:        8, col: 7, endLine: 8, endCol: 11,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.pp: object must have at least 3 properties"
		{
			fixture:     "Types/annotation-inherits-pattern-prop-01/invalid-minproperties-violated.raml",
			msgContains: "object must have at least 3 properties",
			line:        13, col: 7, endLine: 15, endCol: 15,
		},
		// error chain: "validate shapes" -> "check domain extension: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/annotations-used-in-type-01/invalid-wrong-value-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        20, col: 27, endLine: 20, endCol: 32,
		},
		// error chain: "resolve domain extensions" -> "resolve domain extension" -> "get referenced shape: reference \"UndefinedAnnotation\" not found"
		{
			fixture:     "Types/annotations-used-in-type-02/invalid-undefined-annotation.raml",
			msgContains: "reference \"UndefinedAnnotation\" not found",
			line:        21, col: 9, endLine: 21, endCol: 30,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.prop1: invalid type, got int, expected string"
		{
			fixture:     "Types/annotations-used-in-type-03/invalid-wrong-nested-property-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        22, col: 31, endLine: 22, endCol: 41,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate property $.unionArray" -> "validate array item $.unionArray[0]: C:\\Sources\\go-raml\\raml-tck\\Types\\array-of-datatype-unions-01\\invalid-example-property.raml:21:7" -> "value does not match any type" -> "validate union member: validate properties: missing required properties: property2" -> "validate union member: validate properties: missing required properties: property3"
		{
			fixture:     "Types/array-of-datatype-unions-01/invalid-example-property.raml",
			msgContains: "property3",
			line:        15, col: 3, endLine: 15, endCol: 14,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate property $.unionArray" -> "validate array item $.unionArray[1]: C:\\Sources\\go-raml\\raml-tck\\Types\\array-of-datatype-unions-02\\invalid-example-property.raml:21:7" -> "value does not match any type" -> "validate union member: validate properties: validate property $.unionArray[1].property2: invalid type, got int, expected string" -> "validate union member: validate properties: validate property $.unionArray[1].property1: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/array-of-datatype-unions-02/invalid-example-property.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        15, col: 3, endLine: 15, endCol: 14,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate array item $[1]: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/array-property/invalid-string-in-number-array.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        8, col: 7, endLine: 9, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got int, expected map[string]any"
		{
			fixture:     "Types/complex-example-01/invalid-wrong-structure.raml",
			msgContains: "invalid type, got int, expected map[string]any",
			line:        34, col: 14, endLine: 34, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got int, expected map[string]any"
		{
			fixture:     "Types/complex-example-02/invalid-wrong-structure.raml",
			msgContains: "invalid type, got int, expected map[string]any",
			line:        52, col: 14, endLine: 52, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.items: validate array item $.items[0]: validate properties: validate property $.items[0].r1: invalid type, got int, expected string" -> "validate object shape: C:\\Sources\\go-raml\\raml-tck\\Types\\complex-used-in-annotations-01\\invalid-multiple-errors.raml:27:7" -> "validate property"
		{
			fixture:     "Types/complex-used-in-annotations-01/invalid-multiple-errors.raml",
			msgContains: "invalid type, got int, expected string",
			line:        12, col: 9, endLine: 17, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "items must be either a reference or an inline type"
		{
			fixture:     "Types/datatypes-array-01/invalid.raml",
			msgContains: "items must be either a reference or an inline type",
			line:        19, col: 12, endLine: 19, endCol: 22,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.users: validate array item $.users[0]: validate properties: validate property $.users[0].age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/datatypes-array-02/invalid-wrong-example-types.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        20, col: 7, endLine: 28, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate property $.union1: C:\\Sources\\go-raml\\raml-tck\\Types\\datatypes-union-01\\invalid-example-property.raml:21:7" -> "value does not match any type" -> "validate union member: validate properties: missing required properties: property1" -> "validate union member: validate properties: missing required properties: property4"
		{
			fixture:     "Types/datatypes-union-01/invalid-example-property.raml",
			msgContains: "property4",
			line:        15, col: 3, endLine: 15, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make shape type" -> "make json shape" -> "unmarshal json schema doc: unexpected EOF"
		{
			fixture:     "Types/defined-with-jsonschema/invalid-json-schema.raml",
			msgContains: "unexpected EOF",
			line:        5, col: 3, endLine: 5, endCol: 9,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Types/determine-default-types/invalid-unknown-property.raml",
			msgContains: "unknown facet",
			line:        7, col: 5, endLine: 7, endCol: 10,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"asdasd\" not found"
		{
			fixture:     "Types/implicitly-defined-type/invalid-inexisting-base-type.raml",
			msgContains: "reference \"asdasd\" not found",
			line:        5, col: 3, endLine: 5, endCol: 6,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: length must be greater than 5" -> "validate example: length must be less than 3"
		{
			fixture:     "Types/inherit-and-extend-constraints-01/invalid-minmaxlength-violated.raml",
			msgContains: "length must be less than 3",
			line:        9, col: 14, endLine: 9, endCol: 19,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\inherit-and-extend-constraints-02\\invalid-lesser-constraints.raml:6:3" -> "unwrap shape" -> "merge shapes" -> "minLength constraint violation"
		{
			fixture:     "Types/inherit-and-extend-constraints-02/invalid-lesser-constraints.raml",
			msgContains: "minLength constraint violation",
			line:        8, col: 16, endLine: 8, endCol: 17,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\inherit-and-extend-constraints-03\\invalid-make-non-required.raml:7:3" -> "unwrap shape" -> "merge shapes" -> "inherit properties" -> "cannot make required property optional"
		{
			fixture:     "Types/inherit-and-extend-constraints-03/invalid-make-non-required.raml",
			msgContains: "cannot make required property optional",
			line:        10, col: 7, endLine: 10, endCol: 12,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate default: invalid type, got string, expected bool"
		{
			fixture:     "Types/inherit-boolean/invalid-default-value.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        7, col: 18, endLine: 7, endCol: 21,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got int, expected string"
		{
			fixture:     "Types/inherit-datetime/invalid-date-only-example.raml",
			msgContains: "invalid type, got int, expected string",
			line:        7, col: 14, endLine: 7, endCol: 20,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "invalid format"
		{
			fixture:     "Types/inherit-datetime/invalid-datetime-format.raml",
			msgContains: "invalid format",
			line:        7, col: 13, endLine: 7, endCol: 22,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got int, expected string"
		{
			fixture:     "Types/inherit-datetime/invalid-datetime-only-example.raml",
			msgContains: "invalid type, got int, expected string",
			line:        7, col: 14, endLine: 7, endCol: 20,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Types/inherit-datetime/invalid-time-only-example.raml",
			msgContains: "unknown facet",
			line:        7, col: 5, endLine: 7, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "unknown facet"
		{
			fixture:     "Types/inherit-datetime/invalid-time-only-format.raml",
			msgContains: "unknown facet",
			line:        7, col: 5, endLine: 7, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "maxLength" -> "decode: cannot unmarshal !!int `-1` into uint64"
		{
			fixture:     "Types/inherit-file/invalid-length.raml",
			msgContains: "cannot unmarshal !!int `-1` into uint64",
			line:        8, col: 16, endLine: 8, endCol: 18,
		},
		// error chain: "validate shapes" -> "check type" -> "minimum must be less than or equal to maximum"
		{
			fixture:     "Types/inherit-integer-min-max/invalid-conflict-minmax.raml",
			msgContains: "minimum must be less than or equal to maximum",
			line:        5, col: 5, endLine: 5, endCol: 13,
		},
		// error chain: "validate shapes" -> "check type" -> "minimum must be less than or equal to maximum"
		{
			fixture:     "Types/inherit-number-min-max/invalid-conflict.raml",
			msgContains: "minimum must be less than or equal to maximum",
			line:        5, col: 5, endLine: 5, endCol: 13,
		},
		// error chain: "validate shapes" -> "check type" -> "invalid format"
		{
			fixture:     "Types/inherit-number-min-max/invalid-wrong-format.raml",
			msgContains: "invalid format",
			line:        7, col: 17, endLine: 7, endCol: 26,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/inherit-number-with-decimals/invalid-wrong-decimal-point.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        5, col: 16, endLine: 5, endCol: 27,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: object must have at least 2 properties"
		{
			fixture:     "Types/inherit-pattern-property-01/invalid-minproperties-violated.raml",
			msgContains: "object must have at least 2 properties",
			line:        14, col: 7, endLine: 14, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: object must have not more than 1 properties"
		{
			fixture:     "Types/inherit-pattern-property-02/invalid-max-properties-violated.raml",
			msgContains: "object must have not more than 1 properties",
			line:        9, col: 7, endLine: 10, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "minLength" -> "decode: cannot unmarshal !!int `-2` into uint64"
		{
			fixture:     "Types/inherit-string-min-max/invalid-minmax-values.raml",
			msgContains: "cannot unmarshal !!int `-2` into uint64",
			line:        7, col: 20, endLine: 7, endCol: 22,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: missing required properties: name"
		{
			fixture:     "Types/inheritance-01/invalid-wrong-type-missing-req.raml",
			msgContains: "missing required properties",
			line:        17, col: 7, endLine: 18, endCol: 23,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.someProperty: validate properties: validate property $.someProperty.age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/inheritance-02/invalid-unknown-prop.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        23, col: 7, endLine: 25, endCol: 18,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"hello\" not found"
		{
			fixture:     "Types/inheritance-03/invalid-unknown-parent-type.raml",
			msgContains: "reference \"hello\" not found",
			line:        5, col: 3, endLine: 5, endCol: 11,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: get referenced shape: reference \"fsdf\" not found"
		{
			fixture:     "Types/inline-baseuriparameters/invalid-type-declaration.raml",
			msgContains: "reference \"fsdf\" not found",
			line:        6, col: 3, endLine: 6, endCol: 10,
		},
		// error chain: "parse api" -> "decode api: mapping values are not allowed in this context"
		{
			fixture:     "Types/inline-query-string/invalid-type-declaration.raml",
			msgContains: "mapping values are not allowed in this context",
			line:        8, col: 1, endLine: 0, endCol: 0,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate object shape: C:\\Sources\\go-raml\\raml-tck\\Types\\inline-request-body\\invalid-type-declaration.raml:14:11" -> "validate property" -> "unknown facet"
		{
			fixture:     "Types/inline-request-body/invalid-type-declaration.raml",
			msgContains: "unknown facet",
			line:        16, col: 13, endLine: 16, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate items" -> "validate example: must match pattern ^\\d+\\-\\w+$"
		{
			fixture:     "Types/inline-request-headers/invalid-type-declaration.raml",
			msgContains: "must match pattern ^\\d+\\-\\w+$",
			line:        12, col: 20, endLine: 12, endCol: 33,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate object shape: C:\\Sources\\go-raml\\raml-tck\\Types\\inline-response-body\\invalid-type-declaration.raml:16:15" -> "validate property" -> "validate example: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/inline-response-body/invalid-type-declaration.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        18, col: 26, endLine: 18, endCol: 29,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate items" -> "validate example: must match pattern ^\\d+\\-\\w+$"
		{
			fixture:     "Types/inline-response-headers/invalid-type-declaration.raml",
			msgContains: "must match pattern ^\\d+\\-\\w+$",
			line:        14, col: 24, endLine: 14, endCol: 33,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "unmarshal uri parameters" -> "make new shape yaml" -> "make shape type" -> "mapping node is not allowed"
		{
			fixture:     "Types/inline-uri-parameters/invalid-type-declaration.raml",
			msgContains: "mapping node is not allowed",
			line:        11, col: 11, endLine: 14, endCol: 25,
		},
		// error chain: "parse api" -> "parse uses library" -> "check fragment kind: unexpected fragment frag != kind: API != Library"
		{
			fixture:     "Types/lib-trait-with-param/invalid-missing-lib-tag.raml",
			msgContains: "API != Library",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.name: invalid type, got <nil>, expected string"
		{
			fixture:     "Types/lib-with-included-json-01/invalid-required-val-missing.raml",
			msgContains: "invalid type, got <nil>, expected string",
			line:        9, col: 14, endLine: 9, endCol: 37,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.data: invalid type, got <nil>, expected map[string]any"
		{
			fixture:     "Types/lib-with-included-json-02/invalid-missing-req-property.raml",
			msgContains: "invalid type, got <nil>, expected map[string]any",
			line:        11, col: 14, endLine: 11, endCol: 37,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.comment: invalid type, got <nil>, expected string"
		{
			fixture:     "Types/lib-with-simple-type-01/invalid-requirement-violated.raml",
			msgContains: "invalid type, got <nil>, expected string",
			line:        10, col: 7, endLine: 11, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.size: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/lib-with-simple-type-02/invalid-wrong-value-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        10, col: 7, endLine: 11, endCol: 15,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: invalid type, got string, expected bool"
		{
			fixture:     "Types/lib-with-simple-type-03/invalid-wrong-example-type.raml",
			msgContains: "invalid type, got string, expected bool",
			line:        6, col: 14, endLine: 6, endCol: 20,
		},
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\multiple-inheritance\\invalid-incompatible-types.raml:10:3" -> "unwrap shape" -> "unwrap parents" -> "multiple parents unwrap" -> "merge shapes" -> "cannot inherit from different type"
		{
			fixture:     "Types/multiple-inheritance/invalid-incompatible-types.raml",
			msgContains: "cannot inherit from different type",
			line:        11, col: 12, endLine: 11, endCol: 18,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression" -> "resolve: C:\\Sources\\go-raml\\raml-tck\\Types\\multiple-recurrent-definitions-01\\invalid.raml:11:3" -> "resolve: C:\\Sources\\go-raml\\raml-tck\\Types\\multiple-recurrent-definitions-01\\invalid.raml:8:3" -> "resolve: C:\\Sources\\go-raml\\raml-tck\\Types\\multiple-recurrent-definitions-01\\invalid.raml:5:3" -> "cyclic type reference"
		{
			fixture:     "Types/multiple-recurrent-definitions-01/invalid.raml",
			msgContains: "cyclic type reference",
			line:        5, col: 3, endLine: 5, endCol: 11,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Types/multiple-recurrent-definitions-02/invalid.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.selfReference: validate recursive shape: validate array item $.selfReference[0]: validate properties: validate property $.selfReference[0].selfReference: validate recursive shape: validate array item $.selfReference[0].selfReference[0]: validate properties: missing required properties: someProperty"
		{
			fixture:     "Types/nested-self-reference/invalid-property-name.raml",
			msgContains: "someProperty",
			line:        12, col: 7, endLine: 19, endCol: 25,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.someProperty: validate recursive shape: validate properties: validate property $.someProperty.someProperty: validate recursive shape: validate properties: validate property $.someProperty.someProperty.someProperty: validate recursive shape: invalid type, got <nil>, expected map[string]any"
		{
			fixture:     "Types/not-required-property/invalid-missing-required.raml",
			msgContains: "invalid type, got <nil>, expected map[string]any",
			line:        9, col: 9, endLine: 11, endCol: 31,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate pattern property" -> "value does not match any type" -> "validate union member: invalid type, got int, expected []any" -> "validate union member: invalid type, got int, expected string"
		{
			fixture:     "Types/pattern-string-array-property/invalid-wrong-value-type.raml",
			msgContains: "validate shapes",
			line:        -1, col: -1, endLine: -1, endCol: -1,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: object must have at least 2 properties"
		{
			fixture:     "Types/pattern-string-property-01/invalid-minproperties-violated.raml",
			msgContains: "object must have at least 2 properties",
			line:        9, col: 7, endLine: 9, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate pattern property $.prop2: C:\\Sources\\go-raml\\raml-tck\\Types\\pattern-string-property-02\\invalid-unexpected-type.raml:7:7" -> "validate pattern property: invalid type, got map[string]interface {}, expected string"
		{
			fixture:     "Types/pattern-string-property-02/invalid-unexpected-type.raml",
			msgContains: "invalid type, got map[string]interface {}, expected string",
			line:        7, col: 7, endLine: 7, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.prop: validate array item $.prop[0]: invalid type, got string, expected map[string]any"
		{
			fixture:     "Types/property-array-of-datatypes/invalid-array-item-type.raml",
			msgContains: "invalid type, got string, expected map[string]any",
			line:        19, col: 7, endLine: 23, endCol: 29,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.prop: validate array item $.prop[0]: invalid type, got map[string]interface {}, expected string"
		{
			fixture:     "Types/property-array-of-scalars/invalid-array-item-type.raml",
			msgContains: "invalid type, got map[string]interface {}, expected string",
			line:        19, col: 7, endLine: 21, endCol: 15,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression" -> "visit array items" -> "resolve: C:\\Sources\\go-raml\\raml-tck\\Types\\recurrent-array-definition\\invalid.raml:5:3" -> "cyclic type reference"
		{
			fixture:     "Types/recurrent-array-definition/invalid.raml",
			msgContains: "cyclic type reference",
			line:        5, col: 3, endLine: 5, endCol: 11,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression: self recursion SomeType"
		{
			fixture:     "Types/recurrent-definition/invalid.raml",
			msgContains: "self recursion SomeType",
			line:        5, col: 3, endLine: 5, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "decode endpoint" -> "make operation" -> "decode operation" -> "make request" -> "decode media type node" -> "append request body" -> "make body" -> "decode body" -> "make new shape yaml" -> "make shape type" -> "mapping node is not allowed"
		{
			fixture:     "Types/restrictions-conflict/invalid.raml",
			msgContains: "mapping node is not allowed",
			line:        10, col: 11, endLine: 12, endCol: 26,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate pattern property $.X: C:\\Sources\\go-raml\\raml-tck\\Types\\reuse-datatypes-01\\invalid-expected-type.raml:15:7" -> "validate pattern property: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/reuse-datatypes-01/invalid-expected-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        15, col: 7, endLine: 15, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate pattern property $.X: C:\\Sources\\go-raml\\raml-tck\\Types\\reuse-datatypes-02\\invalid-expected-type.raml:15:7" -> "validate pattern property: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/reuse-datatypes-02/invalid-expected-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        15, col: 7, endLine: 15, endCol: 11,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "decode" -> "`type` and `schema` are mutually exclusive"
		{
			fixture:     "Types/scheme/invalid-schema-and-type.raml",
			msgContains: "`type` and `schema` are mutually exclusive",
			line:        6, col: 5, endLine: 6, endCol: 9,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.name: invalid type, got int, expected string"
		{
			fixture:     "Types/single-string-property/invalid-example-type.raml",
			msgContains: "invalid type, got int, expected string",
			line:        12, col: 7, endLine: 12, endCol: 14,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "decode" -> "decode value node" -> "decode example" -> "make example" -> "make node" -> "json unmarshal: invalid character 'p' looking for beginning of object key string"
		{
			fixture:     "Types/single-type-json-example/invalid-json-example.raml",
			msgContains: "invalid character 'p' looking for beginning of object key string",
			line:        7, col: 14, endLine: 7, endCol: 37,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.y: invalid type, got int, expected bool"
		{
			fixture:     "Types/single-type-with-example-01/invalid-example-prop-type.raml",
			msgContains: "invalid type, got int, expected bool",
			line:        10, col: 7, endLine: 11, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: unexpected additional property \"z\""
		{
			fixture:     "Types/single-type-with-example-02/invalid-example-property.raml",
			msgContains: "unexpected additional property \"z\"",
			line:        10, col: 7, endLine: 12, endCol: 13,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.y: value must be one of (val1, val2, 3)"
		{
			fixture:     "Types/single-type-with-example-03/invalid-enum-value.raml",
			msgContains: "value must be one of (val1, val2, 3)",
			line:        10, col: 7, endLine: 10, endCol: 14,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.items: array must have at least 5 items"
		{
			fixture:     "Types/single-type-with-example-04/invalid-failed-array-constraints.raml",
			msgContains: "array must have at least 5 items",
			line:        16, col: 7, endLine: 25, endCol: 12,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.items: array must have at least 5 items"
		{
			fixture:     "Types/single-type-with-example-06/invalid-failed-array-minitems.raml",
			msgContains: "array must have at least 5 items",
			line:        12, col: 7, endLine: 12, endCol: 25,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.age: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/single-type-with-example-07/invalid-example-type.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        9, col: 7, endLine: 10, endCol: 15,
		},
		// error chain: "parse api" -> "decode api" -> "schemas and types are mutually exclusive"
		{
			fixture:     "Types/types-and-schemas/invalid-exclusive.raml",
			msgContains: "schemas and types are mutually exclusive",
			line:        17, col: 3, endLine: 17, endCol: 101,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "Types/types-constraits-conflict/invalid-constraints-conflict.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "unwrap shapes" -> "unwrap fragments: unwrapping: C:\\Sources\\go-raml\\raml-tck\\Types\\union-in-array\\invalid-types-conflict.raml:5:4" -> "unwrap shape" -> "unwrap parents" -> "multiple parents unwrap" -> "failed to find compatible union member"
		{
			fixture:     "Types/union-in-array/invalid-types-conflict.raml",
			msgContains: "failed to find compatible union member",
			line:        5, col: 13, endLine: 5, endCol: 19,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example" -> "validate properties" -> "validate property $.unionArray1: C:\\Sources\\go-raml\\raml-tck\\Types\\union-of-scalar-arrays\\invalid-example-array-elements.raml:11:7" -> "value does not match any type" -> "validate union member: validate array item $.unionArray1[1]: invalid type, got int, expected string" -> "validate union member: validate array item $.unionArray1[0]: invalid type, got string, expected a numeric type or json.Number"
		{
			fixture:     "Types/union-of-scalar-arrays/invalid-example-array-elements.raml",
			msgContains: "invalid type, got string, expected a numeric type or json.Number",
			line:        5, col: 3, endLine: 5, endCol: 14,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.z: length must be greater than 5"
		{
			fixture:     "Types/use-as-property-type-01/invalid-violated-minlength.raml",
			msgContains: "length must be greater than 5",
			line:        11, col: 7, endLine: 11, endCol: 11,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.z: must match pattern .5"
		{
			fixture:     "Types/use-as-property-type-02/invalid-pattern-violated.raml",
			msgContains: "must match pattern .5",
			line:        11, col: 7, endLine: 11, endCol: 12,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: value must be greater than or equal to 5" -> "validate example: validate properties: validate property $.z: value must be greater than or equal to 5"
		{
			fixture:     "Types/use-as-property-type-03/invalid-violated-minmax.raml",
			msgContains: "value must be greater than or equal to 5",
			line:        7, col: 14, endLine: 7, endCol: 15,
		},
		// error chain: "validate shapes" -> "check domain extension: validate properties: validate property $.cc: array must have at least 5 items"
		{
			fixture:     "Types/used-in-annotations/invalid-failed-array-minitems.raml",
			msgContains: "array must have at least 5 items",
			line:        13, col: 6, endLine: 13, endCol: 18,
		},
		// error chain: "parse library" -> "decode library" -> "parse types" -> "unmarshal types: make shape" -> "make concrete shape" -> "unmarshal yaml nodes" -> "unmarshal object facet" -> "unmarshal property: C:\\Sources\\go-raml\\raml-tck\\Types\\xml-serialization\\invalid-wrapped-value.raml:6:7" -> "make property" -> "make shape" -> "decode" -> "decode value node" -> "decode xml" -> "decode attribute: cannot unmarshal !!str `fsdf` into bool"
		{
			fixture:     "Types/xml-serialization/invalid-wrapped-value.raml",
			msgContains: "cannot unmarshal !!str `fsdf` into bool",
			line:        9, col: 22, endLine: 9, endCol: 26,
		},
		// error chain: "parse api" -> "decode api" -> "make endpoint" -> "duplicated endpoint"
		{
			fixture:     "spec-examples/APIs/duplicated-uris-invalid.raml",
			msgContains: "duplicated endpoint",
			line:        12, col: 12, endLine: 12, endCol: 12,
		},
		// TODO: currently produces no error -- needs parser fix
		// {
		// 	fixture:     "spec-examples/APIs/external-type-extend-invalid.raml",
		// 	msgContains: "TODO",
		// 	line: -1, col: -1, endLine: -1, endCol: -1,
		// },
		// error chain: "resolve shapes" -> "resolve shape" -> "resolve link" -> "make concrete shape" -> "unmarshal yaml nodes" -> "type-specific are not allowed for JSON external types"
		{
			fixture:     "spec-examples/APIs/external-types-invalid.raml",
			msgContains: "type-specific are not allowed for JSON external types",
			line:        7, col: 5, endLine: 7, endCol: 15,
		},
		// error chain: "resolve shapes" -> "resolve shape" -> "visit type expression" -> "make concrete shape yaml: C:\\Sources\\go-raml\\raml-tck\\spec-examples\\APIs\\invalid-discriminator-usage.raml:11:3" -> "unmarshal yaml nodes" -> "discriminator is not allowed on union types"
		{
			fixture:     "spec-examples/APIs/invalid-discriminator-usage.raml",
			msgContains: "discriminator is not allowed on union types",
			line:        13, col: 5, endLine: 13, endCol: 18,
		},
		// error chain: "validate shapes" -> "check type" -> "minimum must be less than or equal to maximum"
		{
			fixture:     "spec-examples/APIs/multiple-inheritance-3-invalid.raml",
			msgContains: "minimum must be less than or equal to maximum",
			line:        11, col: 3, endLine: 11, endCol: 10,
		},
		// error chain: "validate shapes" -> "validate shape commons" -> "validate example: validate properties: validate property $.comment: invalid type, got <nil>, expected string"
		{
			fixture:     "spec-examples/APIs/null-type-invalid.raml",
			msgContains: "invalid type, got <nil>, expected string",
			line:        13, col: 7, endLine: 14, endCol: 15,
		},
		// error chain: "parse api" -> "decode api" -> "unmarshal resource type definitions" -> "make resource type definition" -> "decode resource type definition" -> "resource type method must be an HTTP method"
		{
			fixture:     "spec-examples/APIs/resourcetypes-traits-no-subresources-invalid.raml",
			msgContains: "resource type method must be an HTTP method",
			line:        9, col: 5, endLine: 9, endCol: 12,
		},
		// error chain: "parse api" -> "decode api" -> "parse types" -> "unmarshal types: make shape" -> "decode" -> "`type` and `schema` are mutually exclusive"
		{
			fixture:     "spec-examples/APIs/type-schema-invalid.raml",
			msgContains: "`type` and `schema` are mutually exclusive",
			line:        9, col: 5, endLine: 9, endCol: 9,
		},
	}

	var totalTests int
	var missingAssertions int

	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			path := filepath.Join(tckDir, filepath.FromSlash(tc.fixture))
			opts := append([]ParseOpt{OptWithUnwrap(), OptWithValidate()}, tc.extraOpts...)
			_, err := ParseFromPath(path, opts...)
			if err == nil {
				t.Fatalf("expected an error, got none")
			}

			node := findSTNode(err, tc.msgContains)
			if node == nil {
				t.Fatalf("no StackTrace node with message containing %q\nfull error:%s",
					tc.msgContains, formatSTDetails(err, nil))
			}
			if tc.line != -1 {
				if node.Position == nil {
					t.Errorf("node %q: expected position line %d col %d endLine %d endCol %d, got no position",
						node.Message, tc.line, tc.col, tc.endLine, tc.endCol)
				} else {
					p := node.Position
					if p.Line != tc.line {
						t.Errorf("node %q: line: want %d, got %d", node.Message, tc.line, p.Line)
					}
					if tc.col != -1 && p.Column != tc.col {
						t.Errorf("node %q: col: want %d, got %d", node.Message, tc.col, p.Column)
					}
					if tc.endLine != -1 && p.EndLine != tc.endLine {
						t.Errorf("node %q: endLine: want %d, got %d", node.Message, tc.endLine, p.EndLine)
					}
					if tc.endCol != -1 && p.EndColumn != tc.endCol {
						t.Errorf("node %q: endCol: want %d, got %d", node.Message, tc.endCol, p.EndColumn)
					}
				}
			}
		})
		totalTests++
	}

	// Compare active assertions against the ground truth of invalid fixtures.
	missingAssertions = totalInvalidFiles - totalTests
	if missingAssertions < 0 {
		missingAssertions = 0
	}
	t.Logf("TCK Invalid: %d total invalid fixtures, %d assertions active, %d missing",
		totalInvalidFiles, totalTests, missingAssertions)
}

// Test_TCKErrorDiagnostics prints the full error details (message, position)
// produced by go-raml for a handful of invalid TCK fixtures.  It always passes;
// its purpose is to make actual error output visible so we can write precise
// assertions in Test_TCKInvalidErrors.
// Remove irrelevant/already fixed fixtures, add new ones
func Test_TCKErrorDiagnostics(t *testing.T) {
	tckDir := "./raml-tck"
	if _, err := os.Stat(tckDir); err != nil {
		t.Skipf("raml-tck directory not found at %s: %v", tckDir, err)
	}

	probes := []string{
		"Root/documentation/invalid-wrong-format.raml",
	}

	for _, rel := range probes {
		path := filepath.Join(tckDir, filepath.FromSlash(rel))
		absPath, _ := filepath.Abs(path)
		lines, _ := readFileLines(path)
		_, err := ParseFromPath(path, OptWithUnwrap(), OptWithValidate())
		if err == nil {
			t.Logf("%s\n  => no error (unexpected)", absPath)
			continue
		}
		t.Logf("%s\n  => %s", absPath, formatSTDetails(err, lines))
	}
}

// findSTNode walks the full StackTrace tree (Wrapped chain + List) and returns
// the first node whose Message contains substr.
func findSTNode(err error, substr string) *stacktrace.StackTrace {
	var root *stacktrace.StackTrace
	if !errors.As(err, &root) {
		return nil
	}
	var walk func(n *stacktrace.StackTrace) *stacktrace.StackTrace
	walk = func(n *stacktrace.StackTrace) *stacktrace.StackTrace {
		if n == nil {
			return nil
		}
		if strings.Contains(n.Message, substr) {
			return n
		}
		if found := walk(n.Wrapped); found != nil {
			return found
		}
		for _, child := range n.List {
			if found := walk(child); found != nil {
				return found
			}
		}
		return nil
	}
	return walk(root)
}

// formatSTDetails walks a stacktrace error and returns a compact human-readable
// summary of every node's message, position, and (when lines != nil) the source
// text captured by that position range.
func formatSTDetails(err error, lines []string) string {
	var st *stacktrace.StackTrace
	if !errors.As(err, &st) {
		return fmt.Sprintf("(plain) %v", err)
	}
	var b strings.Builder
	var walk func(n *stacktrace.StackTrace, depth int)
	walk = func(n *stacktrace.StackTrace, depth int) {
		if n == nil {
			return
		}
		indent := strings.Repeat("  ", depth)
		pos := ""
		if n.Position != nil {
			pos = fmt.Sprintf(" [line %d col %d endLine %d endCol %d]",
				n.Position.Line, n.Position.Column, n.Position.EndLine, n.Position.EndColumn)
			if txt := extractPositionText(lines, n.Position); txt != "" {
				pos += fmt.Sprintf(" %q", txt)
			}
		}
		typ := ""
		if n.Type != nil {
			typ = fmt.Sprintf(" (%s)", *n.Type)
		}
		fmt.Fprintf(&b, "\n%s%s%s%s", indent, n.Message, pos, typ)
		walk(n.Wrapped, depth+1)
		for _, child := range n.List {
			walk(child, depth+1)
		}
	}
	walk(st, 0)
	return b.String()
}

// readFileLines reads a file and returns its lines (without newline characters).
func readFileLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"), nil
}

// extractPositionText extracts the source text covered by pos from lines.
// Lines and columns are 1-based; EndColumn is 1-based inclusive.
// Returns an empty string when the position is absent or out of range.
func extractPositionText(lines []string, pos *stacktrace.Position) string {
	if lines == nil || pos == nil || pos.Line == 0 || pos.Line > len(lines) {
		return ""
	}
	startLine := pos.Line - 1 // 0-indexed
	startCol := pos.Column - 1
	if startCol < 0 {
		startCol = 0
	}
	endLine := startLine
	if pos.EndLine > pos.Line {
		endLine = pos.EndLine - 1
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	clamp := func(v, max int) int {
		if v > max {
			return max
		}
		return v
	}

	if startLine == endLine {
		line := lines[startLine]
		sc := clamp(startCol, len(line))
		ec := len(line)
		if pos.EndColumn > 0 && pos.EndLine == pos.Line {
			ec = clamp(pos.EndColumn-1, len(line))
		}
		return line[sc:ec]
	}

	// Multi-line: first line from startCol, middle lines in full, last line up to EndColumn.
	var parts []string
	first := lines[startLine]
	parts = append(parts, fmt.Sprintf("L%d:%s", pos.Line, first[clamp(startCol, len(first)):]))
	for i := startLine + 1; i < endLine; i++ {
		parts = append(parts, fmt.Sprintf("L%d:%s", i+1, lines[i]))
	}
	last := lines[endLine]
	ec := len(last)
	if pos.EndColumn > 0 {
		ec = clamp(pos.EndColumn-1, len(last))
	}
	parts = append(parts, fmt.Sprintf("L%d:%s", pos.EndLine, last[:ec]))
	return strings.Join(parts, " | ")
}
