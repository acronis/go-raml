package raml

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"gopkg.in/yaml.v3"

	"github.com/acronis/go-raml/rdt"
)

// RdtVisitor defines a struct that implements the visitor
type RdtVisitor struct {
	rdt.BaserdtParserVisitor // Embedding the base visitor class

	Target Shape
}

func NewRdtVisitor(target Shape) *RdtVisitor {
	return &RdtVisitor{Target: target}
}

func (visitor *RdtVisitor) Visit(tree antlr.ParseTree) (*Shape, error) {
	switch t := tree.(type) {
	case *rdt.EntrypointContext:
		return visitor.VisitEntrypoint(t)
	case *rdt.ExpressionContext:
		return visitor.VisitExpression(t)
	case *rdt.TypeContext:
		return visitor.VisitType(t)
	case *rdt.PrimitiveContext:
		return visitor.VisitPrimitive(t)
	case *rdt.OptionalContext:
		return visitor.VisitOptional(t)
	case *rdt.ArrayContext:
		return visitor.VisitArray(t)
	case *rdt.UnionContext:
		return visitor.VisitUnion(t)
	case *rdt.GroupContext:
		return visitor.VisitGroup(t)
	case *rdt.ReferenceContext:
		return visitor.VisitReference(t)
	}
	return nil, fmt.Errorf("unknown node type %T", tree)
}

func (visitor *RdtVisitor) VisitChildren(node antlr.RuleNode) ([]*Shape, error) {
	var shapes []*Shape
	for _, n := range node.GetChildren() {
		// Skip terminal nodes
		if _, ok := n.(*antlr.TerminalNodeImpl); ok {
			continue
		}
		s, err := visitor.Visit(n.(antlr.ParseTree))
		if err != nil {
			return nil, err
		}
		shapes = append(shapes, s)
	}
	return shapes, nil
}

func (visitor *RdtVisitor) VisitEntrypoint(ctx *rdt.EntrypointContext) (*Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (visitor *RdtVisitor) VisitExpression(ctx *rdt.ExpressionContext) (*Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (visitor *RdtVisitor) VisitType(ctx *rdt.TypeContext) (*Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (visitor *RdtVisitor) VisitPrimitive(ctx *rdt.PrimitiveContext) (*Shape, error) {
	base := visitor.Target.Base()
	s, err := MakeConcreteShape(base, ctx.GetText(), make([]*yaml.Node, 0))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (visitor *RdtVisitor) VisitOptional(ctx *rdt.OptionalContext) (*Shape, error) {
	s, err := visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
	if err != nil {
		return nil, err
	}
	nilShape, _ := MakeConcreteShape(visitor.Target.Base(), "nil", make([]*yaml.Node, 0))
	base := *visitor.Target.Base()
	base.Type = TypeUnion
	var unionShape Shape = &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: []*Shape{s, &nilShape},
		},
	}
	return &unionShape, nil
}

func (visitor *RdtVisitor) VisitArray(ctx *rdt.ArrayContext) (*Shape, error) {
	s, err := visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
	if err != nil {
		return nil, err
	}
	base := *visitor.Target.Base()
	base.Type = TypeArray
	var arrayShape Shape = &ArrayShape{
		BaseShape: base,
		ArrayFacets: ArrayFacets{
			Items: s,
		},
	}
	return &arrayShape, nil
}

func (visitor *RdtVisitor) VisitUnion(ctx *rdt.UnionContext) (*Shape, error) {
	ss, err := visitor.VisitChildren(ctx)
	if err != nil {
		return nil, err
	}
	base := *visitor.Target.Base()
	base.Type = TypeUnion
	var unionShape Shape = &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: ss,
		},
	}
	return &unionShape, nil
}

func (visitor *RdtVisitor) VisitGroup(ctx *rdt.GroupContext) (*Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (visitor *RdtVisitor) VisitReference(ctx *rdt.ReferenceContext) (*Shape, error) {
	frag := GetRegistry().GetFragment(visitor.Target.Base().Location).(*Library)

	// External ref - lib.Type
	// Internal ref - Type
	shapeType := ctx.GetText()
	parts := strings.Split(shapeType, ".")
	var ref *Shape
	if len(parts) == 1 {
		ref = frag.Types[parts[0]]
		if ref == nil {
			return nil, fmt.Errorf("reference %s not found", parts[0])
		}
	} else if len(parts) == 2 {
		lib := frag.Uses[parts[0]]
		if lib == nil {
			return nil, fmt.Errorf("library %s not found", parts[0])
		}
		ref = lib.Types[parts[1]]
		if ref == nil {
			return nil, fmt.Errorf("reference %s not found", parts[1])
		}
	} else {
		return nil, fmt.Errorf("invalid reference %s", shapeType)
	}
	if err := Resolve(ref); err != nil {
		return nil, err
	}
	s, err := MakeConcreteShape(visitor.Target.Base(), (*ref).Base().Type, visitor.Target.(*UnknownShape).facets)
	if err != nil {
		return nil, err
	}
	s.Base().Inherits = append(visitor.Target.Base().Inherits, ref)
	return &s, nil
}
