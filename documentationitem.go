package raml

import (
	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type DocumentationItem struct {
	ID int64

	Title   string
	Content string

	Link *DocumentationItemFragment

	CustomDomainProperties *orderedmap.OrderedMap[string, *DomainExtension]

	Location string
	stacktrace.Position
	raml *RAML
}

func (di *DocumentationItem) decode(node *yaml.Node) error {
	for i := 0; i != len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case "title":
			if err := valueNode.Decode(&di.Title); err != nil {
				return StacktraceNewWrapped("decode title", err, di.Location, WithNodePosition(valueNode))
			}
		case "content":
			if err := valueNode.Decode(&di.Content); err != nil {
				return StacktraceNewWrapped("decode content", err, di.Location, WithNodePosition(valueNode))
			}
		default:
			if IsCustomDomainExtensionNode(keyNode.Value) {
				name, de, err := di.raml.unmarshalCustomDomainExtension(di.Location, keyNode, valueNode)
				if err != nil {
					return StacktraceNewWrapped("unmarshal custom domain extension", err, di.Location, WithNodePosition(valueNode))
				}
				di.CustomDomainProperties.Set(name, de)
			} else {
				return StacktraceNew("unknown field", di.Location, stacktrace.WithInfo("field", keyNode.Value))
			}
		}
	}
	return nil
}

func (r *RAML) unmarshalDocumentationItems(node *yaml.Node, location string) ([]*DocumentationItem, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, StacktraceNew("documentation item must be a mapping node", location, WithNodePosition(node))
	}

	documentationItems := make([]*DocumentationItem, len(node.Content))
	for i, itemNode := range node.Content {
		documentationItem, err := r.unmarshalDocumentationItem(itemNode, location)
		if err != nil {
			return nil, StacktraceNewWrapped("unmarshal documentation item", err, location, WithNodePosition(itemNode))
		}
		documentationItems[i] = documentationItem
	}
	return documentationItems, nil
}

func (r *RAML) unmarshalDocumentationItem(node *yaml.Node, location string) (*DocumentationItem, error) {
	if node.Kind != yaml.MappingNode {
		return nil, StacktraceNew("documentation item must be a mapping node", location, WithNodePosition(node))
	}

	documentationItem := &DocumentationItem{
		CustomDomainProperties: orderedmap.New[string, *DomainExtension](0),

		raml:     r,
		Location: location,
	}
	if err := documentationItem.decode(node); err != nil {
		return nil, StacktraceNewWrapped("decode documentation item", err, location, WithNodePosition(node))
	}
	return documentationItem, nil
}
