package raml

import (
	"fmt"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
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
	var st *stacktrace.StackTrace
	st = nil
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			for pair := f.AnnotationTypes.Oldest(); pair != nil; pair = pair.Next() {
				k, shape := pair.Key, pair.Value
				if shape == nil {
					se := stacktrace.New("shape is nil", f.Location,
						stacktrace.WithType(stacktrace.TypeUnwrapping))
					if st == nil {
						st = se
					} else {
						st = st.Append(se)
					}
					continue
				}
				position := (*shape).Base().Position
				us, err := r.UnwrapShape(shape, make([]Shape, 0))
				if err != nil {
					se := StacktraceNewWrapped("unwrap shape", err, f.Location,
						stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
					if st == nil {
						st = se
					} else {
						st = st.Append(se)
					}
					continue
				}
				ptr := &us
				f.AnnotationTypes.Set(k, ptr)
				r.PutAnnotationTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
			for pair := f.Types.Oldest(); pair != nil; pair = pair.Next() {
				k, shape := pair.Key, pair.Value
				if shape == nil {
					se := stacktrace.New("shape is nil", f.Location,
						stacktrace.WithType(stacktrace.TypeUnwrapping))
					if st == nil {
						st = se
					} else {
						st = st.Append(se)
					}
					continue
				}
				position := (*shape).Base().Position
				us, err := r.UnwrapShape(shape, make([]Shape, 0))
				if err != nil {
					se := StacktraceNewWrapped("unwrap shape", err, f.Location,
						stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
					if st == nil {
						st = se
					} else {
						st = st.Append(se)
					}
					continue
				}
				ptr := &us
				f.Types.Set(k, ptr)
				r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
				r.PutShapePtr(ptr)
			}
		case *DataType:
			if f.Shape == nil {
				return stacktrace.New("shape is nil", f.Location,
					stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
			position := (*f.Shape).Base().Position
			us, err := r.UnwrapShape(f.Shape, make([]Shape, 0))
			if err != nil {
				return StacktraceNewWrapped("unwrap shape", err, f.Location,
					stacktrace.WithType(stacktrace.TypeUnwrapping), stacktrace.WithPosition(&position))
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
			se := StacktraceNewWrapped("get annotation from fragment", err, db.Base().Location,
				stacktrace.WithPosition(&db.Base().Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			if st == nil {
				st = se
			} else {
				st = st.Append(se)
			}
			continue
		}
		item.DefinedBy = ptr
	}
	if st != nil {
		return st
	}
	return nil
}

// InheritBase writes inheritable properties of sourceBase into targetBase.
// Copies necessary properties and modifies targetBase in-place.
func (r *RAML) inheritBase(sourceBase *BaseShape, targetBase *BaseShape) {
	// Back-propagate shape ID from child to parent
	// to keep the original ID when inheriting properties.
	sourceBase.ID = targetBase.ID

	// TODO: Probably the copies must be implemented in Clone method.
	customShapeFacets := orderedmap.New[string, *Node](targetBase.CustomShapeFacets.Len())
	for pair := targetBase.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		customShapeFacets.Set(k, item)
	}
	for pair := sourceBase.CustomShapeFacets.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		if _, ok := targetBase.CustomShapeFacets.Get(k); !ok {
			customShapeFacets.Set(k, item)
		}
	}
	targetBase.CustomShapeFacets = customShapeFacets

	customDomainProperties := orderedmap.New[string, *DomainExtension](targetBase.CustomDomainProperties.Len())
	for pair := targetBase.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		customDomainProperties.Set(k, item)
	}
	for pair := sourceBase.CustomDomainProperties.Oldest(); pair != nil; pair = pair.Next() {
		k, item := pair.Key, pair.Value
		if _, ok := targetBase.CustomDomainProperties.Get(k); !ok {
			customDomainProperties.Set(k, item)
		}
	}
	targetBase.CustomDomainProperties = customDomainProperties
	// TODO: CustomShapeFacetDefinitions are not inheritable in context of unwrapper.
	//  But maybe they can be inheritable in other context?
}

// Inherit merges source shape into target shape.
func (r *RAML) Inherit(source Shape, target Shape) (Shape, error) {
	var st *stacktrace.StackTrace
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

	switch {
	case isSourceUnion && !isTargetUnion:
		var filtered []*Shape
		for _, item := range sourceUnion.AnyOf {
			i := *item
			// If at least one union member has any type, the whole union is considered as any type.
			if _, ok := i.(*AnyShape); ok {
				return target, nil
			}
			if i.Base().Type == target.Base().Type {
				// Deep copy with ID change is required since we create new union members from source members
				cs := target.Clone()
				// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
				cs.Base().ID = generateShapeID()
				ms, err := cs.Inherit(i)
				if err != nil {
					se := StacktraceNewWrapped("merge shapes", err, target.Base().Location,
						stacktrace.WithPosition(&target.Base().Position))
					if st == nil {
						st = se
					} else {
						st = st.Append(se)
					}
					// Skip shapes that didn't pass inheritance check
					continue
				}
				filtered = append(filtered, &ms)
			}
		}
		if len(filtered) == 0 {
			se := stacktrace.New("failed to find compatible union member", target.Base().Location,
				stacktrace.WithPosition(&target.Base().Position))
			if st != nil {
				se = se.Append(st)
			}
			return nil, se
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
	case isTargetUnion && !isSourceUnion:
		for _, item := range targetUnion.AnyOf {
			// Merge will raise an error in case any of union members has incompatible type
			_, err := (*item).Inherit(source)
			if err != nil {
				se := StacktraceNewWrapped("merge shapes", err, target.Base().Location,
					stacktrace.WithPosition(&target.Base().Position))
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
				continue
			}
		}
		if st != nil {
			return nil, st
		}
		return targetUnion, nil
	}
	// Homogenous types produce same type
	ms, err := target.Inherit(source)
	if err != nil {
		return nil, StacktraceNewWrapped("merge shapes", err, target.Base().Location,
			stacktrace.WithPosition(&target.Base().Position))
	}
	return ms, nil
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
		if item.Base().ID == base.ID {
			base.Inherits = nil
			return &RecursiveShape{BaseShape: *base, Head: &item}, nil
		}
	}
	history = append(history, target)

	var source Shape
	switch {
	case base.Alias != nil:
		us, err := r.UnwrapShape(base.Alias, history)
		if err != nil {
			return nil, StacktraceNewWrapped("alias unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		// Alias simply points to another shape, so we just change the name and return it as is.
		us.Base().Name = base.Name
		return us, nil
	case base.Link != nil:
		us, err := r.UnwrapShape(base.Link.Shape, history)
		if err != nil {
			return nil, StacktraceNewWrapped("link unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		source = us
		base.Link = nil
	case len(base.Inherits) > 0:
		inherits := base.Inherits
		unwrappedInherits := make([]*Shape, len(inherits))
		// TODO: Fix multiple inheritance unwrap.
		// Multiple inheritance members must be checked for compatibility with each other before unwrapping.
		ss, err := r.UnwrapShape(inherits[0], history)
		if err != nil {
			return nil, StacktraceNewWrapped("parent unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		unwrappedInherits[0] = &ss
		for i := 1; i < len(inherits); i++ {
			us, errUnwrap := r.UnwrapShape(inherits[i], history)
			if errUnwrap != nil {
				return nil, errUnwrap
			}
			unwrappedInherits[i] = &us
			_, errUnwrap = ss.Inherit(us)
			if errUnwrap != nil {
				return nil, StacktraceNewWrapped("multiple parents unwrap", errUnwrap, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
		}
		source = ss
		base.Inherits = unwrappedInherits
	}

	if arrayShape, isArray := target.(*ArrayShape); isArray && arrayShape.Items != nil {
		us, err := r.UnwrapShape(arrayShape.Items, history)
		if err != nil {
			return nil, StacktraceNewWrapped("array item unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*arrayShape.Items = us
	} else if objShape, ok := target.(*ObjectShape); ok {
		if objShape.Properties != nil {
			for pair := objShape.Properties.Oldest(); pair != nil; pair = pair.Next() {
				prop := pair.Value
				us, err := r.UnwrapShape(prop.Shape, history)
				if err != nil {
					return nil, StacktraceNewWrapped("object property unwrap", err, base.Location,
						stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				*prop.Shape = us
			}
		}
		if objShape.PatternProperties != nil {
			for pair := objShape.PatternProperties.Oldest(); pair != nil; pair = pair.Next() {
				prop := pair.Value
				us, err := r.UnwrapShape(prop.Shape, history)
				if err != nil {
					return nil, StacktraceNewWrapped("object pattern property unwrap", err, base.Location,
						stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
				}
				*prop.Shape = us
			}
		}
	} else if unionShape, isUnion := target.(*UnionShape); isUnion {
		for _, item := range unionShape.AnyOf {
			us, err := r.UnwrapShape(item, history)
			if err != nil {
				return nil, StacktraceNewWrapped("union unwrap", err, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
			*item = us
		}
	}

	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		us, err := r.UnwrapShape(prop.Shape, history)
		if err != nil {
			return nil, StacktraceNewWrapped("custom shape facet definition unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*prop.Shape = us
	}

	if source != nil {
		ms, err := r.Inherit(source, target)
		if err != nil {
			return nil, StacktraceNewWrapped("merge shapes", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		ms.Base().unwrapped = true
		return ms, nil
	}
	base.unwrapped = true
	return target, nil
}
