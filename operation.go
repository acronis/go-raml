package raml

import (
	"strconv"
	"strings"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type HTTPAction interface {
	appendBody(node *yaml.Node, mediaType string) error
}

type Operation struct {
	ID int64

	DisplayName string
	Description string

	Traits []*Trait

	Protocols []string
	SecuredBy []*SecurityScheme

	Method          string
	Headers         *orderedmap.OrderedMap[string, Property]
	QueryParameters *orderedmap.OrderedMap[string, Property] // TODO: Maybe can be combined?
	QueryString     *BaseShape
	Request         *Request
	Responses       *orderedmap.OrderedMap[int, *Response]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeOperation(method string, location string, node *yaml.Node) (*Operation, error) {
	operation := &Operation{
		Method: method,

		Protocols: r.globalProtocols,
		SecuredBy: r.globalSecuredBy,

		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		raml:     r,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		Location: location,
	}

	if err := operation.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode operation", err, location, WithNodePosition(node))
	}

	return operation, nil
}

func (r *RAML) makeParametrizedOperation(location string, node *yaml.Node) (*Operation, error) {
	operation := &Operation{
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		raml:     r,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		Location: location,
	}

	if err := operation.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode operation", err, location, WithNodePosition(node))
	}

	return operation, nil
}

func (r *RAML) unmarshalHeaders(node *yaml.Node, location string) (*orderedmap.OrderedMap[string, Property], error) {
	if node.Tag == TagNull {
		return nil, nil
	} else if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("headers must be a mapping node", location, WithNodePosition(node))
	}

	headers := orderedmap.New[string, Property](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		propertyName, hasImplicitOptional := r.chompImplicitOptional(nodeName)
		property, err := r.makeProperty(nodeName, propertyName, data, location, hasImplicitOptional)
		if err != nil {
			return nil, StacktraceNewWrapped("make property", err, location,
				WithNodePosition(data))
		}
		headers.Set(property.Name, property)
		r.PutTypeDefinitionIntoFragment(location, property.Base)
	}
	return headers, nil
}

func (r *RAML) unmarshalQueryParameters(node *yaml.Node, location string) (*orderedmap.OrderedMap[string, Property], error) {
	if node.Tag == TagNull {
		return nil, nil
	} else if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("queryParameters must be a mapping node", location, WithNodePosition(node))
	}

	queryParameters := orderedmap.New[string, Property](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		nodeName := node.Content[j].Value
		data := node.Content[j+1]

		propertyName, hasImplicitOptional := r.chompImplicitOptional(nodeName)
		property, err := r.makeProperty(nodeName, propertyName, data, location, hasImplicitOptional)
		if err != nil {
			return nil, StacktraceNewWrapped("make property", err, location,
				WithNodePosition(data))
		}
		queryParameters.Set(property.Name, property)
		r.PutTypeDefinitionIntoFragment(location, property.Base)
	}
	return queryParameters, nil
}

func (r *RAML) unmarshalQueryString(node *yaml.Node, location string, name string) (*BaseShape, error) {
	shape, err := r.makeNewShapeYAML(node, name, location)
	if err != nil {
		return nil, StacktraceNewWrapped("make new shape yaml", err, location, WithNodePosition(node))
	}
	r.PutTypeDefinitionIntoFragment(location, shape)
	return shape, nil
}

func (o *Operation) decode(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("operation must be a mapping node", o.Location, WithNodePosition(node))
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case FacetDisplayName:
			if err := valueNode.Decode(&o.DisplayName); err != nil {
				return StacktraceNewWrapped("decode displayName", err, o.Location, WithNodePosition(valueNode))
			}
		case FacetDescription:
			if err := valueNode.Decode(&o.Description); err != nil {
				return StacktraceNewWrapped("decode description", err, o.Location, WithNodePosition(valueNode))
			}
		case FacetProtocols:
			if err := valueNode.Decode(&o.Protocols); err != nil {
				return StacktraceNewWrapped("decode protocols", err, o.Location, WithNodePosition(valueNode))
			}
		case FacetSecuredBy:
			securitySchemes, err := o.raml.makeSecuritySchemes(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("make security schemes", err, o.Location, WithNodePosition(valueNode))
			}
			o.SecuredBy = securitySchemes
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
		case FacetBody:
			request, err := o.raml.makeRequest(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("make request", err, o.Location, WithNodePosition(valueNode))
			}
			o.Request = request
		case FacetResponses:
			responses, err := o.raml.makeResponses(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("make responses", err, o.Location, WithNodePosition(valueNode))
			}
			o.Responses = responses
		case FacetIs:
			traits, err := o.raml.makeTraits(valueNode, o.Location)
			if err != nil {
				return StacktraceNewWrapped("make traits", err, o.Location, WithNodePosition(valueNode))
			}
			o.Traits = traits
		default:
			if IsCustomDomainExtensionNode(keyNode.Value) {
				name, de, err := o.raml.unmarshalCustomDomainExtension(o.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, o.Location, WithNodePosition(valueNode))
				}
				o.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown field", o.Location, stacktrace.WithInfo("field", keyNode.Value))
			}
		}
	}
	return nil
}

