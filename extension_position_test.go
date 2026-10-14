package raml

import (
	"errors"
	"strings"
	"testing"

	"github.com/acronis/go-stacktrace"
)

// libWith returns a RAML library whose annotation types are the ones below
// and whose root carries the given annotation usage.
func libWith(usage string) string {
	return `#%RAML 1.0 Library
annotationTypes:
  Meta:
    type: object
    additionalProperties: false
    properties:
      always_on: boolean
      consent_level: string
      nested?:
        type: object
        properties:
          depth: integer
  Metas:
    type: array
    items:
      type: object
      properties:
        id: string
        always_on: boolean
  Numbers:
    type: array
    items: integer
types:
  Thing: object
` + usage
}

// positionOf returns the 1-based line and column of the first occurrence of
// needle in usage, which is the tail of src.
func positionOf(t *testing.T, src, usage, needle string) stacktrace.Position {
	t.Helper()
	start := strings.Index(src, usage)
	i := strings.Index(usage, needle)
	if start < 0 || i < 0 {
		t.Fatalf("%q not found in the usage", needle)
	}
	i += start
	before := src[:i]
	line := strings.Count(before, "\n") + 1
	col := i - strings.LastIndex(before, "\n")
	return stacktrace.Position{Line: line, Column: col}
}

// extensionFrame parses the library and returns the "check domain extension"
// frame of the validation error.
func extensionFrame(t *testing.T, src string) *stacktrace.StackTrace {
	t.Helper()
	return frameWithPrefix(t, src, "check domain extension")
}

// frameWithPrefix parses the library, validates it and returns the frame of
// the validation error whose message starts with prefix.
func frameWithPrefix(t *testing.T, src, prefix string) *stacktrace.StackTrace {
	t.Helper()
	r, err := ParseFromString(src, "lib.raml", t.TempDir())
	if err != nil {
		t.Fatalf("ParseFromString() error = %v", err)
	}
	err = r.ValidateShapes()
	if err == nil {
		t.Fatal("ValidateShapes() succeeded, want a validation error")
	}
	st, ok := stacktrace.Unwrap(err)
	if !ok {
		t.Fatalf("error %v carries no stack trace", err)
	}
	var found *stacktrace.StackTrace
	var walk func(n *stacktrace.StackTrace)
	walk = func(n *stacktrace.StackTrace) {
		for ; n != nil && found == nil; n = n.Wrapped {
			if strings.HasPrefix(n.Message, prefix) {
				found = n
				return
			}
			for _, child := range n.List {
				walk(child)
			}
		}
	}
	walk(st)
	if found == nil {
		t.Fatalf("no %q frame in %v", prefix, errors.Unwrap(err))
	}
	return found
}

func TestDomainExtensionErrorPointsAtTheProperty(t *testing.T) {
	tests := []struct {
		name   string
		usage  string
		needle string
	}{
		{
			name:   "wrong type of a property",
			usage:  "(Meta):\n  consent_level: NONE\n  always_on: \"maybe\"\n",
			needle: "always_on:",
		},
		{
			name:   "wrong type of a nested property",
			usage:  "(Meta):\n  always_on: true\n  consent_level: NONE\n  nested:\n    depth: deep\n",
			needle: "depth:",
		},
		{
			name:   "wrong type of a property in an array item",
			usage:  "(Metas):\n  - id: first\n    always_on: true\n  - id: second\n    always_on: \"maybe\"\n",
			needle: "always_on: \"maybe\"",
		},
		{
			name:   "wrong type of an array element",
			usage:  "(Numbers):\n  - 1\n  - two\n",
			needle: "two",
		},
		{
			name:   "unexpected property",
			usage:  "(Meta):\n  always_on: true\n  consent_level: NONE\n  bogus: 1\n",
			needle: "bogus:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := libWith(tt.usage)

			got := extensionFrame(t, src)

			want := positionOf(t, src, tt.usage, tt.needle)
			if got.Position == nil || *got.Position != want {
				t.Errorf("position = %v, want %v", got.Position, want)
			}
		})
	}
}

func TestDomainExtensionMissingPropertyKeepsTheObjectPosition(t *testing.T) {
	usage := "(Meta):\n  always_on: true\n"
	src := libWith(usage)

	got := extensionFrame(t, src)

	want := positionOf(t, src, usage, "always_on: true")
	if got.Position == nil || got.Position.Line != want.Line {
		t.Errorf("position = %v, want the line of the object start %v", got.Position, want)
	}
}

func TestDomainExtensionErrorTextIsUnchanged(t *testing.T) {
	src := libWith("(Meta):\n  consent_level: NONE\n  always_on: \"maybe\"\n")

	got := extensionFrame(t, src)

	const want = "check domain extension: validate properties: validate property $.always_on: " +
		"validate property $.always_on: invalid type, got string, expected bool"
	if got.Message != want {
		t.Errorf("message = %q, want %q", got.Message, want)
	}
}

func TestCustomFacetErrorPointsAtTheProperty(t *testing.T) {
	usage := `  Parent:
    facets:
      limits:
        type: object
        properties:
          mode: string
          timeout: integer
  Child:
    type: Parent
    limits:
      mode: fast
      timeout: "abc"
`
	src := "#%RAML 1.0 Library\ntypes:\n" + usage

	got := frameWithPrefix(t, src, "validate custom facet")

	want := positionOf(t, src, usage, "timeout: \"abc\"")
	if got.Position == nil || *got.Position != want {
		t.Errorf("position = %v, want %v", got.Position, want)
	}
}
