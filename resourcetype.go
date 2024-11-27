package raml

import (
	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type ResourceTypeDefinition struct {
	ID int64

	Usage       string
	Description string

	Source        []*yaml.Node
	VariableNodes map[string]*yaml.Node
	Link          *ResourceTypeFragment

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

type ResourceType struct {
	ID int64

	Name string

	Location string
	stacktrace.Position
	raml *RAML
}
