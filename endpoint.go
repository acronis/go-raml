package raml

import (
	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type EndPoint struct {
	ID int64

	URI           string
	FullURI       string
	URIParameters *orderedmap.OrderedMap[string, *BaseShape]

	DisplayName string
	Description string

	ResourceType *ResourceType
	Traits       []*Trait
	// TODO: Maybe introduce "Security" with resolved security scheme?
	SecuredBy []*SecurityScheme

	EndPoints  *orderedmap.OrderedMap[string, *EndPoint]
	Operations *orderedmap.OrderedMap[string, *Operation]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeEndpoint(valueNode *yaml.Node, location string, uri string, parent string) (*EndPoint, error) {
	endPoint := &EndPoint{
		URI:     uri,
		FullURI: parent + uri,

		EndPoints:  orderedmap.New[string, *EndPoint](0),
		Operations: orderedmap.New[string, *Operation](0),

		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		raml:     r,
		Position: stacktrace.Position{Line: valueNode.Line, Column: valueNode.Column},
		Location: location,
	}

	if _, ok := r.endPoints[endPoint.FullURI]; ok {
		return nil, StacktraceNew("duplicated endpoint", location, WithNodePosition(valueNode), stacktrace.WithInfo("uri", endPoint.FullURI))
	}

	if err := endPoint.decode(valueNode); err != nil {
		return nil, StacktraceNewWrapped("decode endpoint", err, location, WithNodePosition(valueNode))
	}

	r.endPoints[endPoint.FullURI] = endPoint

	return endPoint, nil
}

func (e *EndPoint) decode(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("endpoint must be a mapping node", e.Location, WithNodePosition(node))
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case FacetDisplayName:
			if err := valueNode.Decode(&e.DisplayName); err != nil {
				return StacktraceNewWrapped("decode displayName", err, e.Location, WithNodePosition(valueNode))
			}
		case FacetDescription:
			if err := valueNode.Decode(&e.Description); err != nil {
				return StacktraceNewWrapped("decode description", err, e.Location, WithNodePosition(valueNode))
			}
		case FacetUriParameters:
			if err := e.unmarshalURIParameters(valueNode); err != nil {
				return StacktraceNewWrapped("unmarshal uri parameters", err, e.Location, WithNodePosition(valueNode))
			}
		case FacetType:
			// TODO: ResourceType
		case FacetIs:
			traits, err := e.raml.makeTraits(valueNode, e.Location)
			if err != nil {
				return StacktraceNewWrapped("make traits", err, e.Location, WithNodePosition(valueNode))
			}
			e.Traits = traits
		case FacetSecuredBy:
			securitySchemes, err := e.raml.makeSecuritySchemes(valueNode, e.Location)
			if err != nil {
				return StacktraceNewWrapped("make security schemes", err, e.Location, WithNodePosition(valueNode))
			}
			e.SecuredBy = securitySchemes
		case "get", "post", "put", "delete", "patch", "options", "head", "trace", "connect":
			operation, err := e.raml.makeOperation(keyNode.Value, e.Location, valueNode)
			if err != nil {
				return StacktraceNewWrapped("make operation", err, e.Location, WithNodePosition(valueNode))
			}
			e.Operations.Set(keyNode.Value, operation)
		default:
			switch {
			case IsCustomDomainExtensionNode(keyNode.Value):
				name, de, err := e.raml.unmarshalCustomDomainExtension(e.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, e.Location, WithNodePosition(valueNode))
				}
				e.CustomDomainProperties.Set(name, de)
			case IsEndPoint(keyNode.Value):
				endpoint, err := e.raml.makeEndpoint(valueNode, e.Location, keyNode.Value, e.FullURI)
				if err != nil {
					return StacktraceNewWrapped("make endpoint", err, e.Location, WithNodePosition(valueNode))
				}
				e.EndPoints.Set(keyNode.Value, endpoint)
			default:
				return StacktraceNew("unknown field", e.Location, stacktrace.WithInfo("field", keyNode.Value))
			}
		}
	}
	return nil
}

func (e *EndPoint) unmarshalURIParameters(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("uriParameters must be a mapping node", e.Location, WithNodePosition(node))
	}

	e.URIParameters = orderedmap.New[string, *BaseShape](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		shape, err := e.raml.makeNewShapeYAML(data, nodeName, e.Location)
		if err != nil {
			return StacktraceNewWrapped("make new shape yaml", err, e.Location, WithNodePosition(data))
		}
		e.URIParameters.Set(nodeName, shape)
		e.raml.PutTypeDefinitionIntoFragment(e.Location, shape)
	}
	return nil
}

func IsEndPoint(name string) bool {
	return name != "" && name[0] == '/'
}
