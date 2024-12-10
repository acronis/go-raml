lexer grammar rdtLexer;

// Priority 0
LPAREN: '(';
RPAREN: ')';
PIPE: '|';
ARRAY_NOTATION: '[]';
OPTIONAL_NOTATION: '?';
DOT: '.';

// Priority 1
STRING_TYPE: 'string';
INTEGER_TYPE: 'integer';
NUMBER_TYPE: 'number';
BOOLEAN_TYPE: 'boolean';
DATETIME_TYPE: 'datetime';
TIME_ONLY_TYPE: 'time-only';
DATETIME_ONLY_TYPE: 'datetime-only';
DATE_ONLY_TYPE: 'date-only';
FILE_TYPE: 'file';
NIL_TYPE: 'nil';
ANY_TYPE: 'any';
ARRAY_TYPE: 'array';
OBJECT_TYPE: 'object';
UNION_TYPE: 'union';

// Priority 2
IDENTIFIER: ([0-9] | [a-z] | [A-Z] | '_' | '-')+;

WS: [ \t] -> channel(HIDDEN);