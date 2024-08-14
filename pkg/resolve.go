package goraml

import (
	rdt "github.com/acronis/go-raml/pkg/rdt"
	"github.com/antlr4-go/antlr/v4"
)

func ResolveShapes() error {
	for _, shape := range GetRegistry().UnresolvedShapes {
		if err := Resolve(shape); err != nil {
			return err
		}
	}

	return nil
}

func ResolveMultipleInheritance(target Shape) (*Shape, error) {
	inherits := target.Base().Inherits
	for _, inherit := range inherits {
		if err := Resolve(inherit); err != nil {
			return nil, err
		}
	}
	// Multiple inheritance validation to be performed in a separate validation stage
	s, err := MakeConcreteShape(target.Base(), (*inherits[0]).Base().Type, target.(*UnknownShape).facets)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func ResolveLink(target Shape) (*Shape, error) {
	link := target.Base().Link
	if err := Resolve(link.Shape); err != nil {
		return nil, err
	}
	s, err := MakeConcreteShape(target.Base(), (*link.Shape).Base().Type, target.(*UnknownShape).facets)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Resolves an unknown shape in-place.
func Resolve(shape *Shape) error {
	target := *shape
	// Skip already resolved and JSON shapes
	if _, ok := target.(*UnknownShape); !ok {
		return nil
	} else if _, ok := target.(*JSONShape); ok {
		return nil
	}

	link := target.Base().Link
	if link != nil {
		s, err := ResolveLink(target)
		if err != nil {
			return err
		}
		*shape = *s
		GetRegistry().ResolvedShapes = append(GetRegistry().ResolvedShapes, shape)
		return nil
	}

	shapeType := target.Base().Type
	if shapeType == COMPOSITE {
		// Special case for multiple inheritance
		s, err := ResolveMultipleInheritance(target)
		if err != nil {
			return err
		}
		*shape = *s
		GetRegistry().ResolvedShapes = append(GetRegistry().ResolvedShapes, shape)
		return nil
	}

	is := antlr.NewInputStream(shapeType)
	lexer := rdt.NewrdtLexer(is)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	rdtParser := rdt.NewrdtParser(tokens)
	visitor := NewRdtVisitor(target)
	tree := rdtParser.Entrypoint()

	s, err := visitor.Visit(tree)
	if err != nil {
		return err
	}
	*shape = *s
	GetRegistry().ResolvedShapes = append(GetRegistry().ResolvedShapes, shape)
	return nil
}
