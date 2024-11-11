package raml

import (
	"path/filepath"
	"strings"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type SecuritySchemeDefinition struct {
	ID int64

	Type        SecuritySchemeType
	DisplayName string
	Description string
	Settings    SecuritySchemeSettings
	DescribedBy *SecuritySchemeDescription

	Link *SecuritySchemeFragment

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeSecuritySchemeDefinition(node *yaml.Node, location string) (*SecuritySchemeDefinition, error) {
	securitySchemeDef := &SecuritySchemeDefinition{
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		Location: location,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		raml:     r,
	}

	if err := securitySchemeDef.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode security scheme definition", err, location, WithNodePosition(node))
	}

	return securitySchemeDef, nil
}

func (r *RAML) makeNullSecuritySchemeDefinition(node *yaml.Node, location string) *SecuritySchemeDefinition {
	return &SecuritySchemeDefinition{
		Type:     NullAuthType,
		Settings: &NullAuthScheme{Location: location},

		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		Location: location,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		raml:     r,
	}
}

func (ssd *SecuritySchemeDefinition) decode(node *yaml.Node) error {
	if node.Tag == TagInclude {
		baseDir := filepath.Dir(ssd.Location)
		securitySchemeFrag, err := ssd.raml.parseSecuritySchemeFragment(filepath.Join(baseDir, node.Value))
		if err != nil {
			return StacktraceNewWrapped("parse security scheme fragment", err, ssd.Location, WithNodePosition(node))
		}
		ssd.Link = securitySchemeFrag
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("security scheme definition must be a mapping node", ssd.Location, WithNodePosition(node))
	}

	var typeNode *yaml.Node
	var settingsNode *yaml.Node
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case "type":
			if err := valueNode.Decode(&ssd.Type); err != nil {
				return StacktraceNewWrapped("decode type", err, ssd.Location, WithNodePosition(valueNode))
			}
			typeNode = valueNode
		case "displayName":
			if err := valueNode.Decode(&ssd.DisplayName); err != nil {
				return StacktraceNewWrapped("decode displayName", err, ssd.Location, WithNodePosition(valueNode))
			}
		case "description":
			if err := valueNode.Decode(&ssd.Description); err != nil {
				return StacktraceNewWrapped("decode description", err, ssd.Location, WithNodePosition(valueNode))
			}
		case "describedBy":
			describedBy, err := ssd.raml.makeSecuritySchemeDescription(ssd.Location, valueNode)
			if err != nil {
				return StacktraceNewWrapped("make security scheme description", err, ssd.Location)
			}
			ssd.DescribedBy = describedBy
		case "settings":
			settingsNode = valueNode
		default:
			if IsCustomDomainExtensionNode(keyNode.Value) {
				name, de, err := ssd.raml.unmarshalCustomDomainExtension(ssd.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, ssd.Location, WithNodePosition(valueNode))
				}
				ssd.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown key in security scheme definition", ssd.Location, WithNodePosition(keyNode), stacktrace.WithInfo("key", keyNode.Value))
			}
		}
	}
	if typeNode == nil {
		return StacktraceNew("security scheme definition must have a type", ssd.Location)
	}

	settings, err := ssd.raml.MakeSecuritySchemeSettings(ssd.Type, settingsNode, ssd.Location)
	if err != nil {
		return StacktraceNewWrapped("make security scheme settings", err, ssd.Location)
	}
	ssd.Settings = settings

	return nil
}

type SecuritySchemeDescription struct {
	ID int64

	Headers         *orderedmap.OrderedMap[string, Property]
	QueryParameters *orderedmap.OrderedMap[string, Property] // TODO: Maybe can be combined?
	QueryString     *BaseShape
	Responses       *orderedmap.OrderedMap[int, *Response]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeSecuritySchemeDescription(location string, node *yaml.Node) (*SecuritySchemeDescription, error) {
	ssd := &SecuritySchemeDescription{
		Headers:         orderedmap.New[string, Property](0),
		QueryParameters: orderedmap.New[string, Property](0),
		Responses:       orderedmap.New[int, *Response](0),

		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),
		Location:               location,
		raml:                   r,
	}

	if err := ssd.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode security scheme description", err, location)
	}

	return ssd, nil
}

