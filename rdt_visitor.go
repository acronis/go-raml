package raml

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"

	"github.com/acronis/go-raml/v2/rdt"
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

func (visitor *RdtVisitor) VisitUnionMembers(node antlr.RuleNode, target *UnknownShape) ([]*BaseShape, error) {
	var shapes []*BaseShape
	children := node.GetChildren()
	// Each union member is paired with a pipe separator so we increment by 2.
	// type1 | type2 | type3
	// ^     ^ ^     ^ ^
	// 0     1 2     3 4
	for i := 0; i < len(children); i += 2 {
		baseResolved, implicitAnonShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
		s, err := visitor.Visit(children[i].(antlr.ParseTree), implicitAnonShape.(*UnknownShape))
		if err != nil {
			return nil, fmt.Errorf("visit children: %w", err)
		}
		// Replace with resolved shape
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
	// Resolve target shape into union shape since this is the base.
	shape, err := visitor.raml.MakeConcreteShapeYAML(target.Base(), TypeUnion, target.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape yaml: %w", err)
	}
	unionShape := shape.(*UnionShape)

	// Create new anonymous shape for union member and continue resolving the expression for it.
	baseResolved, anonResolvedShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
	s, err := visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), anonResolvedShape.(*UnknownShape))
	if err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}
	// Replace with resolved shape
	baseResolved.SetShape(s)

	// Nil shape is also anonymous here and doesn't share the base shape with the target.
	baseNil, _, _ := visitor.raml.MakeNewShape("", TypeNil, target.Location, &target.Position)

	unionShape.UnionFacets.AnyOf = []*BaseShape{baseResolved, baseNil}
	return unionShape, nil
}

func (visitor *RdtVisitor) VisitArray(ctx *rdt.ArrayContext, target *UnknownShape) (Shape, error) {
	// Resolve target shape into array shape since this is the base.
	shape, err := visitor.raml.MakeConcreteShapeYAML(target.Base(), TypeArray, target.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape yaml: %w", err)
	}
	arrayShape := shape.(*ArrayShape)

	// Create new anonymous shape for items and continue resolving the expression for it.
	itemsBase, itemsShape, _ := visitor.raml.MakeNewShape("", "", target.Location, &target.Position)
	itemsShape, err = visitor.Visit(ctx.GetChildren()[0].(antlr.ParseTree), itemsShape.(*UnknownShape))
	if err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}
	// Replace with resolved shape
	itemsBase.SetShape(itemsShape)

	arrayShape.ArrayFacets.Items = itemsBase
	return arrayShape, nil
}

func (visitor *RdtVisitor) VisitUnion(ctx *rdt.UnionContext, target *UnknownShape) (Shape, error) {
	// Resolve target shape into union shape since this is the base.
	shape, err := visitor.raml.MakeConcreteShapeYAML(target.Base(), TypeUnion, target.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape yaml: %w", err)
	}
	unionShape := shape.(*UnionShape)

	ss, err := visitor.VisitUnionMembers(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("visit children: %w", err)
	}

	unionShape.UnionFacets.AnyOf = ss
	return unionShape, nil
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
