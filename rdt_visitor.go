package raml

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"

	"github.com/acronis/go-raml/rdt"
)

// RdtVisitor defines a struct that implements the visitor
type RdtVisitor struct {
	rdt.BaserdtParserVisitor // Embedding the base visitor class
	raml                     *RAML
}

func NewRdtVisitor(rml *RAML) *RdtVisitor {
	return &RdtVisitor{raml: rml}
}

func (visitor *RdtVisitor) Visit(tree antlr.ParseTree, target *UnknownShape) (Shape, error) {
	// Target is required to isolate anonymous shapes created by Union, Optional and Array syntax.
	// This is done to avoid sharing base shape properties between the original type and implicitly created type.
	switch t := tree.(type) {
	case *rdt.EntrypointContext:
		return visitor.VisitEntrypoint(t, target)
	case *rdt.ExpressionContext:
		return visitor.VisitExpression(t, target)
	case *rdt.TypeContext:
		return visitor.VisitType(t, target)
	case *rdt.PrimitiveContext:
		return visitor.VisitPrimitive(t, target)
	case *rdt.OptionalContext:
		return visitor.VisitOptional(t, target)
	case *rdt.ArrayContext:
		return visitor.VisitArray(t, target)
	case *rdt.UnionContext:
		return visitor.VisitUnion(t, target)
	case *rdt.GroupContext:
		return visitor.VisitGroup(t, target)
	case *rdt.ReferenceContext:
		return visitor.VisitReference(t, target)
	}
	return nil, fmt.Errorf("unknown node type %T", tree)
}

func (visitor *RdtVisitor) VisitChildren(node antlr.RuleNode, target *UnknownShape) ([]*BaseShape, error) {
	var shapes []*BaseShape
	for _, n := range node.GetChildren() {
		// Skip terminal nodes
		if _, ok := n.(*antlr.TerminalNodeImpl); ok {
			continue
		}
		baseResolved, implicitAnonShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
		s, err := visitor.Visit(n.(antlr.ParseTree), implicitAnonShape.(*UnknownShape))
		if err != nil {
			return nil, fmt.Errorf("visit children: %w", err)
		}
		baseResolved.SetShape(s)
		shapes = append(shapes, baseResolved)
	}
	return shapes, nil
}

func (visitor *RdtVisitor) VisitEntrypoint(ctx *rdt.EntrypointContext, target *UnknownShape) (Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), target)
}

func (visitor *RdtVisitor) VisitExpression(ctx *rdt.ExpressionContext, target *UnknownShape) (Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), target)
}

func (visitor *RdtVisitor) VisitType(ctx *rdt.TypeContext, target *UnknownShape) (Shape, error) {
	return visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), target)
}

func (visitor *RdtVisitor) VisitPrimitive(ctx *rdt.PrimitiveContext, target *UnknownShape) (Shape, error) {
	s, err := visitor.raml.MakeConcreteShapeYAML(target.Base(), ctx.GetText(), nil)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return s, nil
}

func (visitor *RdtVisitor) VisitOptional(ctx *rdt.OptionalContext, target *UnknownShape) (Shape, error) {
	// Passed target shape becomes anonymous here because union shape takes its place later.
	baseResolved, anonResolvedShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
	s, err := visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), anonResolvedShape.(*UnknownShape))
	if err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}
	// Replace with resolved shape
	baseResolved.SetShape(s)

	// Nil shape is also anonymous here and doesn't share the base shape with the target.
	baseNil, _, _ := visitor.raml.MakeNewShape("", TypeNil, target.Location, &target.Position)

	// We transfer base to new shape
	// TODO: Need some kind of conversion interface.
	base := target.Base()
	base.Type = TypeUnion
	return &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: []*BaseShape{baseResolved, baseNil},
		},
	}, nil
}

func (visitor *RdtVisitor) VisitArray(ctx *rdt.ArrayContext, target *UnknownShape) (Shape, error) {
	// Passed target shape becomes anonymous here because union shape takes its place later.
	baseResolved, anonResolvedShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
	s, err := visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), anonResolvedShape.(*UnknownShape))
	if err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}
	// Replace with resolved shape
	baseResolved.SetShape(s)

	// We transfer base to new shape
	// TODO: Need some kind of conversion interface.
	base := target.Base()
	base.Type = TypeArray
	return &ArrayShape{
		BaseShape: base,
		ArrayFacets: ArrayFacets{
			Items: baseResolved,
		},
	}, nil
}

func (visitor *RdtVisitor) VisitUnion(ctx *rdt.UnionContext, target *UnknownShape) (Shape, error) {
	ss, err := visitor.VisitChildren(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("visit children: %w", err)
	}
	base := target.Base()
	base.Type = TypeUnion
	return &UnionShape{
		BaseShape: base,
		UnionFacets: UnionFacets{
			AnyOf: ss,
		},
	}, nil
}

func (visitor *RdtVisitor) VisitGroup(ctx *rdt.GroupContext, target *UnknownShape) (Shape, error) {
	// First and last nodes are terminal nodes
	// ( expression )
	// ^     ^      ^
	// 0     1      2
	return visitor.Visit(ctx.GetChildren()[1].(antlr.ParseTree), target)
}

func (visitor *RdtVisitor) VisitReference(ctx *rdt.ReferenceContext, target *UnknownShape) (Shape, error) {
	shapeType := ctx.GetText()
	ref, err := visitor.raml.GetReferencedType(shapeType, target.Location)
	if err != nil {
		return nil, fmt.Errorf("get referenced shape: %w", err)
	}
	if ref.ID == target.ID {
		return nil, fmt.Errorf("self recursion %s", shapeType)
	}
	if errResolveShape := visitor.raml.resolveShape(ref); errResolveShape != nil {
		return nil, fmt.Errorf("resolve: %w", errResolveShape)
	}
	s, err := visitor.raml.MakeConcreteShapeYAML(target.Base(), ref.Type, target.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	// If target.facets is nil (makeNewShapeYAML returned nil instead of empty array) then reference is an alias.
	s.Base().TypeLabel = shapeType
	if target.facets == nil {
		s.Base().Alias = ref
	} else {
		s.Base().Inherits = append(s.Base().Inherits, ref)
	}
	return s, nil
}
