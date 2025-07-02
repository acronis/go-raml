package raml

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"

	"github.com/acronis/go-stacktrace"

	"github.com/acronis/go-raml/v2/rdt"
)

/*
resolveShapes resolves all unresolved (UnknownShape) shapes in the RAML.

NOTE: Unresolved shapes is a linked list that is populated by the YAML parser. Shape parsing occurs in two places:

1. During the RAML fragment parsing. Shapes that could not be determined during the parsing process
are stored in `unresolvedShapes` as UnknownShape. UnknownShape includes YAML nodes that can be parsed later.

2. During the shape resolution. YAML parser is invoked on YAML nodes that UnknownShape has stored.

This helps to avoid additional traversals of nested shapes since the traverse is already done by YAML parser and it will
generate UnknownShapes and add them to `unresolvedShapes` recursively as they occur.
*/
func (r *RAML) resolveShapes() error {
	var st *stacktrace.StackTrace
	for r.unresolvedShapes.Len() > 0 {
		v := r.unresolvedShapes.Front()
		base, ok := v.Value.(*BaseShape)
		if !ok {
			return fmt.Errorf("invalid unresolved shape: Value is not *BaseShape: %T", v.Value)
		}
		if err := r.resolveShape(base); err != nil {
			se := StacktraceNewWrapped("resolve shape", err, base.Location,
				stacktrace.WithPosition(&base.Position),
				stacktrace.WithType(StacktraceTypeResolving))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
		}
		r.unresolvedShapes.Remove(v)
	}
	if st != nil {
		return st
	}

	return nil
}

// resolveDomainExtensions resolves all domain extensions in the RAML.
func (r *RAML) resolveDomainExtensions() error {
	var st *stacktrace.StackTrace
	for _, de := range r.domainExtensions {
		if err := r.resolveDomainExtension(de); err != nil {
			se := StacktraceNewWrapped("resolve domain extension", err, de.Location,
				stacktrace.WithPosition(&de.Position),
				stacktrace.WithType(StacktraceTypeResolving))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
	}
	if st != nil {
		return st
	}

	return nil
}

func (r *RAML) resolveDomainExtension(de *DomainExtension) error {
	ref, err := r.GetReferencedAnnotationType(de.Name, de.Location)
	if err != nil {
		return fmt.Errorf("get referenced shape: %w", err)
	}

	de.DefinedBy = ref

	return nil
}

func (r *RAML) resolveMultipleInheritance(base *BaseShape, shape *UnknownShape) (Shape, error) {
	inherits := base.Inherits
	for _, inherit := range inherits {
		if err := r.resolveShape(inherit); err != nil {
			return nil, fmt.Errorf("resolve inherit: %w", err)
		}
	}
	// Multiple inheritance validation to be performed in a separate validation stage
	s, err := r.MakeConcreteShapeYAML(base, inherits[0].Type, shape.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return s, nil
}

func (r *RAML) resolveLink(base *BaseShape, shape *UnknownShape) (Shape, error) {
	linkShape := base.Link.Shape
	if err := r.resolveShape(linkShape); err != nil {
		return nil, fmt.Errorf("resolve link shape: %w", err)
	}
	s, err := r.MakeConcreteShapeYAML(base, linkShape.Type, shape.facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return s, nil
}

type CustomErrorListener struct {
	*antlr.DefaultErrorListener // Embed default which ensures we fit the interface
	Stacktrace                  *stacktrace.StackTrace
	position                    stacktrace.Position
	location                    string
}

func (c *CustomErrorListener) SyntaxError(
	_ antlr.Recognizer,
	offendingSymbol interface{},
	_, _ int,
	msg string,
	_ antlr.RecognitionException,
) {
	symbolInfoOpt := stacktrace.WithInfo("offendingSymbol", offendingSymbol)
	posOpt := stacktrace.WithPosition(
		&stacktrace.Position{
			Line:   c.position.Line,
			Column: c.position.Column,
		},
	)
	if c.Stacktrace == nil {
		c.Stacktrace = StacktraceNew("antlr error", c.location)
	}
	c.Stacktrace = c.Stacktrace.Append(StacktraceNew(msg, c.location, posOpt, symbolInfoOpt))
}

// resolveShape resolves an unknown shape in-place.
// NOTE: This function is not thread-safe. Use Clone() to create a copy of the shape before resolving if necessary.
func (r *RAML) resolveShape(base *BaseShape) error {
	shape := base.Shape
	if shape == nil {
		return fmt.Errorf("shape is nil")
	}

	// Skip already resolved shapes
	unknownShape, ok := shape.(*UnknownShape)
	if !ok {
		return nil
	}

	if base.Link != nil {
		s, err := r.resolveLink(base, unknownShape)
		if err != nil {
			return StacktraceNewWrapped("resolve link", err, base.Location,
				stacktrace.WithPosition(&base.Position))
		}
		base.SetShape(s)
		return nil
	}

	shapeType := base.Type
	if shapeType == TypeComposite {
		// Special case for multiple inheritance
		s, err := r.resolveMultipleInheritance(base, unknownShape)
		if err != nil {
			return StacktraceNewWrapped("resolve multiple inheritance", err, base.Location,
				stacktrace.WithPosition(&base.Position))
		}
		base.SetShape(s)
		return nil
	}

	is := antlr.NewInputStream(shapeType)
	customListener := &CustomErrorListener{
		location: base.Location,
		position: base.Position,
	}
	lexer := rdt.NewrdtLexer(is)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(customListener)

	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	rdtParser := rdt.NewrdtParser(tokens)
	rdtParser.RemoveErrorListeners()
	rdtParser.AddErrorListener(customListener)

	visitor := NewRdtVisitor(r)
	tree := rdtParser.Entrypoint()

	if customListener.Stacktrace != nil {
		return customListener.Stacktrace
	}

	s, err := visitor.Visit(tree, unknownShape)
	if err != nil {
		return StacktraceNewWrapped("visit type expression", err, base.Location,
			stacktrace.WithPosition(&base.Position))
	}
	base.SetShape(s)
	return nil
}
