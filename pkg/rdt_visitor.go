package goraml

import (
	"fmt"
	"strings"

	rdt "github.com/acronis/go-raml/pkg/rdt"
	"github.com/antlr4-go/antlr/v4"
	"gopkg.in/yaml.v3"
)

// Define a struct that implements the visitor
type RdtVisitor struct {
	rdt.BaserdtParserVisitor // Embedding the base visitor class

	Target Shape
}

func NewRdtVisitor(target Shape) *RdtVisitor {
	return &RdtVisitor{Target: target}
}

func (c *RdtVisitor) Visit(tree antlr.ParseTree) (*Shape, error) {
	switch t := tree.(type) {
	case *rdt.EntrypointContext:
		return c.VisitEntrypoint(t)
	case *rdt.ExpressionContext:
		return c.VisitExpression(t)
	case *rdt.TypeContext:
		return c.VisitType(t)
	case *rdt.PrimitiveContext:
		return c.VisitPrimitive(t)
	case *rdt.OptionalContext:
		return c.VisitOptional(t)
	case *rdt.ArrayContext:
		return c.VisitArray(t)
	case *rdt.UnionContext:
		return c.VisitUnion(t)
	case *rdt.GroupContext:
		return c.VisitGroup(t)
	case *rdt.ReferenceContext:
		return c.VisitReference(t)
	}
	return nil, fmt.Errorf("unknown node type %T", tree)
}

func (c *RdtVisitor) VisitChildren(node antlr.RuleNode) ([]*Shape, error) {
	var shapes []*Shape
	for _, n := range node.GetChildren() {
		// Skip terminal nodes
		if _, ok := n.(*antlr.TerminalNodeImpl); ok {
			continue
		}
		s, err := c.Visit(n.(antlr.ParseTree))
		if err != nil {
			return nil, err
		}
		shapes = append(shapes, s)
	}
	return shapes, nil
}

func (v *RdtVisitor) VisitEntrypoint(ctx *rdt.EntrypointContext) (*Shape, error) {
	return v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (v *RdtVisitor) VisitExpression(ctx *rdt.ExpressionContext) (*Shape, error) {
	return v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (v *RdtVisitor) VisitType(ctx *rdt.TypeContext) (*Shape, error) {
	return v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (v *RdtVisitor) VisitPrimitive(ctx *rdt.PrimitiveContext) (*Shape, error) {
	base := v.Target.Base()
	s, err := MakeConcreteShape(base, ctx.GetText(), make([]*yaml.Node, 0))
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (v *RdtVisitor) VisitOptional(ctx *rdt.OptionalContext) (*Shape, error) {
	s, err := v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
	if err != nil {
		return nil, err
	}
	nilShape, _ := MakeConcreteShape(v.Target.Base(), "nil", make([]*yaml.Node, 0))
	base := *v.Target.Base()
	base.Type = UNION
	var unionShape Shape = &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: []*Shape{s, &nilShape},
		},
	}
	return &unionShape, nil
}

func (v *RdtVisitor) VisitArray(ctx *rdt.ArrayContext) (*Shape, error) {
	s, err := v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
	if err != nil {
		return nil, err
	}
	base := *v.Target.Base()
	base.Type = ARRAY
	var arrayShape Shape = &ArrayShape{
		BaseShape: base,
		ArrayFacets: ArrayFacets{
			Items: s,
		},
	}
	return &arrayShape, nil
}

func (v *RdtVisitor) VisitUnion(ctx *rdt.UnionContext) (*Shape, error) {
	ss, err := v.VisitChildren(ctx)
	if err != nil {
		return nil, err
	}
	base := *v.Target.Base()
	base.Type = UNION
	var unionShape Shape = &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: ss,
		},
	}
	return &unionShape, nil
}

func (v *RdtVisitor) VisitGroup(ctx *rdt.GroupContext) (*Shape, error) {
	return v.Visit(ctx.GetChildren()[0].(antlr.ParseTree))
}

func (v *RdtVisitor) VisitReference(ctx *rdt.ReferenceContext) (*Shape, error) {
	frag := GetRegistry().GetFragment(v.Target.Base().Location).(*Library)

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
	s, err := MakeConcreteShape(v.Target.Base(), (*ref).Base().Type, v.Target.(*UnknownShape).facets)
	if err != nil {
		return nil, err
	}
	s.Base().Inherits = append(v.Target.Base().Inherits, ref)
	return &s, nil
}