func (o *Operation) merge(source *Operation) {
	// TODO: Shapes merge
	if o.DisplayName == "" {
		o.DisplayName = source.DisplayName
	}
	if o.Description == "" {
		o.Description = source.Description
	}
	o.Protocols = append(o.Protocols, source.Protocols...)
	o.SecuredBy = append(o.SecuredBy, source.SecuredBy...)
	if o.Headers == nil {
		o.Headers = source.Headers
	} else if source.Headers != nil {
		for pair := source.Headers.Oldest(); pair != nil; pair = pair.Next() {
			key := pair.Key
			value := pair.Value
			if _, ok := o.Headers.Get(key); !ok {
				o.Headers.Set(key, value)
			}
		}
	}
	if o.QueryParameters == nil {
		o.QueryParameters = source.QueryParameters
	} else if source.QueryParameters != nil {
		for pair := source.QueryParameters.Oldest(); pair != nil; pair = pair.Next() {
			key := pair.Key
			value := pair.Value
			if _, ok := o.QueryParameters.Get(key); !ok {
				o.QueryParameters.Set(key, value)
			}
		}
	}
	if o.QueryString == nil {
		o.QueryString = source.QueryString
	}
	if o.Request == nil {
		o.Request = source.Request
	}
	if o.Responses == nil {
		o.Responses = source.Responses
	} else if source.Responses != nil {
		for pair := source.Responses.Oldest(); pair != nil; pair = pair.Next() {
			key := pair.Key
			value := pair.Value
			if _, ok := o.Responses.Get(key); !ok {
				o.Responses.Set(key, value)
			}
		}
	}
	if o.CustomDomainProperties == nil {
		o.CustomDomainProperties = source.CustomDomainProperties
	} else if source.CustomDomainProperties != nil {
		for pair := source.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
			key := pair.Key
			value := pair.Value
			if _, ok := o.CustomDomainProperties.Get(key); !ok {
				o.CustomDomainProperties.Set(key, value)
			}
		}
	}
}

type Request struct {
	ID int64

	Bodies *orderedmap.OrderedMap[string, *Body]

	// NOTE: Request cannot have custom annotations, those are defined on Operation level in RAML.

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeRequest(node *yaml.Node, location string) (*Request, error) {
	request := &Request{
		Bodies: orderedmap.New[string, *Body](),

		raml:     r,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		Location: location,
	}

	if err := r.decodeMediaTypeNode(request, node, location); err != nil {
		return nil, StacktraceNewWrapped("decode media type node", err, location, WithNodePosition(node))
	}

	return request, nil
}

func (r *Request) appendBody(node *yaml.Node, mediaType string) error {
	body, err := r.raml.makeBody(node, r.Location, mediaType)
	if err != nil {
		return StacktraceNewWrapped("make body", err, r.Location, WithNodePosition(node))
	}
	r.Bodies.Set(mediaType, body)
	return nil
}

func (r *RAML) decodeMediaTypeNode(action HTTPAction, node *yaml.Node, location string) error {
	// TODO: Common for request/responses
	switch node.Kind {
	case yaml.ScalarNode:
		if node.Tag == TagNull {
			return nil
		}
		if r.globalMediaType == nil {
			return StacktraceNew("explicit media type is required", location, WithNodePosition(node))
		}
		for _, mediaType := range r.globalMediaType {
			if err := action.appendBody(node, mediaType); err != nil {
				return StacktraceNewWrapped("append request body", err, location, WithNodePosition(node))
			}
		}
	case yaml.MappingNode:
		mediaTypeNodes, err := r.collectMediaTypes(node, location)
		if err != nil {
			return StacktraceNewWrapped("collect media types", err, location, WithNodePosition(node))
		}
		if mediaTypeNodes == nil {
			if r.globalMediaType == nil {
				return StacktraceNew("explicit media type is required", location, WithNodePosition(node))
			}
			for _, mediaType := range r.globalMediaType {
				if err = action.appendBody(node, mediaType); err != nil {
					return StacktraceNewWrapped("append request body", err, location, WithNodePosition(node))
				}
			}
		} else {
			for i := 0; i != len(mediaTypeNodes); i += 2 {
				keyNode := mediaTypeNodes[i]
				valueNode := mediaTypeNodes[i+1]
				if err = action.appendBody(valueNode, keyNode.Value); err != nil {
					return StacktraceNewWrapped("append request body", err, location, WithNodePosition(node))
				}
			}
		}
	default:
		return StacktraceNew("request must be either scalar or mapping node", location, WithNodePosition(node))
	}
	return nil
}

func (r *RAML) collectMediaTypes(node *yaml.Node, location string) ([]*yaml.Node, error) {
	var foreignNodes []*yaml.Node
	var mediaTypes []*yaml.Node
	for i := 0; i != len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		if strings.IndexByte(keyNode.Value, '/') == -1 {
			foreignNodes = append(foreignNodes, keyNode)
		} else {
			mediaTypes = append(mediaTypes, keyNode, valueNode)
		}
	}

	// If there are media types and there are foreign nodes, then it is an error.
	if mediaTypes != nil && len(foreignNodes) > 0 {
		st := StacktraceNew("Found unexpected keys instead of valid media types", location, WithNodePosition(node))
		for _, fNode := range foreignNodes {
			st = st.Append(StacktraceNew("Unexpected key", location, WithNodePosition(fNode)))
		}
		return nil, st
	}

	return mediaTypes, nil
}

