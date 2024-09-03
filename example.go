package raml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// makeExample creates an example from the given value node
func (r *RAML) makeExample(value *yaml.Node, name string, location string) (*Example, error) {
	// TODO: Example can be either a scalar or a map
	n, err := r.makeNode(value, location)
	if err != nil {
		return nil, fmt.Errorf("make node: %w", err)
	}
	return &Example{
		Name:     name,
		Data:     n,
		Location: location,
		Position: Position{Line: value.Line, Column: value.Column},
		raml:     r,
	}, nil
}

// Example represents an example of a shape
type Example struct {
	Id          string
	Name        string
	DisplayName string
	Description string
	Data        *Node

	CustomDomainProperties CustomDomainProperties

	Location string
	Position
	raml *RAML
}