func (o *SecuritySchemeDescription) decode(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("security scheme describedBy must be a mapping node", o.Location, WithNodePosition(node))
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case FacetHeaders:
			headers, err := o.raml.unmarshalHeaders(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal headers", err, o.Location, WithNodePosition(valueNode))
			}
			o.Headers = headers
		case FacetQueryParameters:
			if o.QueryString != nil {
				return StacktraceNew("queryParameters and queryString are mutually exclusive", o.Location, WithNodePosition(valueNode))
			}
			params, err := o.raml.unmarshalQueryParameters(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal query parameters", err, o.Location, WithNodePosition(valueNode))
			}
			o.QueryParameters = params
		case FacetQueryString:
			if o.QueryParameters != nil {
				return StacktraceNew("queryParameters and queryString are mutually exclusive", o.Location, WithNodePosition(valueNode))
			}
			shape, err := o.raml.unmarshalQueryString(valueNode, o.Location, keyNode.Value)
			if err != nil {
				return StacktraceNewWrapped("unmarshal query string", err, o.Location, WithNodePosition(valueNode))
			}
			o.QueryString = shape
		case FacetResponses:
			responses, err := o.raml.makeResponses(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("make responses", err, o.Location, WithNodePosition(valueNode))
			}
			o.Responses = responses
		default:
			if IsCustomDomainExtensionNode(keyNode.Value) {
				name, de, err := o.raml.unmarshalCustomDomainExtension(o.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, o.Location, WithNodePosition(valueNode))
				}
				o.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown key in describedBy", o.Location, WithNodePosition(keyNode), stacktrace.WithInfo("key", keyNode.Value))
			}
		}
	}
	return nil
}

func (o *SecuritySchemeDescription) makeOperation() *Operation {
	return &Operation{
		Headers:         o.Headers,
		QueryParameters: o.QueryParameters,
		QueryString:     o.QueryString,
		Responses:       o.Responses,

		CustomDomainProperties: o.CustomDomainProperties,
	}
}

func (r *RAML) MakeSecuritySchemeSettings(schemeType SecuritySchemeType, settingsNode *yaml.Node, location string) (SecuritySchemeSettings, error) {
	var schemeSettings SecuritySchemeSettings
	switch schemeType {
	case BasicAuthType:
		schemeSettings = &BasicAuthSchemeSettings{Location: location}
	case DigestAuthType:
		schemeSettings = &DigestAuthSchemeSettings{Location: location}
	case PassThroughAuthType:
		schemeSettings = &PassThroughSchemeSettings{Location: location}
	case OAuth1AuthType:
		schemeSettings = &OAuth1SchemeSettings{Location: location}
	case OAuth2AuthType:
		schemeSettings = &OAuth2SchemeSettings{Location: location}
	default:
		if strings.HasPrefix(string(schemeType), "x-") {
			schemeSettings = &CustomAuthScheme{Location: location}
		} else {
			return nil, StacktraceNew("unknown security scheme type", location, stacktrace.WithInfo("type", schemeType))
		}
	}

	if settingsNode != nil && settingsNode.Tag != TagNull {
		if settingsNode.Kind != yaml.MappingNode {
			return nil, StacktraceNew("security scheme settings must be a mapping node", location, WithNodePosition(settingsNode))
		}
		if err := schemeSettings.Decode(settingsNode); err != nil {
			return nil, StacktraceNewWrapped("decode settings", err, location, WithNodePosition(settingsNode))
		}
	}

	return schemeSettings, nil
}

type SecuritySchemeSettings interface {
	Apply(params map[string]any) error
	Decode(node *yaml.Node) error
}

