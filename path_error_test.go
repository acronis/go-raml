package raml

import (
	"errors"
	"fmt"
	"testing"

	"github.com/acronis/go-stacktrace"
)

func pos(line, col int) stacktrace.Position { return stacktrace.Position{Line: line, Column: col} }

func TestDataNode_errorPosition(t *testing.T) {
	root := &DataNode{
		ValuePos: pos(1, 1),
		Value: &ValueNode{Mapping: &MappingValue{Entries: []MappingEntry{
			{Key: "a.b", KeyPos: pos(2, 1), Value: &ValueNode{Mapping: &MappingValue{Entries: []MappingEntry{
				{Key: "c", KeyPos: pos(3, 3)},
			}}}},
			{Key: "a", KeyPos: pos(4, 1), Value: &ValueNode{Sequence: &SequenceValue{Items: []SequenceItem{
				{ValuePos: pos(5, 3)},
				{ValuePos: pos(6, 3), Value: &ValueNode{Mapping: &MappingValue{Entries: []MappingEntry{
					{Key: "k[0]", KeyPos: pos(7, 5)},
				}}}},
			}}}},
		}}},
	}

	tests := []struct {
		name string
		err  error
		want stacktrace.Position
	}{
		{"no path keeps the start of the value", errors.New("plain"), pos(1, 1)},
		{"key with a dot", withPath("$.a.b.c", errors.New("x")), pos(3, 3)},
		{"the key with the dot itself", withPath("$.a.b", errors.New("x")), pos(2, 1)},
		{"sequence item", withPath("$.a[0]", errors.New("x")), pos(5, 3)},
		{"key with a bracket in an item", withPath("$.a[1].k[0]", errors.New("x")), pos(7, 5)},
		{"unknown key stays at the last position reached", withPath("$.a[1].nope", errors.New("x")), pos(6, 3)},
		{"index out of range", withPath("$.a[9]", errors.New("x")), pos(4, 1)},
		{"the innermost path wins", withPath("$.a", fmt.Errorf("w: %w", withPath("$.a[0]", errors.New("x")))), pos(5, 3)},
		{"the root path is the start of the value", withPath("$", errors.New("x")), pos(1, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := root.errorPosition(tt.err)

			if got.Line != tt.want.Line || got.Column != tt.want.Column {
				t.Errorf("errorPosition() = %d:%d, want %d:%d", got.Line, got.Column, tt.want.Line, tt.want.Column)
			}
		})
	}
}

func TestDataNode_errorPosition_withoutValue(t *testing.T) {
	n := &DataNode{ValuePos: pos(3, 4)}

	got := n.errorPosition(withPath("$.a", errors.New("x")))

	if got.Line != 3 || got.Column != 4 {
		t.Errorf("errorPosition() = %d:%d, want 3:4", got.Line, got.Column)
	}
}

func TestDataNode_errorPosition_rangeCoversTheWholeProperty(t *testing.T) {
	n := &DataNode{
		ValuePos: pos(1, 1),
		Value: &ValueNode{Mapping: &MappingValue{Entries: []MappingEntry{{
			Key:      "k",
			KeyPos:   stacktrace.Position{Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
			ValuePos: stacktrace.Position{Line: 2, Column: 4, EndLine: 3, EndColumn: 10},
			Value:    &ValueNode{},
		}}}},
	}

	got := n.errorPosition(withPath("$.k", errors.New("x")))

	want := stacktrace.Position{Line: 2, Column: 1, EndLine: 3, EndColumn: 10}
	if got != want {
		t.Errorf("errorPosition() = %+v, want %+v", got, want)
	}
}
