// Code generated from ./rdtParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package rdt // rdtParser

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type rdtParser struct {
	*antlr.BaseParser
}

var RdtParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func rdtparserParserInit() {
	staticData := &RdtParserParserStaticData
	staticData.LiteralNames = []string{
		"", "'('", "')'", "'|'", "'[]'", "'?'", "'.'", "'string'", "'integer'",
		"'number'", "'boolean'", "'datetime'", "'time-only'", "'datetime-only'",
		"'date-only'", "'file'", "'nil'", "'any'", "'array'", "'object'", "'union'",
	}
	staticData.SymbolicNames = []string{
		"", "LPAREN", "RPAREN", "PIPE", "ARRAY_NOTATION", "OPTIONAL_NOTATION",
		"DOT", "STRING_TYPE", "INTEGER_TYPE", "NUMBER_TYPE", "BOOLEAN_TYPE",
		"DATETIME_TYPE", "TIME_ONLY_TYPE", "DATETIME_ONLY_TYPE", "DATE_ONLY_TYPE",
		"FILE_TYPE", "NIL_TYPE", "ANY_TYPE", "ARRAY_TYPE", "OBJECT_TYPE", "UNION_TYPE",
		"IDENTIFIER", "WS",
	}
	staticData.RuleNames = []string{
		"entrypoint", "expression", "type", "primitive", "optional", "array",
		"union", "group", "reference",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 22, 95, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 1, 0, 1, 0, 1, 0, 1,
		1, 1, 1, 3, 1, 24, 8, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 31, 8, 2,
		1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 3, 4, 38, 8, 4, 1, 4, 1, 4, 1, 5, 1, 5, 1,
		5, 3, 5, 45, 8, 5, 1, 5, 1, 5, 1, 6, 1, 6, 5, 6, 51, 8, 6, 10, 6, 12, 6,
		54, 9, 6, 1, 6, 1, 6, 5, 6, 58, 8, 6, 10, 6, 12, 6, 61, 9, 6, 1, 6, 1,
		6, 5, 6, 65, 8, 6, 10, 6, 12, 6, 68, 9, 6, 4, 6, 70, 8, 6, 11, 6, 12, 6,
		71, 1, 7, 1, 7, 5, 7, 76, 8, 7, 10, 7, 12, 7, 79, 9, 7, 1, 7, 1, 7, 5,
		7, 83, 8, 7, 10, 7, 12, 7, 86, 9, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 3, 8,
		93, 8, 8, 1, 8, 0, 0, 9, 0, 2, 4, 6, 8, 10, 12, 14, 16, 0, 1, 1, 0, 7,
		20, 101, 0, 18, 1, 0, 0, 0, 2, 23, 1, 0, 0, 0, 4, 30, 1, 0, 0, 0, 6, 32,
		1, 0, 0, 0, 8, 37, 1, 0, 0, 0, 10, 44, 1, 0, 0, 0, 12, 48, 1, 0, 0, 0,
		14, 73, 1, 0, 0, 0, 16, 89, 1, 0, 0, 0, 18, 19, 3, 2, 1, 0, 19, 20, 5,
		0, 0, 1, 20, 1, 1, 0, 0, 0, 21, 24, 3, 4, 2, 0, 22, 24, 3, 12, 6, 0, 23,
		21, 1, 0, 0, 0, 23, 22, 1, 0, 0, 0, 24, 3, 1, 0, 0, 0, 25, 31, 3, 6, 3,
		0, 26, 31, 3, 14, 7, 0, 27, 31, 3, 16, 8, 0, 28, 31, 3, 10, 5, 0, 29, 31,
		3, 8, 4, 0, 30, 25, 1, 0, 0, 0, 30, 26, 1, 0, 0, 0, 30, 27, 1, 0, 0, 0,
		30, 28, 1, 0, 0, 0, 30, 29, 1, 0, 0, 0, 31, 5, 1, 0, 0, 0, 32, 33, 7, 0,
		0, 0, 33, 7, 1, 0, 0, 0, 34, 38, 3, 6, 3, 0, 35, 38, 3, 14, 7, 0, 36, 38,
		3, 16, 8, 0, 37, 34, 1, 0, 0, 0, 37, 35, 1, 0, 0, 0, 37, 36, 1, 0, 0, 0,
		38, 39, 1, 0, 0, 0, 39, 40, 5, 5, 0, 0, 40, 9, 1, 0, 0, 0, 41, 45, 3, 6,
		3, 0, 42, 45, 3, 14, 7, 0, 43, 45, 3, 16, 8, 0, 44, 41, 1, 0, 0, 0, 44,
		42, 1, 0, 0, 0, 44, 43, 1, 0, 0, 0, 45, 46, 1, 0, 0, 0, 46, 47, 5, 4, 0,
		0, 47, 11, 1, 0, 0, 0, 48, 52, 3, 4, 2, 0, 49, 51, 5, 22, 0, 0, 50, 49,
		1, 0, 0, 0, 51, 54, 1, 0, 0, 0, 52, 50, 1, 0, 0, 0, 52, 53, 1, 0, 0, 0,
		53, 69, 1, 0, 0, 0, 54, 52, 1, 0, 0, 0, 55, 59, 5, 3, 0, 0, 56, 58, 5,
		22, 0, 0, 57, 56, 1, 0, 0, 0, 58, 61, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 59,
		60, 1, 0, 0, 0, 60, 62, 1, 0, 0, 0, 61, 59, 1, 0, 0, 0, 62, 66, 3, 4, 2,
		0, 63, 65, 5, 22, 0, 0, 64, 63, 1, 0, 0, 0, 65, 68, 1, 0, 0, 0, 66, 64,
		1, 0, 0, 0, 66, 67, 1, 0, 0, 0, 67, 70, 1, 0, 0, 0, 68, 66, 1, 0, 0, 0,
		69, 55, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 69, 1, 0, 0, 0, 71, 72, 1,
		0, 0, 0, 72, 13, 1, 0, 0, 0, 73, 77, 5, 1, 0, 0, 74, 76, 5, 22, 0, 0, 75,
		74, 1, 0, 0, 0, 76, 79, 1, 0, 0, 0, 77, 75, 1, 0, 0, 0, 77, 78, 1, 0, 0,
		0, 78, 80, 1, 0, 0, 0, 79, 77, 1, 0, 0, 0, 80, 84, 3, 2, 1, 0, 81, 83,
		5, 22, 0, 0, 82, 81, 1, 0, 0, 0, 83, 86, 1, 0, 0, 0, 84, 82, 1, 0, 0, 0,
		84, 85, 1, 0, 0, 0, 85, 87, 1, 0, 0, 0, 86, 84, 1, 0, 0, 0, 87, 88, 5,
		2, 0, 0, 88, 15, 1, 0, 0, 0, 89, 92, 5, 21, 0, 0, 90, 91, 5, 6, 0, 0, 91,
		93, 5, 21, 0, 0, 92, 90, 1, 0, 0, 0, 92, 93, 1, 0, 0, 0, 93, 17, 1, 0,
		0, 0, 11, 23, 30, 37, 44, 52, 59, 66, 71, 77, 84, 92,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// rdtParserInit initializes any static state used to implement rdtParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewrdtParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func RdtParserInit() {
	staticData := &RdtParserParserStaticData
	staticData.once.Do(rdtparserParserInit)
}

// NewrdtParser produces a new parser instance for the optional input antlr.TokenStream.
func NewrdtParser(input antlr.TokenStream) *rdtParser {
	RdtParserInit()
	this := new(rdtParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &RdtParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "rdtParser.g4"

	return this
}

// rdtParser tokens.
const (
	rdtParserEOF                = antlr.TokenEOF
	rdtParserLPAREN             = 1
	rdtParserRPAREN             = 2
	rdtParserPIPE               = 3
	rdtParserARRAY_NOTATION     = 4
	rdtParserOPTIONAL_NOTATION  = 5
	rdtParserDOT                = 6
	rdtParserSTRING_TYPE        = 7
	rdtParserINTEGER_TYPE       = 8
	rdtParserNUMBER_TYPE        = 9
	rdtParserBOOLEAN_TYPE       = 10
	rdtParserDATETIME_TYPE      = 11
	rdtParserTIME_ONLY_TYPE     = 12
	rdtParserDATETIME_ONLY_TYPE = 13
	rdtParserDATE_ONLY_TYPE     = 14
	rdtParserFILE_TYPE          = 15
	rdtParserNIL_TYPE           = 16
	rdtParserANY_TYPE           = 17
	rdtParserARRAY_TYPE         = 18
	rdtParserOBJECT_TYPE        = 19
	rdtParserUNION_TYPE         = 20
	rdtParserIDENTIFIER         = 21
	rdtParserWS                 = 22
)

// rdtParser rules.
const (
	rdtParserRULE_entrypoint = 0
	rdtParserRULE_expression = 1
	rdtParserRULE_type       = 2
	rdtParserRULE_primitive  = 3
	rdtParserRULE_optional   = 4
	rdtParserRULE_array      = 5
	rdtParserRULE_union      = 6
	rdtParserRULE_group      = 7
	rdtParserRULE_reference  = 8
)

// IEntrypointContext is an interface to support dynamic dispatch.
type IEntrypointContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Expression() IExpressionContext
	EOF() antlr.TerminalNode

	// IsEntrypointContext differentiates from other interfaces.
	IsEntrypointContext()
}

type EntrypointContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEntrypointContext() *EntrypointContext {
	var p = new(EntrypointContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_entrypoint
	return p
}

func InitEmptyEntrypointContext(p *EntrypointContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_entrypoint
}

func (*EntrypointContext) IsEntrypointContext() {}

func NewEntrypointContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EntrypointContext {
	var p = new(EntrypointContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_entrypoint

	return p
}

func (s *EntrypointContext) GetParser() antlr.Parser { return s.parser }

func (s *EntrypointContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *EntrypointContext) EOF() antlr.TerminalNode {
	return s.GetToken(rdtParserEOF, 0)
}

func (s *EntrypointContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EntrypointContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EntrypointContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitEntrypoint(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Entrypoint() (localctx IEntrypointContext) {
	localctx = NewEntrypointContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, rdtParserRULE_entrypoint)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(18)
		p.Expression()
	}
	{
		p.SetState(19)
		p.Match(rdtParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Type_() ITypeContext
	Union() IUnionContext

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) Type_() ITypeContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeContext)
}

func (s *ExpressionContext) Union() IUnionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IUnionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IUnionContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, rdtParserRULE_expression)
	p.SetState(23)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(21)
			p.Type_()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(22)
			p.Union()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ITypeContext is an interface to support dynamic dispatch.
type ITypeContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	Primitive() IPrimitiveContext
	Group() IGroupContext
	Reference() IReferenceContext
	Array() IArrayContext
	Optional() IOptionalContext

	// IsTypeContext differentiates from other interfaces.
	IsTypeContext()
}

type TypeContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTypeContext() *TypeContext {
	var p = new(TypeContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_type
	return p
}

func InitEmptyTypeContext(p *TypeContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_type
}

func (*TypeContext) IsTypeContext() {}

func NewTypeContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TypeContext {
	var p = new(TypeContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_type

	return p
}

func (s *TypeContext) GetParser() antlr.Parser { return s.parser }

func (s *TypeContext) Primitive() IPrimitiveContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimitiveContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimitiveContext)
}

func (s *TypeContext) Group() IGroupContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGroupContext)
}

