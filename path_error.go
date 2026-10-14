package raml

import (
	"errors"
	"strconv"
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

// errorPosition returns where in the source of n the validation error err is
// about: the property or item named by the innermost path of err, or the
// nearest ancestor of it that the source has. It returns the start of the
// value when err names no path.
func (n *DataNode) errorPosition(err error) stacktrace.Position {
	path, ok := deepestPath(err)
	if !ok || n.Value == nil {
		return n.ValuePos
	}
	return positionAtPath(n.Value, strings.TrimPrefix(path, "$"), n.ValuePos)
}

// positionAtPath follows path ("[0]", ".key", ".a.b[2]") through v and
// returns the position of the last step it could follow, or fallback.
func positionAtPath(v *ValueNode, path string, fallback stacktrace.Position) stacktrace.Position {
	pos := fallback
	for path != "" && v != nil {
		switch {
		case path[0] == '[' && v.Sequence != nil:
			end := strings.IndexByte(path, ']')
			if end < 0 {
				return pos
			}
			idx, err := strconv.Atoi(path[1:end])
			if err != nil || idx < 0 || idx >= len(v.Sequence.Items) {
				return pos
			}
			item := v.Sequence.Items[idx]
			pos, v, path = item.ValuePos, item.Value, path[end+1:]
		case path[0] == '.' && v.Mapping != nil:
			entry, rest, ok := matchEntry(v.Mapping, path[1:])
			if !ok {
				return pos
			}
			// From the key to the end of the value: the whole property.
			pos, v, path = entry.KeyPos, entry.Value, rest
			if entry.ValuePos.EndLine > 0 {
				pos.EndLine, pos.EndColumn = entry.ValuePos.EndLine, entry.ValuePos.EndColumn
			}
		default:
			return pos
		}
	}
	return pos
}

// matchEntry finds the entry whose key starts rest. Keys may contain dots and
// brackets, so the longest key that ends at a path separator wins.
func matchEntry(m *MappingValue, rest string) (MappingEntry, string, bool) {
	var best MappingEntry
	found := false
	for _, e := range m.Entries {
		if !strings.HasPrefix(rest, e.Key) {
			continue
		}
		tail := rest[len(e.Key):]
		if tail != "" && tail[0] != '.' && tail[0] != '[' {
			continue
		}
		if !found || len(e.Key) > len(best.Key) {
			best, found = e, true
		}
	}
	if !found {
		return MappingEntry{}, "", false
	}
	return best, rest[len(best.Key):], true
}