type Response struct {
	ID int64

	DisplayName string
	Description string

	StatusCode int
	Headers    *orderedmap.OrderedMap[string, Property]
	Bodies     *orderedmap.OrderedMap[string, *Body]

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeResponses(node *yaml.Node, location string) (*orderedmap.OrderedMap[int, *Response], error) {
	if node.Tag == TagNull {
		return nil, nil
	} else if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("responses must be a mapping node", location, WithNodePosition(node))
	}

	responses := orderedmap.New[int, *Response](len(node.Content) / 2)
	for j := 0; j != len(node.Content); j += 2 {
		statusCode := node.Content[j]
		data := node.Content[j+1]

		if !IsStatusCode(statusCode.Value) {
			return nil, StacktraceNew("status code must be a 3-digit number", location, WithNodePosition(statusCode))
		}
		intStatusCode, err := strconv.Atoi(statusCode.Value)
		if err != nil {
			return nil, StacktraceNewWrapped("parse status code", err, location, WithNodePosition(statusCode))
		}

		response, err := r.makeResponse(data, location, intStatusCode)
		if err != nil {
			return nil, StacktraceNewWrapped("make response", err, location, WithNodePosition(data))
		}
		responses.Set(intStatusCode, response)
	}
	return responses, nil
}

func (r *RAML) makeResponse(node *yaml.Node, location string, statusCode int) (*Response, error) {
	response := &Response{
		StatusCode: statusCode,
		Headers:    orderedmap.New[string, Property](0),
		Bodies:     orderedmap.New[string, *Body](0),

		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		raml:     r,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		Location: location,
	}

	if err := response.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode response", err, location, WithNodePosition(node))
	}

	return response, nil
}

func (r *Response) decode(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("responses must be a mapping node", r.Location, WithNodePosition(node))
	}

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case FacetDisplayName:
			if err := valueNode.Decode(&r.DisplayName); err != nil {
				return StacktraceNewWrapped("decode displayName", err, r.Location, WithNodePosition(valueNode))
			}
		case FacetDescription:
			if err := valueNode.Decode(&r.Description); err != nil {
				return StacktraceNewWrapped("decode description", err, r.Location, WithNodePosition(valueNode))
			}
		case FacetHeaders:
			headers, err := r.raml.unmarshalHeaders(valueNode, r.Location)
			if err != nil {
				return StacktraceNewWrapped("unmarshal headers", err, r.Location, WithNodePosition(valueNode))
			}
			r.Headers = headers
		case "body":
			if err := r.raml.decodeMediaTypeNode(r, valueNode, r.Location); err != nil {
				return StacktraceNewWrapped("decode media type node", err, r.Location, WithNodePosition(valueNode))
			}
		default:
			if IsCustomDomainExtensionNode(keyNode.Value) {
				name, de, err := r.raml.unmarshalCustomDomainExtension(r.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, r.Location, WithNodePosition(valueNode))
				}
				r.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown field", r.Location, stacktrace.WithInfo("field", keyNode.Value))
			}
		}
	}
	return nil
}

func (r *Response) appendBody(node *yaml.Node, mediaType string) error {
	body, err := r.raml.makeBody(node, r.Location, mediaType)
	if err != nil {
		return StacktraceNewWrapped("make body", err, r.Location, WithNodePosition(node))
	}
	r.Bodies.Set(mediaType, body)
	return nil
}

func IsStatusCode(value string) bool {
	return len(value) == 3 &&
		value[0] >= '1' && value[0] <= '5' &&
		value[1] >= '0' && value[1] <= '9' &&
		value[2] >= '0' && value[2] <= '9'
}

type Body struct {
	ID int64

	MediaType string
	Shape     *BaseShape // can be null!

	// NOTE: Body cannot have annotations. Those are defined on Shape level in RAML.

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeBody(node *yaml.Node, location string, mediaType string) (*Body, error) {
	body := &Body{
		MediaType: mediaType,

		raml:     r,
		Position: stacktrace.Position{Line: node.Line, Column: node.Column},
		Location: location,
	}

	if err := body.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode body", err, location, WithNodePosition(node))
	}

	return body, nil
}

func (b *Body) decode(node *yaml.Node) error {
	shape, err := b.raml.makeNewShapeYAML(node, "response", b.Location)
	if err != nil {
		return StacktraceNewWrapped("make new shape yaml", err, b.Location, WithNodePosition(node))
	}
	b.Shape = shape
	b.raml.PutTypeDefinitionIntoFragment(b.Location, shape)
	return nil
}
