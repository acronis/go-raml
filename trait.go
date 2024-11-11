package raml

import (
	"path/filepath"
	"strings"

	"github.com/acronis/go-stacktrace"
	"gopkg.in/yaml.v3"
)

type VariableInfo struct {
	Name      string
	Substring string
	Action    string
}

// Trait must be indexed YAML nodes with special unmarshalling logic since this trait is an Operation template.
type TraitDefinition struct {
	ID int64

	Usage string

	Source            *yaml.Node
	DeclaredVariables map[string]struct{}
	NodeVariableIndex map[int][]VariableInfo
	Precompiled       *Operation

	Link *TraitFragment

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeTraitDefinition(valueNode *yaml.Node, location string) (*TraitDefinition, error) {
	traitDef := &TraitDefinition{
		DeclaredVariables: make(map[string]struct{}),
		NodeVariableIndex: make(map[int][]VariableInfo),

		raml:     r,
		Position: stacktrace.Position{Line: valueNode.Line, Column: valueNode.Column},
		Location: location,
	}

	if err := traitDef.decode(valueNode); err != nil {
		return nil, StacktraceNewWrapped("decode trait definition", err, location, WithNodePosition(valueNode))
	}

	if err := traitDef.collectVariablesIndex(valueNode, 0); err != nil {
		return nil, StacktraceNewWrapped("collect NodeVariableIndex", err, location, WithNodePosition(valueNode))
	}

	if len(traitDef.DeclaredVariables) == 0 {
		operation, err := traitDef.compile(nil)
		if err != nil {
			return nil, StacktraceNewWrapped("compile trait definition", err, location, WithNodePosition(valueNode))
		}
		traitDef.Precompiled = operation
	}

	return traitDef, nil
}

func (t *TraitDefinition) decode(node *yaml.Node) error {
	if node.Tag == TagNull {
		return nil
	} else if node.Tag == TagInclude {
		baseDir := filepath.Dir(t.Location)
		traitFrag, err := t.raml.parseTraitFragment(filepath.Join(baseDir, node.Value))
		if err != nil {
			return StacktraceNewWrapped("parse trait fragment", err, t.Location, WithNodePosition(node))
		}
		t.Link = traitFrag
		return nil
	} else if node.Kind != yaml.MappingNode {
		return StacktraceNew("trait definition must be a mapping node", t.Location, WithNodePosition(node))
	}

	content := make([]*yaml.Node, 0, len(node.Content))
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]
		switch keyNode.Value {
		case "usage":
			if err := valueNode.Decode(&t.Usage); err != nil {
				return StacktraceNewWrapped("decode usage", err, t.Location, WithNodePosition(valueNode))
			}
		default:
			content = append(content, keyNode, valueNode)
		}
	}
	t.Source = &yaml.Node{
		Kind:    node.Kind,
		Tag:     node.Tag,
		Content: content,
		Line:    node.Line,
		Column:  node.Column,
	}
	return nil
}

func (t *TraitDefinition) collectVariablesIndex(node *yaml.Node, idx int) error {
	if node.Kind == yaml.ScalarNode {
		if err := t.findVariable(node, idx); err != nil {
			return StacktraceNewWrapped("find variable", err, t.Location, WithNodePosition(node))
		}
	}

	for i := 0; i < len(node.Content); i++ {
		if err := t.collectVariablesIndex(node.Content[i], idx+i); err != nil {
			return StacktraceNewWrapped("collect NodeVariableIndex", err, t.Location, WithNodePosition(node.Content[i]))
		}
	}
	return nil
}

func (t *TraitDefinition) findVariable(node *yaml.Node, idx int) error {
	if node.Kind != yaml.ScalarNode {
		return StacktraceNew("variable must be a scalar node", t.Location, WithNodePosition(node))
	} else if node.Tag != TagStr {
		return nil
	}

	varMatches := TemplateVariableRe.FindAllStringSubmatch(node.Value, -1)
	if len(varMatches) == 0 {
		return nil
	}
	vars := make([]VariableInfo, len(varMatches))
	for i, match := range varMatches {
		substring := match[0]
		name := match[1]
		action := match[3]
		t.DeclaredVariables[name] = struct{}{}
		vars[i] = VariableInfo{
			Name:      name,
			Substring: substring,
			Action:    action,
		}
	}
	t.NodeVariableIndex[idx] = vars
	return nil
}

