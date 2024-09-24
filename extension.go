package raml

import (
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-stacktrace"
)

type DomainExtension struct {
	ID        string
	Name      string
	Extension *Node
	DefinedBy *Shape

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) unmarshalCustomDomainExtension(location string, keyNode *yaml.Node,
	valueNode *yaml.Node,
) (string, *DomainExtension, error) {
	name := keyNode.Value[1 : len(keyNode.Value)-1]
	if name == "" {
		return "", nil, stacktrace.New("annotation name must not be empty", location,
			WithNodePosition(keyNode))
	}
	n, err := r.makeRootNode(valueNode, location)
	if err != nil {
		return "", nil, StacktraceNewWrapped("make node", err, location,
			WithNodePosition(valueNode))
	}
	de := &DomainExtension{
		Name:      name,
		Extension: n,
		Location:  location,
		Position:  stacktrace.Position{Line: keyNode.Line, Column: keyNode.Column},
		raml:      r,
	}
	r.domainExtensions = append(r.domainExtensions, de)
	return name, de, nil
}

func IsCustomDomainExtensionNode(name string) bool {
	return name != "" && name[0] == '(' && name[len(name)-1] == ')'
}
