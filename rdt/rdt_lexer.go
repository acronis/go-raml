// Code generated from ./rdtLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package rdt

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type rdtLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var RdtLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func rdtlexerLexerInit() {
	staticData := &RdtLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
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
		"LPAREN", "RPAREN", "PIPE", "ARRAY_NOTATION", "OPTIONAL_NOTATION", "DOT",
		"STRING_TYPE", "INTEGER_TYPE", "NUMBER_TYPE", "BOOLEAN_TYPE", "DATETIME_TYPE",
		"TIME_ONLY_TYPE", "DATETIME_ONLY_TYPE", "DATE_ONLY_TYPE", "FILE_TYPE",
		"NIL_TYPE", "ANY_TYPE", "ARRAY_TYPE", "OBJECT_TYPE", "UNION_TYPE", "IDENTIFIER",
		"WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 22, 172, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7,
		1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 8, 1, 8,
		1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 9, 1, 10, 1, 10,
		1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 1,
		11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12, 1, 12, 1, 12, 1, 12,
		1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1, 12, 1,
		13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 13, 1, 14,
		1, 14, 1, 14, 1, 14, 1, 14, 1, 15, 1, 15, 1, 15, 1, 15, 1, 16, 1, 16, 1,
		16, 1, 16, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 17, 1, 18, 1, 18, 1, 18,
		1, 18, 1, 18, 1, 18, 1, 18, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1, 19, 1,
		20, 4, 20, 165, 8, 20, 11, 20, 12, 20, 166, 1, 21, 1, 21, 1, 21, 1, 21,
		0, 0, 22, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9, 19,
		10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 15, 31, 16, 33, 17, 35, 18, 37,
		19, 39, 20, 41, 21, 43, 22, 1, 0, 2, 5, 0, 45, 45, 48, 57, 65, 90, 95,
		95, 97, 122, 2, 0, 9, 9, 32, 32, 172, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0,
		0, 0, 5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0,
		0, 0, 13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0,
		0, 0, 0, 21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1,
		0, 0, 0, 0, 29, 1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35,
		1, 0, 0, 0, 0, 37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0,
		43, 1, 0, 0, 0, 1, 45, 1, 0, 0, 0, 3, 47, 1, 0, 0, 0, 5, 49, 1, 0, 0, 0,
		7, 51, 1, 0, 0, 0, 9, 54, 1, 0, 0, 0, 11, 56, 1, 0, 0, 0, 13, 58, 1, 0,
		0, 0, 15, 65, 1, 0, 0, 0, 17, 73, 1, 0, 0, 0, 19, 80, 1, 0, 0, 0, 21, 88,
		1, 0, 0, 0, 23, 97, 1, 0, 0, 0, 25, 107, 1, 0, 0, 0, 27, 121, 1, 0, 0,
		0, 29, 131, 1, 0, 0, 0, 31, 136, 1, 0, 0, 0, 33, 140, 1, 0, 0, 0, 35, 144,
		1, 0, 0, 0, 37, 150, 1, 0, 0, 0, 39, 157, 1, 0, 0, 0, 41, 164, 1, 0, 0,
		0, 43, 168, 1, 0, 0, 0, 45, 46, 5, 40, 0, 0, 46, 2, 1, 0, 0, 0, 47, 48,
		5, 41, 0, 0, 48, 4, 1, 0, 0, 0, 49, 50, 5, 124, 0, 0, 50, 6, 1, 0, 0, 0,
		51, 52, 5, 91, 0, 0, 52, 53, 5, 93, 0, 0, 53, 8, 1, 0, 0, 0, 54, 55, 5,
		63, 0, 0, 55, 10, 1, 0, 0, 0, 56, 57, 5, 46, 0, 0, 57, 12, 1, 0, 0, 0,
		58, 59, 5, 115, 0, 0, 59, 60, 5, 116, 0, 0, 60, 61, 5, 114, 0, 0, 61, 62,
		5, 105, 0, 0, 62, 63, 5, 110, 0, 0, 63, 64, 5, 103, 0, 0, 64, 14, 1, 0,
		0, 0, 65, 66, 5, 105, 0, 0, 66, 67, 5, 110, 0, 0, 67, 68, 5, 116, 0, 0,
		68, 69, 5, 101, 0, 0, 69, 70, 5, 103, 0, 0, 70, 71, 5, 101, 0, 0, 71, 72,
		5, 114, 0, 0, 72, 16, 1, 0, 0, 0, 73, 74, 5, 110, 0, 0, 74, 75, 5, 117,
		0, 0, 75, 76, 5, 109, 0, 0, 76, 77, 5, 98, 0, 0, 77, 78, 5, 101, 0, 0,
		78, 79, 5, 114, 0, 0, 79, 18, 1, 0, 0, 0, 80, 81, 5, 98, 0, 0, 81, 82,
		5, 111, 0, 0, 82, 83, 5, 111, 0, 0, 83, 84, 5, 108, 0, 0, 84, 85, 5, 101,
		0, 0, 85, 86, 5, 97, 0, 0, 86, 87, 5, 110, 0, 0, 87, 20, 1, 0, 0, 0, 88,
		89, 5, 100, 0, 0, 89, 90, 5, 97, 0, 0, 90, 91, 5, 116, 0, 0, 91, 92, 5,
		101, 0, 0, 92, 93, 5, 116, 0, 0, 93, 94, 5, 105, 0, 0, 94, 95, 5, 109,
		0, 0, 95, 96, 5, 101, 0, 0, 96, 22, 1, 0, 0, 0, 97, 98, 5, 116, 0, 0, 98,
		99, 5, 105, 0, 0, 99, 100, 5, 109, 0, 0, 100, 101, 5, 101, 0, 0, 101, 102,
		5, 45, 0, 0, 102, 103, 5, 111, 0, 0, 103, 104, 5, 110, 0, 0, 104, 105,
		5, 108, 0, 0, 105, 106, 5, 121, 0, 0, 106, 24, 1, 0, 0, 0, 107, 108, 5,
		100, 0, 0, 108, 109, 5, 97, 0, 0, 109, 110, 5, 116, 0, 0, 110, 111, 5,
		101, 0, 0, 111, 112, 5, 116, 0, 0, 112, 113, 5, 105, 0, 0, 113, 114, 5,
		109, 0, 0, 114, 115, 5, 101, 0, 0, 115, 116, 5, 45, 0, 0, 116, 117, 5,
		111, 0, 0, 117, 118, 5, 110, 0, 0, 118, 119, 5, 108, 0, 0, 119, 120, 5,
		121, 0, 0, 120, 26, 1, 0, 0, 0, 121, 122, 5, 100, 0, 0, 122, 123, 5, 97,
		0, 0, 123, 124, 5, 116, 0, 0, 124, 125, 5, 101, 0, 0, 125, 126, 5, 45,
		0, 0, 126, 127, 5, 111, 0, 0, 127, 128, 5, 110, 0, 0, 128, 129, 5, 108,
		0, 0, 129, 130, 5, 121, 0, 0, 130, 28, 1, 0, 0, 0, 131, 132, 5, 102, 0,
		0, 132, 133, 5, 105, 0, 0, 133, 134, 5, 108, 0, 0, 134, 135, 5, 101, 0,
		0, 135, 30, 1, 0, 0, 0, 136, 137, 5, 110, 0, 0, 137, 138, 5, 105, 0, 0,
		138, 139, 5, 108, 0, 0, 139, 32, 1, 0, 0, 0, 140, 141, 5, 97, 0, 0, 141,
		142, 5, 110, 0, 0, 142, 143, 5, 121, 0, 0, 143, 34, 1, 0, 0, 0, 144, 145,
		5, 97, 0, 0, 145, 146, 5, 114, 0, 0, 146, 147, 5, 114, 0, 0, 147, 148,
		5, 97, 0, 0, 148, 149, 5, 121, 0, 0, 149, 36, 1, 0, 0, 0, 150, 151, 5,
		111, 0, 0, 151, 152, 5, 98, 0, 0, 152, 153, 5, 106, 0, 0, 153, 154, 5,
		101, 0, 0, 154, 155, 5, 99, 0, 0, 155, 156, 5, 116, 0, 0, 156, 38, 1, 0,
		0, 0, 157, 158, 5, 117, 0, 0, 158, 159, 5, 110, 0, 0, 159, 160, 5, 105,
		0, 0, 160, 161, 5, 111, 0, 0, 161, 162, 5, 110, 0, 0, 162, 40, 1, 0, 0,
		0, 163, 165, 7, 0, 0, 0, 164, 163, 1, 0, 0, 0, 165, 166, 1, 0, 0, 0, 166,
		164, 1, 0, 0, 0, 166, 167, 1, 0, 0, 0, 167, 42, 1, 0, 0, 0, 168, 169, 7,
		1, 0, 0, 169, 170, 1, 0, 0, 0, 170, 171, 6, 21, 0, 0, 171, 44, 1, 0, 0,
		0, 3, 0, 164, 166, 1, 0, 1, 0,
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

// rdtLexerInit initializes any static state used to implement rdtLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewrdtLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func RdtLexerInit() {
	staticData := &RdtLexerLexerStaticData
	staticData.once.Do(rdtlexerLexerInit)
}

// NewrdtLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewrdtLexer(input antlr.CharStream) *rdtLexer {
	RdtLexerInit()
	l := new(rdtLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &RdtLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "rdtLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// rdtLexer tokens.
const (
	rdtLexerLPAREN             = 1
	rdtLexerRPAREN             = 2
	rdtLexerPIPE               = 3
	rdtLexerARRAY_NOTATION     = 4
	rdtLexerOPTIONAL_NOTATION  = 5
	rdtLexerDOT                = 6
	rdtLexerSTRING_TYPE        = 7
	rdtLexerINTEGER_TYPE       = 8
	rdtLexerNUMBER_TYPE        = 9
	rdtLexerBOOLEAN_TYPE       = 10
	rdtLexerDATETIME_TYPE      = 11
	rdtLexerTIME_ONLY_TYPE     = 12
	rdtLexerDATETIME_ONLY_TYPE = 13
	rdtLexerDATE_ONLY_TYPE     = 14
	rdtLexerFILE_TYPE          = 15
	rdtLexerNIL_TYPE           = 16
	rdtLexerANY_TYPE           = 17
	rdtLexerARRAY_TYPE         = 18
	rdtLexerOBJECT_TYPE        = 19
	rdtLexerUNION_TYPE         = 20
	rdtLexerIDENTIFIER         = 21
	rdtLexerWS                 = 22
)
