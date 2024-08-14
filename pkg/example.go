package goraml

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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

func (e *Example) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!include" {
		baseDir := filepath.Dir(e.Location)
		example, err := ReadExample(filepath.Join(baseDir, value.Value))
		if err != nil {
			return err
		}
		e.Link = example
		return nil
	}
	e.MIME = DetectContentType(value.Value)
	e.Value = value.Value
	return nil
}

// TODO: Detect content type based on the file header
func DetectContentType(value string) string {
	return "text/plain"
}

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

// TODO: Move to parse.go?
func ReadExample(path string) (*Example, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &Example{
		Value:    string(bytes),
		Location: path,
		MIME:     DetectFileMimeType(path),
	}, nil
}
