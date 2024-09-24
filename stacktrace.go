package raml

import (
	"errors"
	"fmt"
	"strings"

	"github.com/acronis/go-stacktrace"
	"gopkg.in/yaml.v3"
)

// NewNodePosition creates a new position from the given node.
func NewNodePosition(node *yaml.Node) *stacktrace.Position {
	return &stacktrace.Position{Line: node.Line, Column: node.Column}
}

// optErrNodePosition is an option to set the position of the error to the position of the given node.
type optErrNodePosition struct {
	pos *stacktrace.Position
}

// Apply sets the position of the error to the given position.
// implements Option
func (o optErrNodePosition) Apply(e *stacktrace.StackTrace) {
	e.Position = o.pos
}

// WithNodePosition sets the position of the error to the position of the given node.
func WithNodePosition(node *yaml.Node) stacktrace.Option {
	return optErrNodePosition{pos: NewNodePosition(node)}
}

// GetYamlError returns the yaml type error from the given error.
// nil if the error is not a yaml type error.
func GetYamlError(err error) *yaml.TypeError {
	var yamlError *yaml.TypeError
	if errors.As(err, &yamlError) {
		return yamlError
	}
	wErr := errors.Unwrap(err)
	if wErr == nil {
		return nil
	}

	if yamlErr := GetYamlError(wErr); yamlErr != nil {
		toAppend := strings.ReplaceAll(err.Error(), yamlErr.Error(), "")
		toAppend = strings.TrimSuffix(toAppend, ": ")
		// insert the error message in the correct order to the first index
		yamlErr.Errors = append([]string{toAppend}, yamlErr.Errors...)
		return yamlErr
	}
	return nil
}

// FixYamlError fixes the yaml type error from the given error.
func FixYamlError(err error) error {
	if err == nil {
		return nil
	}
	if yamlError := GetYamlError(err); yamlError != nil {
		err = fmt.Errorf("%s", strings.Join(yamlError.Errors, ": "))
	}
	return err
}

func StacktraceNewWrapped(msg string, err error, location string, opts ...stacktrace.Option) *stacktrace.StackTrace {
	return stacktrace.NewWrapped(msg, FixYamlError(err), location, opts...)
}
