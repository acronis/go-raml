package raml

import (
	"github.com/acronis/go-raml/stacktrace"
)

type Value[T any] struct {
	Value T
	stacktrace.Position
}
