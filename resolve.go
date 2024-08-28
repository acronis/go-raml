package raml

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/acronis/go-raml/rdt"
)

func (r *RAML) resolveShapes() error {
	for _, shape := range r.GetAllShapesPtr() {
		if err := r.resolveShape(shape); err != nil {
			return fmt.Errorf("resolve shape: %w", err)
		}
	}

	return nil
}

func (r *RAML) resolveDomainExtensions() error {
	for _, de := range r.domainExtensions {
		if err := r.resolveDomainExtension(de); err != nil {
			return fmt.Errorf("resolve domain extension: %w", err)
		}
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
			ref = frag.AnnotationTypes[parts[0]]
			if ref == nil {
				return fmt.Errorf("reference %s not found", parts[0])
			}
		} else if len(parts) == 2 {
			lib := frag.Uses[parts[0]]
			if lib == nil {
				return fmt.Errorf("library %s not found", parts[0])
			}
			ref = lib.Link.AnnotationTypes[parts[1]]
			if ref == nil {
				return fmt.Errorf("reference %s not found", parts[1])
			}
		} else {
			return fmt.Errorf("invalid reference %s", de.Name)
		}
	case *DataType:
		// DataType cannot have local reference to annotation type.
		if len(parts) == 2 {
			lib := frag.Uses[parts[0]]
			if lib == nil {
				return fmt.Errorf("library %s not found", parts[0])
			}
			ref = lib.Link.AnnotationTypes[parts[1]]
			if ref == nil {
				return fmt.Errorf("reference %s not found", parts[1])
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
	s, err := r.makeConcreteShape(target.Base(), (*inherits[0]).Base().Type, target.(*UnknownShape).facets)
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
	s, err := r.makeConcreteShape(target.Base(), (*link.Shape).Base().Type, target.(*UnknownShape).facets)
	if err != nil {
		return nil, fmt.Errorf("make concrete shape: %w", err)
	}
	return &s, nil
}

func (o *ObjectShape) resolveProperties() error {
	// NOTE: This function is not susceptible to cyclic dependency because shape resolution returns as soon as shape is resolved.
	for _, prop := range o.Properties {
		// Traverse into object sub-properties recursively
		if s, ok := (*prop.Shape).(*ObjectShape); ok {
			if err := s.resolveProperties(); err != nil {
				return fmt.Errorf("resolve property shape: %w", err)
			}
		} else if err := o.raml.resolveShape(prop.Shape); err != nil {
			return fmt.Errorf("resolve property shape: %w", err)
		}
	}
	return nil
}

// resolveShape resolves an unknown shape in-place.
// NOTE: This function is not thread-safe. Use Clone() to create a copy of the shape before resolving if necessary.
func (r *RAML) resolveShape(shape *Shape) error {
	target := *shape
	if target.Base().resolved {
		return nil
	}
	// Skip already resolved shapes
	if _, ok := target.(*UnknownShape); !ok {
		target.Base().resolved = true
		return nil
	}

	link := target.Base().Link
	if link != nil {
		s, err := r.resolveLink(target)
		if err != nil {
			return fmt.Errorf("resolve link: %w", err)
		}
		*shape = *s
		(*shape).Base().resolved = true
		return nil
	}

	shapeType := target.Base().Type
	if shapeType == TypeComposite {
		// Special case for multiple inheritance
		s, err := r.resolveMultipleInheritance(target)
		if err != nil {
			return fmt.Errorf("resolve multiple inheritance: %w", err)
		}
		*shape = *s
		(*shape).Base().resolved = true
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
		return fmt.Errorf("visit tree: %w", err)
	}
	sv := *s
	*shape = sv
	if ss, ok := sv.(*ObjectShape); ok && ss.Properties != nil {
		// Unresolved object shape may contain properties that couldn't be taken into account during parsing.
		if err := ss.resolveProperties(); err != nil {
			return fmt.Errorf("resolve property shape: %w", err)
		}
	} else if ss, ok := sv.(*ArrayShape); ok && ss.Items != nil {
		// Unresolved array shape may contain unresolved items shape.
		if err := r.resolveShape(ss.Items); err != nil {
			return fmt.Errorf("resolve array items shape: %w", err)
		}
	}
	(*shape).Base().resolved = true
	return nil
}
