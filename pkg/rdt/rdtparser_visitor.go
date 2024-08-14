// Code generated from ./rdtParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // rdtParser

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by rdtParser.
type rdtParserVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by rdtParser#entrypoint.
	VisitEntrypoint(ctx *EntrypointContext) interface{}

	// Visit a parse tree produced by rdtParser#expression.
	VisitExpression(ctx *ExpressionContext) interface{}

	// Visit a parse tree produced by rdtParser#type.
	VisitType(ctx *TypeContext) interface{}

	// Visit a parse tree produced by rdtParser#primitive.
	VisitPrimitive(ctx *PrimitiveContext) interface{}

	// Visit a parse tree produced by rdtParser#optional.
	VisitOptional(ctx *OptionalContext) interface{}

	// Visit a parse tree produced by rdtParser#array.
	VisitArray(ctx *ArrayContext) interface{}

	// Visit a parse tree produced by rdtParser#union.
	VisitUnion(ctx *UnionContext) interface{}

	// Visit a parse tree produced by rdtParser#group.
	VisitGroup(ctx *GroupContext) interface{}

	// Visit a parse tree produced by rdtParser#reference.
	VisitReference(ctx *ReferenceContext) interface{}
}
