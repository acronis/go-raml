package raml

import (
	"fmt"

	"github.com/acronis/go-raml/stacktrace"
)

/*
UnwrapShapes unwraps all shapes in the RAML.

NOTE: With unwrap, we replace pointers to definitions instead of values by pointers to keep original parents unchanged.
Otherwise, parents will be also modified and unwrap may produce unpredictable results.

Unfortunately, this is required to properly support recursive shapes.
A more sophisticated approach is required to save memory and avoid copies.
*/
func (r *RAML) UnwrapShapes() error {
	// We need to invalidate old cache and re-populate it because references will no longer be valid after unwrapping.
	r.fragmentTypes = make(map[string]map[string]*Shape)
	r.fragmentAnnotationTypes = make(map[string]map[string]*Shape)
	r.shapes = nil
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for pair := f.AnnotationTypes.Oldest(); pair != nil; pair = pair.Next() {
				k, shape := pair.Key, pair.Value
				if shape == nil {
					return stacktrace.New("shape is nil", f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				position := (*shape).Base().Position
				us, err := r.UnwrapShape(shape, make([]Shape, 0))
				if err != nil {
					return stacktrace.NewWrapped("unwrap shape", err, f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
				}
				ptr := &us
				f.AnnotationTypes.Set(k, ptr)
				r.PutAnnotationTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
			for pair := f.Types.Oldest(); pair != nil; pair = pair.Next() {
				k, shape := pair.Key, pair.Value
				if shape == nil {
					return stacktrace.New("shape is nil", f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				position := (*shape).Base().Position
				us, err := r.UnwrapShape(shape, make([]Shape, 0))
				if err != nil {
					return stacktrace.NewWrapped("unwrap shape", err, f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
				}
				ptr := &us
				f.Types.Set(k, ptr)
				r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
		case *DataType:
			if f.Shape == nil {
				return stacktrace.New("shape is nil", f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
			position := (*f.Shape).Base().Position
			us, err := r.UnwrapShape(f.Shape, make([]Shape, 0))
			if err != nil {
				return stacktrace.NewWrapped("unwrap shape", err, f.Location, stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
			}
			ptr := &us
			f.Shape = ptr
			r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
			r.PutShapePtr(ptr)
		}
	}
	// Links to definedBy must be updated after unwrapping.
	for _, item := range r.domainExtensions {
		db := *item.DefinedBy
		ptr, err := r.GetAnnotationTypeFromFragmentPtr(db.Base().Location, db.Base().Name)
		if err != nil {
			return stacktrace.NewWrapped("get annotation from fragment", err, db.Base().Location, stacktrace.WithPosition(&db.Base().Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		item.DefinedBy = ptr
	}
	return nil
}

// InheritBase writes inheritable properties of sourceBase into targetBase.
// Modifies targetBase in-place.
func (r *RAML) inheritBase(sourceBase *BaseShape, targetBase *BaseShape) {
	// Back-propagate shape ID from child to parent
	// to keep the original ID when inheriting properties.
	sourceBase.Id = targetBase.Id
	// TODO: Maybe needs a bool switch for flexibility?
	for pair := sourceBase.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		if _, ok := targetBase.CustomShapeFacets.Get(k); !ok {
			targetBase.CustomShapeFacets.Set(k, item)
		}
	}
	for pair := sourceBase.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		if _, ok := targetBase.CustomDomainProperties.Get(k); !ok {
			targetBase.CustomDomainProperties.Set(k, item)
		}
	}
	// TODO: CustomShapeFacetDefinitions are not inheritable in context of unwrapper. But maybe they can be inheritable in other context?
}

// Inherit merges source shape into target shape.
func (r *RAML) Inherit(source Shape, target Shape) (Shape, error) {
	r.inheritBase(source.Base(), target.Base())
	// If source is recursive, return source as is
	if _, ok := source.(*RecursiveShape); ok {
		return source, nil
	}
	// If source type is any, return target as is
	if _, ok := source.(*AnyShape); ok {
		return target, nil
	}

	sourceUnion, isSourceUnion := source.(*UnionShape)
	targetUnion, isTargetUnion := target.(*UnionShape)

	if isSourceUnion && !isTargetUnion {
		var filtered []*Shape
		for _, item := range sourceUnion.AnyOf {
			i := *item
			// If at least one union member has any type, the whole union is considered as any type.
			if _, ok := i.(*AnyShape); ok {
				return target, nil
			}
			// TODO: Check type compatibility
			if i.Base().Type == target.Base().Type {
				// Deep copy with ID change is required since we create new union members from source members
				cs := target.Clone()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.Base().Id = generateShapeId()
				ms, err := cs.Inherit(i)
				if err != nil {
					return nil, stacktrace.NewWrapped("merge shapes", err, target.Base().Location,
						stacktrace.WithPosition(&target.Base().Position))
				}
				filtered = append(filtered, &ms)
			}
		}
		if len(filtered) == 0 {
			return nil, stacktrace.New("failed to find compatible union member", target.Base().Location,
				stacktrace.WithPosition(&target.Base().Position))
		}
		// If only one union member remains - simplify to target type
		if len(filtered) == 1 {
			return *filtered[0], nil
		}
		// Convert target to union
		target.Base().Type = TypeUnion
		return &UnionShape{
			BaseShape: *target.Base(),
			UnionFacets: UnionFacets{
				AnyOf: filtered,
			},
		}, nil
	} else if isTargetUnion && !isSourceUnion {
		for _, item := range targetUnion.AnyOf {
			// Merge will raise an error in case any of union members has incompatible type
			_, err := (*item).Inherit(source)
			if err != nil {
				return nil, stacktrace.NewWrapped("merge shapes", err, target.Base().Location,
					stacktrace.WithPosition(&target.Base().Position))
			}
		}
		return targetUnion, nil
	} else {
		// Homogenous types produce same type
		ms, err := target.Inherit(source)
		if err != nil {
			return nil, stacktrace.NewWrapped("merge shapes", err, target.Base().Location,
				stacktrace.WithPosition(&target.Base().Position))
		}
		return ms, nil
	}
}

// Recursively copies and unwraps a shape.
// Note that this method removes information about links.
func (r *RAML) UnwrapShape(s *Shape, history []Shape) (Shape, error) {
	if s == nil {
		return nil, fmt.Errorf("shape is nil")
	}
	// Perform deep copy to avoid modifying the original shape
	target := (*s).Clone()

	base := target.Base()
	// Skip already unwrapped shapes
	if base.IsUnwrapped() {
		return target, nil
	}

	for _, item := range history {
		if item.Base().Id == base.Id {
			base.Inherits = nil
			return &RecursiveShape{BaseShape: *base, Head: &item}, nil
		}
	}
	history = append(history, target)

	var source Shape
	if base.Alias != nil {
		us, err := r.UnwrapShape(base.Alias, history)
		if err != nil {
			return nil, stacktrace.NewWrapped("alias unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		// Alias simply points to another shape, so we just change the name and return it as is.
		us.Base().Name = base.Name
		return us, nil
	} else if base.Link != nil {
		// r.inheritBase((*link.Shape).Base(), base)
		us, err := r.UnwrapShape(base.Link.Shape, history)
		if err != nil {
			return nil, stacktrace.NewWrapped("link unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		source = us

		base.Link = nil
	} else if len(base.Inherits) > 0 {
		inherits := base.Inherits
		unwrappedInherits := make([]*Shape, len(inherits))
		// TODO: Fix multiple inheritance unwrap.
		// Multiple inheritance members must be checked for compatibility with each other before unwrapping.
		ss, err := r.UnwrapShape(inherits[0], history)
		if err != nil {
			return nil, stacktrace.NewWrapped("parent unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		unwrappedInherits[0] = &ss
		for i := 1; i < len(inherits); i++ {
			us, err := r.UnwrapShape(inherits[i], history)
			if err != nil {
				return nil, err
			}
			unwrappedInherits[i] = &us
			_, err = ss.Inherit(us)
			if err != nil {
				return nil, stacktrace.NewWrapped("multiple parents unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
		}
		source = ss

		base.Inherits = unwrappedInherits
	}

	if t, ok := target.(*ArrayShape); ok && t.Items != nil {
		us, err := r.UnwrapShape(t.Items, history)
		if err != nil {
			return nil, stacktrace.NewWrapped("array item unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*t.Items = us
	} else if t, ok := target.(*ObjectShape); ok {
		if t.Properties != nil {
			for pair := t.Properties.Oldest(); pair != nil; pair = pair.Next() {
				prop := pair.Value
				us, err := r.UnwrapShape(prop.Shape, history)
				if err != nil {
					return nil, stacktrace.NewWrapped("object property unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				*prop.Shape = us
			}
		}
		if t.PatternProperties != nil {
			for pair := t.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
				prop := pair.Value
				us, err := r.UnwrapShape(prop.Shape, history)
				if err != nil {
					return nil, stacktrace.NewWrapped("object pattern property unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				*prop.Shape = us
			}
		}
	} else if t, ok := target.(*UnionShape); ok {
		for _, item := range t.AnyOf {
			us, err := r.UnwrapShape(item, history)
			if err != nil {
				return nil, stacktrace.NewWrapped("union unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
			*item = us
		}
	}

	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		us, err := r.UnwrapShape(prop.Shape, history)
		if err != nil {
			return nil, stacktrace.NewWrapped("custom shape facet definition unwrap", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*prop.Shape = us
	}

	if source != nil {
		ms, err := r.Inherit(source, target)
		if err != nil {
			return nil, stacktrace.NewWrapped("merge shapes", err, base.Location, stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		ms.Base().unwrapped = true
		return ms, nil
	}
	base.unwrapped = true
	return target, nil
}
