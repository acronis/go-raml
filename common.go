package raml

import (
	"github.com/acronis/go-stacktrace"
)

type Value[T any] struct {
	Value T
	stacktrace.Position
}

var ErrNil error
