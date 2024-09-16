package raml

import (
	"gopkg.in/yaml.v3"
)

type DomainExtension struct {
	Id        string
	Name      string
	Extension *Node
	DefinedBy *Shape

	Location string
	Position
	raml *RAML
}

func (r *RAML) unmarshalCustomDomainExtension(location string, keyNode *yaml.Node, valueNode *yaml.Node) (string, *DomainExtension, error) {
	name := keyNode.Value[1 : len(keyNode.Value)-1]
	if name == "" {
		return "", nil, NewError("annotation name must not be empty", location, WithNodePosition(keyNode))
	}
	n, err := r.makeNode(valueNode, location)
	if err != nil {
		return "", nil, NewWrappedError("make node", err, location, WithNodePosition(valueNode))
	}
	de := &DomainExtension{
		Name:      name,
		Extension: n,
		Location:  location,
		Position:  Position{keyNode.Line, keyNode.Column},
		raml:      r,
	}
	r.domainExtensions = append(r.domainExtensions, de)
	return name, de, nil
}

func IsCustomDomainExtensionNode(name string) bool {
	return name != "" && name[0] == '(' && name[len(name)-1] == ')'
}