func (t *TraitDefinition) compile(params map[string]string) (*Operation, error) {
	// TODO: Proper position handling
	if t.Link != nil {
		return t.Link.Trait.compile(params)
	}
	for k := range params {
		if _, ok := t.DeclaredVariables[k]; !ok {
			return nil, StacktraceNew("unexpected parameter", t.Location, WithNodePosition(t.Source), stacktrace.WithInfo("parameter", k))
		}
	}
	if t.Precompiled != nil {
		return t.Precompiled, nil
	}
	for k := range t.DeclaredVariables {
		if _, ok := params[k]; !ok {
			return nil, StacktraceNew("missing required parameter", t.Location, WithNodePosition(t.Source), stacktrace.WithInfo("parameter", k))
		}
	}
	source := t.compileSource(t.Source, params, 0)
	// TODO: Nodes parametrization must be relative to the document that uses the trait.
	// Otherwise, node content should be resolved relative to trait definition.
	// But how do we know in which context the node was inserted?
	operation, err := t.raml.makeParametrizedOperation(t.Location, source)
	if err != nil {
		return nil, StacktraceNewWrapped("make operation", err, t.Location, WithNodePosition(t.Source))
	}
	return operation, nil
}

func (t *TraitDefinition) compileSource(node *yaml.Node, params map[string]string, idx int) *yaml.Node {
	// Base case: If the node is a scalar and matches a variable in params, replace it
	if node.Kind == yaml.ScalarNode {
		variables, exists := t.NodeVariableIndex[idx]
		if !exists {
			return node
		}
		newNode := *node
		// TODO: Support template action on string
		for _, variable := range variables {
			newNode.Value = strings.Replace(newNode.Value, variable.Substring, params[variable.Name], 1)
		}
		return &newNode
	}

	// Recursive case: If the node is a map or sequence, recursively copy if needed
	modified := false
	content := make([]*yaml.Node, len(node.Content)) // Prepare a new slice for children
	for i := 0; i < len(node.Content); i++ {
		child := t.compileSource(node.Content[i], params, idx+i) // Recurse on children
		content[i] = child

		if child != node.Content[i] {
			modified = true // Mark if any child was modified
		}
	}

	if modified {
		newNode := *node
		newNode.Content = content
		return &newNode // Return a new node if any child was modified
	}
	return node // No modification, return the original node
}

type Trait struct {
	ID int64

	Name   string
	Params map[string]string

	Location string
	stacktrace.Position
	raml *RAML
}

func (r *RAML) makeTraits(valueNode *yaml.Node, location string) ([]*Trait, error) {
	switch valueNode.Kind {
	case yaml.ScalarNode:
		if valueNode.Tag == TagNull {
			return nil, nil
		}
		trait, err := r.makeTrait(valueNode, location)
		if err != nil {
			return nil, StacktraceNewWrapped("make trait", err, location, WithNodePosition(valueNode))
		}
		return []*Trait{trait}, nil
	case yaml.SequenceNode:
		traits := make([]*Trait, len(valueNode.Content))
		for i, node := range valueNode.Content {
			trait, err := r.makeTrait(node, location)
			if err != nil {
				return nil, StacktraceNewWrapped("make trait", err, location, WithNodePosition(node))
			}
			traits[i] = trait
		}
		return traits, nil
	default:
		return nil, StacktraceNew("traits must be either sequence or scalar node", location, WithNodePosition(valueNode))
	}
}

func (r *RAML) makeTrait(valueNode *yaml.Node, location string) (*Trait, error) {
	trait := &Trait{
		Location: location,
		raml:     r,
		Position: stacktrace.Position{Line: valueNode.Line, Column: valueNode.Column},
	}

	if err := trait.decode(valueNode); err != nil {
		return nil, StacktraceNewWrapped("decode trait", err, location, WithNodePosition(valueNode))
	}

	return trait, nil
}

func (t *Trait) decode(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		t.Name = node.Value
	case yaml.MappingNode:
		keyNode := node.Content[0]
		valueNode := node.Content[1]
		t.Name = keyNode.Value
		if err := valueNode.Decode(&t.Params); err != nil {
			return StacktraceNewWrapped("decode trait parameters", err, t.Location, WithNodePosition(valueNode))
		}
	default:
		return StacktraceNew("trait must be either scalar or mapping node", t.Location, WithNodePosition(node))
	}
	return nil
}
