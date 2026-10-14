package raml

import (
	"errors"
	"strings"

	"github.com/acronis/go-stacktrace"
)

// pathError ties an error to the JSON path of the data it is about, such as
// "$[0].always_on". The path is the ctxPath that validate passes down. The
// message is the message of the wrapped error, unchanged.
type pathError struct {
	path string
	err  error
}

func (e *pathError) Error() string { return e.err.Error() }

func (e *pathError) Unwrap() error { return e.err }

func withPath(path string, err error) error {
	return &pathError{path: path, err: err}
}

// deepestPath returns the path of the innermost pathError in the chain of err.
func deepestPath(err error) (string, bool) {
	var path string
	found := false
	for ; err != nil; err = errors.Unwrap(err) {
		if pe, ok := err.(*pathError); ok { //nolint:errorlint // the chain is walked here
			path, found = pe.path, true
		}
	}
	return path, found
}

// nodePositions maps the JSON path of a value in a Node ("$", "$[0]",
// "$.key") to its place in the source: the key of a mapping entry, the item
// of a sequence.
type nodePositions map[string]stacktrace.Position

// positionAt returns the position of the value at path. A path that has no
// entry, for example inside an included file, gets the position of its
// nearest ancestor that has one.
func (n *Node) positionAt(path string) (stacktrace.Position, bool) {
	for {
		if p, ok := n.positions[path]; ok {
			return p, true
		}
		i := strings.LastIndexAny(path, ".[")
		if i < 0 {
			return stacktrace.Position{}, false
		}
		path = path[:i]
	}
}

// errorPosition returns where in the source of n the validation error err
// is about: the property named by the innermost path of err, or the start of
// n when err names no path or the path is unknown.
func (n *Node) errorPosition(err error) stacktrace.Position {
	if path, ok := deepestPath(err); ok {
		if p, found := n.positionAt(path); found {
			return p
		}
	}
	return n.Position
}