func (s *TypeContext) Reference() IReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReferenceContext)
}

func (s *TypeContext) Array() IArrayContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayContext)
}

func (s *TypeContext) Optional() IOptionalContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOptionalContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOptionalContext)
}

func (s *TypeContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TypeContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TypeContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitType(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Type_() (localctx ITypeContext) {
	localctx = NewTypeContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, rdtParserRULE_type)
	p.SetState(30)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(25)
			p.Primitive()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(26)
			p.Group()
		}

	case 3:
		p.EnterOuterAlt(localctx, 3)
		{
			p.SetState(27)
			p.Reference()
		}

	case 4:
		p.EnterOuterAlt(localctx, 4)
		{
			p.SetState(28)
			p.Array()
		}

	case 5:
		p.EnterOuterAlt(localctx, 5)
		{
			p.SetState(29)
			p.Optional()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPrimitiveContext is an interface to support dynamic dispatch.
type IPrimitiveContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING_TYPE() antlr.TerminalNode
	INTEGER_TYPE() antlr.TerminalNode
	NUMBER_TYPE() antlr.TerminalNode
	BOOLEAN_TYPE() antlr.TerminalNode
	DATETIME_TYPE() antlr.TerminalNode
	TIME_ONLY_TYPE() antlr.TerminalNode
	DATETIME_ONLY_TYPE() antlr.TerminalNode
	DATE_ONLY_TYPE() antlr.TerminalNode
	FILE_TYPE() antlr.TerminalNode
	NIL_TYPE() antlr.TerminalNode
	ANY_TYPE() antlr.TerminalNode
	ARRAY_TYPE() antlr.TerminalNode
	OBJECT_TYPE() antlr.TerminalNode
	UNION_TYPE() antlr.TerminalNode

	// IsPrimitiveContext differentiates from other interfaces.
	IsPrimitiveContext()
}

type PrimitiveContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPrimitiveContext() *PrimitiveContext {
	var p = new(PrimitiveContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_primitive
	return p
}

func InitEmptyPrimitiveContext(p *PrimitiveContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_primitive
}

func (*PrimitiveContext) IsPrimitiveContext() {}

func NewPrimitiveContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PrimitiveContext {
	var p = new(PrimitiveContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_primitive

	return p
}

func (s *PrimitiveContext) GetParser() antlr.Parser { return s.parser }

func (s *PrimitiveContext) STRING_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserSTRING_TYPE, 0)
}

func (s *PrimitiveContext) INTEGER_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserINTEGER_TYPE, 0)
}

