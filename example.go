package raml

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Example represents an example of a shape
type Example struct {
	Id              string
	Name            string
	Value           any
	StructuredValue Node
	MIME            string
	Location        string

	// To support !include
	Link *Example
	Position
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (e *Example) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!include" {
		baseDir := filepath.Dir(e.Location)
		example, err := ReadExample(filepath.Join(baseDir, value.Value))
		if err != nil {
			return fmt.Errorf("read example: %w", err)
		}
		e.Link = example
		return nil
	}
	e.MIME = DetectContentType(value.Value)
	e.Value = value.Value
	return nil
}

// DetectContentType detects the content type of the given value.
// TODO: Detect content type based on the file header
func DetectContentType(value string) string {
	return "text/plain"
}

// DetectFileMimeType detects the MIME type of the file at the given path.
func DetectFileMimeType(path string) string {
	switch filepath.Ext(path) {
	default:
		return "text/plain"
	case ".json":
		return "application/json"
	case ".yaml":
		return "application/yaml"
	case ".xml":
		return "application/xml"
	case ".html":
		return "text/html"
	case ".md":
		return "text/markdown"
	case ".raml":
		return "application/x-raml"
	}
}

// ReadExample reads an example from the file at the given path.
// TODO: Move to parse.go?
func ReadExample(path string) (*Example, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return &Example{
		Value:    string(bytes),
		Location: path,
		MIME:     DetectFileMimeType(path),
	}, nil
}
