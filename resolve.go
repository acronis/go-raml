package raml

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/acronis/go-raml/rdt"
	"github.com/acronis/go-raml/stacktrace"
)

/*
Resolve resolves all unresolved (UnknownShape) shapes in the RAML.

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
		s := v.Value.(*Shape)
		if err := r.resolveShape(s); err != nil {
			se := stacktrace.NewWrapped("resolve shape", err, (*s).Base().Location, stacktrace.WithPosition(&(*s).Base().Position),
				stacktrace.WithType(stacktrace.TypeResolving))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
		r.unresolvedShapes.Remove(v)
	}
	if st != nil {
		return st
	}

	return nil
}

func (r *RAML) resolveDomainExtensions() error {
	var st *stacktrace.StackTrace
	for _, de := range r.domainExtensions {
		if err := r.resolveDomainExtension(de); err != nil {
			se := stacktrace.NewWrapped("resolve domain extension", err, de.Location, stacktrace.WithPosition(&de.Position),
				stacktrace.WithType(stacktrace.TypeResolving))
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
	// TODO: Maybe fuse resolution and value validation stage?
	parts := strings.Split(de.Name, ".")

	var ref *Shape
	// TODO: Probably can be done prettier. Needs refactor.
	switch frag := r.GetFragment(de.Location).(type) {
	case *Library:
		if len(parts) == 1 {
			r, ok := frag.AnnotationTypes.Get(parts[0])
			if !ok {
				return fmt.Errorf("reference \"%s\" not found", parts[0])
			}
			ref = r
		} else if len(parts) == 2 {
			lib, ok := frag.Uses.Get(parts[0])
			if !ok {
				return fmt.Errorf("library \"%s\" not found", parts[0])
			}
			ref, ok = lib.Link.AnnotationTypes.Get(parts[1])
			if !ok {
				return fmt.Errorf("reference \"%s\" not found", parts[1])
			}
		} else {
			return fmt.Errorf("invalid reference %s", de.Name)
		}
	case *DataType:
		// DataType cannot have local reference to annotation type.
		if len(parts) == 2 {
			lib, ok := frag.Uses.Get(parts[0])
			if !ok {
				return fmt.Errorf("library \"%s\" not found", parts[0])
			}
			ref, ok = lib.Link.AnnotationTypes.Get(parts[1])
			if !ok {
				return fmt.Errorf("reference \"%s\" not found", parts[1])
			}
		} else {
			return fmt.Errorf("invalid reference %s", de.Name)
		}
	}
	de.DefinedBy = ref

	return nil
}

func (r *RAML) resolveMultipleInheritance(target Shape) (*Shape, error) {
	inherits := target.Base().Inherits
	for _, inherit := range inherits {
		if err := r.resolveShape(inherit); err != nil {
			return nil, fmt.Errorf("resolve inherit: %w", err)
		}
	}
	// Multiple inheritance validation to be performed in a separate validation stage
	s, err := r.MakeConcreteShape(target.Base(), (*inherits[0]).Base().Type, target.(*UnknownShape).facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return &s, nil
}

func (r *RAML) resolveLink(target Shape) (*Shape, error) {
	link := target.Base().Link
	if err := r.resolveShape(link.Shape); err != nil {
		return nil, fmt.Errorf("resolve link shape: %w", err)
	}
	s, err := r.MakeConcreteShape(target.Base(), (*link.Shape).Base().Type, target.(*UnknownShape).facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return &s, nil
}

// resolveShape resolves an unknown shape in-place.
// NOTE: This function is not thread-safe. Use Clone() to create a copy of the shape before resolving if necessary.
func (r *RAML) resolveShape(shape *Shape) error {
	target := *shape
	// Skip already resolved shapes
	if _, ok := target.(*UnknownShape); !ok {
		return nil
	}

	link := target.Base().Link
	if link != nil {
		s, err := r.resolveLink(target)
		if err != nil {
			return stacktrace.NewWrapped("resolve link", err, target.Base().Location, stacktrace.WithPosition(&target.Base().Position))
		}
		*shape = *s
		return nil
	}

	shapeType := target.Base().Type
	if shapeType == TypeComposite {
		// Special case for multiple inheritance
		s, err := r.resolveMultipleInheritance(target)
		if err != nil {
			return stacktrace.NewWrapped("resolve multiple inheritance", err, target.Base().Location, stacktrace.WithPosition(&target.Base().Position))
		}
		*shape = *s
		return nil
	}

	is := antlr.NewInputStream(shapeType)
	lexer := rdt.NewrdtLexer(is)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	rdtParser := rdt.NewrdtParser(tokens)
	visitor := NewRdtVisitor(r)
	tree := rdtParser.Entrypoint()

	s, err := visitor.Visit(tree, target.(*UnknownShape))
	if err != nil {
		return stacktrace.NewWrapped("visit type expression", err, target.Base().Location, stacktrace.WithPosition(&target.Base().Position))
	}
	*shape = *s
	return nil
}
