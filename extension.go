package raml

import (
	"gopkg.in/yaml.v3"
)

type CustomDomainProperties map[string]*DomainExtension

type CustomShapeFacets map[string]*Node
type CustomShapeFacetDefinitions map[string]Property // Object properties share the same syntax with custom shape facets.

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
		return "", nil, NewError("annotation name must not be empty", location, WithNodePosition(keyNode))
	}
	n, err := MakeNode(valueNode, location)
	if err != nil {
		return "", nil, NewWrappedError("make node", err, location, WithNodePosition(valueNode))
	}
	de := &DomainExtension{
		Name:      name,
		Extension: n,
		Location:  location,
		Position:  Position{keyNode.Line, keyNode.Column},
	}
	GetRegistry().DomainExtensions = append(GetRegistry().DomainExtensions, de)
	return name, de, nil
}

func IsCustomDomainExtensionNode(name string) bool {
	return name != "" && name[0] == '(' && name[len(name)-1] == ')'
}
