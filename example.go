package raml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func MakeExample(value *yaml.Node, name string, location string) (*Example, error) {
	// TODO: Example can be either a scalar or a map
	n, err := MakeNode(value, location)
	if err != nil {
		return nil, fmt.Errorf("make node: %w", err)
	}
	return &Example{
		Name:     name,
		Value:    n,
		Location: location,
		Position: Position{Line: value.Line, Column: value.Column},
	}, nil
}

// Example represents an example of a shape
type Example struct {
	Id          string
	Name        string
	DisplayName string
	Description string
	Value       *Node

	CustomDomainProperties CustomDomainProperties

	Location string
	Position
}
