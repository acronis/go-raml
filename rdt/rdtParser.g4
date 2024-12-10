parser grammar rdtParser;

options {
	tokenVocab = rdtLexer;
}

entrypoint: expression EOF;

expression: type | union;

type: primitive | group | reference | array | optional;

primitive:
	STRING_TYPE
	| INTEGER_TYPE
	| NUMBER_TYPE
	| BOOLEAN_TYPE
	| DATETIME_TYPE
	| TIME_ONLY_TYPE
	| DATETIME_ONLY_TYPE
	| DATE_ONLY_TYPE
	| FILE_TYPE
	| NIL_TYPE
	| ANY_TYPE
	| ARRAY_TYPE
	| OBJECT_TYPE
	| UNION_TYPE;

optional: (primitive | group | reference) OPTIONAL_NOTATION;

array: (primitive | group | reference) ARRAY_NOTATION;

union: type WS* (PIPE WS* type WS*)+;

group: LPAREN WS* expression WS* RPAREN;

reference: IDENTIFIER (DOT IDENTIFIER)?;