func (s *PrimitiveContext) NUMBER_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserNUMBER_TYPE, 0)
}

func (s *PrimitiveContext) BOOLEAN_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserBOOLEAN_TYPE, 0)
}

func (s *PrimitiveContext) DATETIME_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserDATETIME_TYPE, 0)
}

func (s *PrimitiveContext) TIME_ONLY_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserTIME_ONLY_TYPE, 0)
}

func (s *PrimitiveContext) DATETIME_ONLY_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserDATETIME_ONLY_TYPE, 0)
}

func (s *PrimitiveContext) DATE_ONLY_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserDATE_ONLY_TYPE, 0)
}

func (s *PrimitiveContext) FILE_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserFILE_TYPE, 0)
}

func (s *PrimitiveContext) NIL_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserNIL_TYPE, 0)
}

func (s *PrimitiveContext) ANY_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserANY_TYPE, 0)
}

func (s *PrimitiveContext) ARRAY_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserARRAY_TYPE, 0)
}

func (s *PrimitiveContext) OBJECT_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserOBJECT_TYPE, 0)
}

func (s *PrimitiveContext) UNION_TYPE() antlr.TerminalNode {
	return s.GetToken(rdtParserUNION_TYPE, 0)
}

func (s *PrimitiveContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PrimitiveContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PrimitiveContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitPrimitive(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Primitive() (localctx IPrimitiveContext) {
	localctx = NewPrimitiveContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, rdtParserRULE_primitive)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(32)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&2097024) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOptionalContext is an interface to support dynamic dispatch.
type IOptionalContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	OPTIONAL_NOTATION() antlr.TerminalNode
	Primitive() IPrimitiveContext
	Group() IGroupContext
	Reference() IReferenceContext

	// IsOptionalContext differentiates from other interfaces.
	IsOptionalContext()
}

type OptionalContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOptionalContext() *OptionalContext {
	var p = new(OptionalContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_optional
	return p
}

func InitEmptyOptionalContext(p *OptionalContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_optional
}

func (*OptionalContext) IsOptionalContext() {}

func NewOptionalContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OptionalContext {
	var p = new(OptionalContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_optional

	return p
}

func (s *OptionalContext) GetParser() antlr.Parser { return s.parser }

func (s *OptionalContext) OPTIONAL_NOTATION() antlr.TerminalNode {
	return s.GetToken(rdtParserOPTIONAL_NOTATION, 0)
}

func (s *OptionalContext) Primitive() IPrimitiveContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimitiveContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimitiveContext)
}

func (s *OptionalContext) Group() IGroupContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGroupContext)
}

func (s *OptionalContext) Reference() IReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReferenceContext)
}

func (s *OptionalContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OptionalContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OptionalContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitOptional(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Optional() (localctx IOptionalContext) {
	localctx = NewOptionalContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, rdtParserRULE_optional)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(37)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case rdtParserSTRING_TYPE, rdtParserINTEGER_TYPE, rdtParserNUMBER_TYPE, rdtParserBOOLEAN_TYPE, rdtParserDATETIME_TYPE, rdtParserTIME_ONLY_TYPE, rdtParserDATETIME_ONLY_TYPE, rdtParserDATE_ONLY_TYPE, rdtParserFILE_TYPE, rdtParserNIL_TYPE, rdtParserANY_TYPE, rdtParserARRAY_TYPE, rdtParserOBJECT_TYPE, rdtParserUNION_TYPE:
		{
			p.SetState(34)
			p.Primitive()
		}

	case rdtParserLPAREN:
		{
			p.SetState(35)
			p.Group()
		}

	case rdtParserIDENTIFIER:
		{
			p.SetState(36)
			p.Reference()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(39)
		p.Match(rdtParserOPTIONAL_NOTATION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayContext is an interface to support dynamic dispatch.
type IArrayContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	ARRAY_NOTATION() antlr.TerminalNode
	Primitive() IPrimitiveContext
	Group() IGroupContext
	Reference() IReferenceContext

	// IsArrayContext differentiates from other interfaces.
	IsArrayContext()
}

type ArrayContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayContext() *ArrayContext {
	var p = new(ArrayContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_array
	return p
}

func InitEmptyArrayContext(p *ArrayContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_array
}

func (*ArrayContext) IsArrayContext() {}

func NewArrayContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayContext {
	var p = new(ArrayContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_array

	return p
}

func (s *ArrayContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayContext) ARRAY_NOTATION() antlr.TerminalNode {
	return s.GetToken(rdtParserARRAY_NOTATION, 0)
}

func (s *ArrayContext) Primitive() IPrimitiveContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPrimitiveContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IPrimitiveContext)
}

func (s *ArrayContext) Group() IGroupContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IGroupContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IGroupContext)
}

func (s *ArrayContext) Reference() IReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReferenceContext)
}

