// Code generated from ./rdtParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package rdt // rdtParser

import "github.com/antlr4-go/antlr/v4"

type BaserdtParserVisitor struct {
	*antlr.BaseParseTreeVisitor
}

func (v *BaserdtParserVisitor) VisitEntrypoint(ctx *EntrypointContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitExpression(ctx *ExpressionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitType(ctx *TypeContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitPrimitive(ctx *PrimitiveContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitOptional(ctx *OptionalContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitArray(ctx *ArrayContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitUnion(ctx *UnionContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitGroup(ctx *GroupContext) interface{} {
	return v.VisitChildren(ctx)
}

func (v *BaserdtParserVisitor) VisitReference(ctx *ReferenceContext) interface{} {
	return v.VisitChildren(ctx)
}