type NullAuthScheme struct {
	stacktrace.Position
	Location string
}

func (s *NullAuthScheme) Decode(node *yaml.Node) error {
	return StacktraceNew("null security scheme has no settings", s.Location, WithNodePosition(node))
}

func (s *NullAuthScheme) Apply(params map[string]any) error {
	return StacktraceNew("null security scheme cannot have parameters", s.Location, stacktrace.WithPosition(&s.Position))
}

type BasicAuthSchemeSettings struct {
	stacktrace.Position
	Location string
}

func (s *BasicAuthSchemeSettings) Decode(node *yaml.Node) error {
	return StacktraceNew("basic security scheme has no settings", s.Location, WithNodePosition(node))
}

func (s *BasicAuthSchemeSettings) Apply(params map[string]any) error {
	return StacktraceNew("basic security scheme cannot have parameters", s.Location, stacktrace.WithPosition(&s.Position))
}

type DigestAuthSchemeSettings struct {
	stacktrace.Position
	Location string
}

func (s *DigestAuthSchemeSettings) Decode(node *yaml.Node) error {
	return StacktraceNew("digest security scheme has no settings", s.Location, WithNodePosition(node))
}

func (s *DigestAuthSchemeSettings) Apply(params map[string]any) error {
	return StacktraceNew("digest security scheme cannot have parameters", s.Location, stacktrace.WithPosition(&s.Position))
}

type PassThroughSchemeSettings struct {
	stacktrace.Position
	Location string
}

func (s *PassThroughSchemeSettings) Decode(node *yaml.Node) error {
	return StacktraceNew("passthrough security scheme has no settings", s.Location, WithNodePosition(node))
}

func (s *PassThroughSchemeSettings) Apply(params map[string]any) error {
	return StacktraceNew("passthrough security scheme cannot have parameters", s.Location, stacktrace.WithPosition(&s.Position))
}

type OAuth1SchemeSettings struct {
	RequestTokenURI     string
	AuthorizationURI    string
	TokenCredentialsURI string
	Signatures          []string

	stacktrace.Position
	Location string
}

func (s *OAuth1SchemeSettings) Decode(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case "requestTokenUri":
			if err := valueNode.Decode(&s.RequestTokenURI); err != nil {
				return StacktraceNewWrapped("decode requestTokenUri", err, s.Location, WithNodePosition(valueNode))
			}
		case "authorizationUri":
			if err := valueNode.Decode(&s.AuthorizationURI); err != nil {
				return StacktraceNewWrapped("decode authorizationUri", err, s.Location, WithNodePosition(valueNode))
			}
		case "tokenCredentialsUri":
			if err := valueNode.Decode(&s.TokenCredentialsURI); err != nil {
				return StacktraceNewWrapped("decode tokenCredentialsUri", err, s.Location, WithNodePosition(valueNode))
			}
		case "signatures":
			if err := valueNode.Decode(&s.Signatures); err != nil {
				return StacktraceNewWrapped("decode signatures", err, s.Location, WithNodePosition(valueNode))
			}
		default:
			return StacktraceNew("unknown key in OAuth1 security scheme settings", s.Location, WithNodePosition(keyNode), stacktrace.WithInfo("key", keyNode.Value))
		}
	}
	s.Position = stacktrace.Position{Line: node.Line, Column: node.Column}
	return nil
}

func (s *OAuth1SchemeSettings) Apply(params map[string]any) error {
	// TODO
	return nil
}

type OAuth2SchemeSettings struct {
	AuthorizationURI    string
	AccessTokenURI      string
	AuthorizationGrants []string
	Scopes              []string

	stacktrace.Position
	Location string
}

