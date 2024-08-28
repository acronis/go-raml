package raml

import "gopkg.in/yaml.v3"

type Position struct {
	Line   int
	Column int
}

type Value[T any] struct {
	Value T
	Position
}

// NewNodePosition creates a new position from the given node.
func NewNodePosition(node *yaml.Node) *Position {
	return &Position{Line: node.Line, Column: node.Column}
}

// NewPosition creates a new position with the given line and column.
func NewPosition(line, column int) *Position {
	return &Position{Line: line, Column: column}
}

// ErrOpt is an option for Error.
type optErrNodePosition struct {
	pos *Position
}

// Apply sets the position of the error to the given position.
// implements ErrOpt
func (o optErrNodePosition) Apply(e *Error) {
	e.Position = o.pos
}

// WithNodePosition sets the position of the error to the position of the given node.
func WithNodePosition(node *yaml.Node) ErrOpt {
	return optErrNodePosition{pos: NewNodePosition(node)}
}
