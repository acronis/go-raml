package raml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type CustomDomainProperties map[string]*DomainExtension

type CustomShapeFacets map[string]*Node
type CustomShapeFacetDefinitions map[string]*Shape

type DomainExtension struct {
	Id        string
	Name      string
	Extension *Node
	DefinedBy *Shape

	Location string
	Position
}

func UnmarshalCustomDomainExtension(location string, keyNode *yaml.Node, valueNode *yaml.Node) (string, *DomainExtension, error) {
	name := keyNode.Value[1 : len(keyNode.Value)-1]
	if name == "" {
		return "", nil, fmt.Errorf("annotation name must not be empty")
	}
	dt, err := MakeNode(valueNode, location)
	if err != nil {
		return "", nil, fmt.Errorf("make node: %w", err)
	}
	de := &DomainExtension{
		Name:      name,
		Extension: dt,
		Location:  location,
	}
	return name, de, nil
}

func IsCustomDomainExtensionNode(name string) bool {
	return name != "" && name[0] == '(' && name[len(name)-1] == ')'
}
