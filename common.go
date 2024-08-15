package raml

type Position struct {
	Line   int
	Column int
}

type Value[T any] struct {
	Value T
	Position
}