func (s *OAuth2SchemeSettings) Decode(node *yaml.Node) error {
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case "authorizationUri":
			if err := valueNode.Decode(&s.AuthorizationURI); err != nil {
				return StacktraceNewWrapped("decode authorizationUri", err, s.Location, WithNodePosition(valueNode))
			}
		case "accessTokenUri":
			if err := valueNode.Decode(&s.AccessTokenURI); err != nil {
				return StacktraceNewWrapped("decode accessTokenUri", err, s.Location, WithNodePosition(valueNode))
			}
		case "authorizationGrants":
			if err := valueNode.Decode(&s.AuthorizationGrants); err != nil {
				return StacktraceNewWrapped("decode authorizationGrants", err, s.Location, WithNodePosition(valueNode))
			}
		case "scopes":
			if err := valueNode.Decode(&s.AuthorizationGrants); err != nil {
				return StacktraceNewWrapped("decode authorizationGrants", err, s.Location, WithNodePosition(valueNode))
			}
		default:
			return StacktraceNew("unknown key in OAuth2 security scheme settings", s.Location, WithNodePosition(keyNode), stacktrace.WithInfo("key", keyNode.Value))
		}
	}
	s.Position = stacktrace.Position{Line: node.Line, Column: node.Column}
	return nil
}

func (s *OAuth2SchemeSettings) Apply(params map[string]any) error {
	// TODO
	return nil
}

type CustomAuthScheme struct {
	stacktrace.Position
	Location string
}

func (s *CustomAuthScheme) Decode(node *yaml.Node) error {
	return StacktraceNew("custom security scheme has no settings", s.Location, WithNodePosition(node))
}

func (s *CustomAuthScheme) Apply(params map[string]any) error {
	// TODO
	return nil
}

type SecurityScheme struct {
	ID int64

	Name       string
	Definition *SecuritySchemeDefinition
	Params     map[string]any

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeSecuritySchemes(valueNode *yaml.Node, location string) ([]*SecurityScheme, error) {
	switch valueNode.Kind {
	case yaml.ScalarNode:
		if valueNode.Tag == TagNull {
			return nil, nil
		}
		securityScheme, err := r.makeSecurityScheme(valueNode, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make security scheme", err, location, WithNodePosition(valueNode))
		}
		return []*SecurityScheme{securityScheme}, nil
	case yaml.SequenceNode:
		securitySchemes := make([]*SecurityScheme, len(valueNode.Content))
		for i, node := range valueNode.Content {
			securityScheme, err := r.makeSecurityScheme(node, location)
			if err != nil {
				return nil, StacktraceNewWrapped("make security scheme", err, location, WithNodePosition(node))
			}
			securitySchemes[i] = securityScheme
		}
		return securitySchemes, nil
	default:
		return nil, StacktraceNew("security schemes must be either sequence or scalar node", location, WithNodePosition(valueNode))
	}
}

func (r *RAML) makeSecurityScheme(valueNode *yaml.Node, location string) (*SecurityScheme, error) {
	securityScheme := &SecurityScheme{
		Location: location,
		raml:     r,
		Position: stacktrace.Position{Line: valueNode.Line, Column: valueNode.Column},
	}

	if err := securityScheme.decode(valueNode); err != nil {
		return nil, StacktraceNewWrapped("decode security scheme", err, location, WithNodePosition(valueNode))
	}

	return securityScheme, nil
}

func (ss *SecurityScheme) decode(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Tag == TagNull {
			ss.Name = "null"
			ss.Definition = ss.raml.makeNullSecuritySchemeDefinition(node, ss.Location)
		} else {
			ss.Name = node.Value
		}
	case yaml.MappingNode:
		keyNode := node.Content[0]
		valueNode := node.Content[1]
		ss.Name = keyNode.Value
		if err := valueNode.Decode(&ss.Params); err != nil {
			return StacktraceNewWrapped("decode security scheme parameters", err, ss.Location, WithNodePosition(valueNode))
		}
	default:
		return StacktraceNew("security scheme must be either scalar or mapping node", ss.Location, WithNodePosition(node))
	}
	return nil
}