func (s *ArrayContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitArray(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Array() (localctx IArrayContext) {
	localctx = NewArrayContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, rdtParserRULE_array)
	p.EnterOuterAlt(localctx, 1)
	p.SetState(44)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case rdtParserSTRING_TYPE, rdtParserINTEGER_TYPE, rdtParserNUMBER_TYPE, rdtParserBOOLEAN_TYPE, rdtParserDATETIME_TYPE, rdtParserTIME_ONLY_TYPE, rdtParserDATETIME_ONLY_TYPE, rdtParserDATE_ONLY_TYPE, rdtParserFILE_TYPE, rdtParserNIL_TYPE, rdtParserANY_TYPE, rdtParserARRAY_TYPE, rdtParserOBJECT_TYPE, rdtParserUNION_TYPE:
		{
			p.SetState(41)
			p.Primitive()
		}

	case rdtParserLPAREN:
		{
			p.SetState(42)
			p.Group()
		}

	case rdtParserIDENTIFIER:
		{
			p.SetState(43)
			p.Reference()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	{
		p.SetState(46)
		p.Match(rdtParserARRAY_NOTATION)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IUnionContext is an interface to support dynamic dispatch.
type IUnionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllType_() []ITypeContext
	Type_(i int) ITypeContext
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode
	AllPIPE() []antlr.TerminalNode
	PIPE(i int) antlr.TerminalNode

	// IsUnionContext differentiates from other interfaces.
	IsUnionContext()
}

type UnionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyUnionContext() *UnionContext {
	var p = new(UnionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_union
	return p
}

func InitEmptyUnionContext(p *UnionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_union
}

func (*UnionContext) IsUnionContext() {}

func NewUnionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *UnionContext {
	var p = new(UnionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_union

	return p
}

func (s *UnionContext) GetParser() antlr.Parser { return s.parser }

func (s *UnionContext) AllType_() []ITypeContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITypeContext); ok {
			len++
		}
	}

	tst := make([]ITypeContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITypeContext); ok {
			tst[i] = t.(ITypeContext)
			i++
		}
	}

	return tst
}

func (s *UnionContext) Type_(i int) ITypeContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITypeContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ITypeContext)
}

func (s *UnionContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(rdtParserWS)
}

func (s *UnionContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(rdtParserWS, i)
}

func (s *UnionContext) AllPIPE() []antlr.TerminalNode {
	return s.GetTokens(rdtParserPIPE)
}

func (s *UnionContext) PIPE(i int) antlr.TerminalNode {
	return s.GetToken(rdtParserPIPE, i)
}

func (s *UnionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *UnionContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitUnion(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Union() (localctx IUnionContext) {
	localctx = NewUnionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, rdtParserRULE_union)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(48)
		p.Type_()
	}
	p.SetState(52)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == rdtParserWS {
		{
			p.SetState(49)
			p.Match(rdtParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(54)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	p.SetState(69)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == rdtParserPIPE {
		{
			p.SetState(55)
			p.Match(rdtParserPIPE)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(59)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		for _la == rdtParserWS {
			{
				p.SetState(56)
				p.Match(rdtParserWS)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

			p.SetState(61)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)
		}
		{
			p.SetState(62)
			p.Type_()
		}
		p.SetState(66)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
		for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
			if _alt == 1 {
				{
					p.SetState(63)
					p.Match(rdtParserWS)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			}
			p.SetState(68)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
			if p.HasError() {
				goto errorExit
			}
		}

		p.SetState(71)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IGroupContext is an interface to support dynamic dispatch.
type IGroupContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	Expression() IExpressionContext
	RPAREN() antlr.TerminalNode
	AllWS() []antlr.TerminalNode
	WS(i int) antlr.TerminalNode

	// IsGroupContext differentiates from other interfaces.
	IsGroupContext()
}

type GroupContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyGroupContext() *GroupContext {
	var p = new(GroupContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_group
	return p
}

func InitEmptyGroupContext(p *GroupContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_group
}

func (*GroupContext) IsGroupContext() {}

func NewGroupContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *GroupContext {
	var p = new(GroupContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_group

	return p
}

func (s *GroupContext) GetParser() antlr.Parser { return s.parser }

func (s *GroupContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(rdtParserLPAREN, 0)
}

func (s *GroupContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *GroupContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(rdtParserRPAREN, 0)
}

func (s *GroupContext) AllWS() []antlr.TerminalNode {
	return s.GetTokens(rdtParserWS)
}

func (s *GroupContext) WS(i int) antlr.TerminalNode {
	return s.GetToken(rdtParserWS, i)
}

func (s *GroupContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *GroupContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *GroupContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitGroup(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Group() (localctx IGroupContext) {
	localctx = NewGroupContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, rdtParserRULE_group)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(73)
		p.Match(rdtParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(77)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == rdtParserWS {
		{
			p.SetState(74)
			p.Match(rdtParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(80)
		p.Expression()
	}
	p.SetState(84)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == rdtParserWS {
		{
			p.SetState(81)
			p.Match(rdtParserWS)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

		p.SetState(86)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(87)
		p.Match(rdtParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReferenceContext is an interface to support dynamic dispatch.
type IReferenceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllIDENTIFIER() []antlr.TerminalNode
	IDENTIFIER(i int) antlr.TerminalNode
	DOT() antlr.TerminalNode

	// IsReferenceContext differentiates from other interfaces.
	IsReferenceContext()
}

type ReferenceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReferenceContext() *ReferenceContext {
	var p = new(ReferenceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_reference
	return p
}

func InitEmptyReferenceContext(p *ReferenceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = rdtParserRULE_reference
}

func (*ReferenceContext) IsReferenceContext() {}

func NewReferenceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReferenceContext {
	var p = new(ReferenceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = rdtParserRULE_reference

	return p
}

func (s *ReferenceContext) GetParser() antlr.Parser { return s.parser }

func (s *ReferenceContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(rdtParserIDENTIFIER)
}

func (s *ReferenceContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(rdtParserIDENTIFIER, i)
}

func (s *ReferenceContext) DOT() antlr.TerminalNode {
	return s.GetToken(rdtParserDOT, 0)
}

func (s *ReferenceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReferenceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReferenceContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case rdtParserVisitor:
		return t.VisitReference(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *rdtParser) Reference() (localctx IReferenceContext) {
	localctx = NewReferenceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, rdtParserRULE_reference)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(89)
		p.Match(rdtParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(92)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == rdtParserDOT {
		{
			p.SetState(90)
			p.Match(rdtParserDOT)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(91)
			p.Match(rdtParserIDENTIFIER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
