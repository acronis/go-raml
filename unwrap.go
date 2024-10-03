package raml

import (
	"fmt"

	"github.com/acronis/go-stacktrace"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func (r *RAML) unwrapTypes(
	types *orderedmap.OrderedMap[string, *Shape],
	f *Library,
	isAnnotationType bool,
) *stacktrace.StackTrace {
	var st *stacktrace.StackTrace
	for pair := types.Oldest(); pair != nil; pair = pair.Next() {
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
		types.Set(k, ptr)
		if isAnnotationType {
			r.PutAnnotationTypeIntoFragment(us.Base().Name, f.Location, ptr)
		} else {
			r.PutTypeIntoFragment(us.Base().Name, f.Location, ptr)
		}
		r.PutShapePtr(ptr)
	}
	return st
}

func (r *RAML) unwrapLibrary(f *Library) *stacktrace.StackTrace {
	st := r.unwrapTypes(f.AnnotationTypes, f, true)
	se := r.unwrapTypes(f.Types, f, false)
	if se != nil {
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
	}
	return st
}

func (r *RAML) unwrapDataType(f *DataType) *stacktrace.StackTrace {
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
	return nil
}

func (r *RAML) unwrapFragments() *stacktrace.StackTrace {
	var st *stacktrace.StackTrace
	for _, frag := range r.fragmentsCache {
		switch f := frag.(type) {
		case *Library:
			se := r.unwrapLibrary(f)
			if se != nil {
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
			}
		case *DataType:
			se := r.unwrapDataType(f)
			if se != nil {
				if st == nil {
					st = se
				} else {
					st = st.Append(se)
				}
			}
		}
	}
	return st
}

func (r *RAML) unwrapDomainExtensions() *stacktrace.StackTrace {
	var st *stacktrace.StackTrace
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
	return st
}

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
	st := r.unwrapFragments()
	// Links to definedBy must be updated after unwrapping.
	se := r.unwrapDomainExtensions()
	if se != nil {
		if st == nil {
			st = se
		} else {
			st = st.Append(se)
		}
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

func (r *RAML) inheritUnionSource(sourceUnion *UnionShape, target Shape) (Shape, error) {
	var filtered []*Shape
	var st *stacktrace.StackTrace
	for _, source := range sourceUnion.AnyOf {
		ss := *source
		// If at least one union member has any type, the whole union is considered as any type.
		if _, ok := ss.(*AnyShape); ok {
			return target, nil
		}
		if ss.Base().Type == target.Base().Type {
			// Deep copy with ID change is required since we create new union members from source members
			cs := target.Clone()
			// TODO: Probably all copied shapes must change IDs since these are actually new shapes.
			cs.Base().ID = generateShapeID()
			is, err := cs.Inherit(ss)
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
			filtered = append(filtered, &is)
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
}

func (r *RAML) inheritUnionTarget(source Shape, targetUnion *UnionShape) (Shape, error) {
	var st *stacktrace.StackTrace
	for _, item := range targetUnion.AnyOf {
		// Merge will raise an error in case any of union members has incompatible type
		_, err := (*item).Inherit(source)
		if err != nil {
			se := StacktraceNewWrapped("merge shapes", err, targetUnion.Base().Location,
				stacktrace.WithPosition(&targetUnion.Base().Position))
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

	switch {
	case isSourceUnion && !isTargetUnion:
		return r.inheritUnionSource(sourceUnion, target)

	case isTargetUnion && !isSourceUnion:
		return r.inheritUnionTarget(source, targetUnion)
	}
	// Homogenous types produce same type
	is, err := target.Inherit(source)
	if err != nil {
		return nil, StacktraceNewWrapped("merge shapes", err, target.Base().Location,
			stacktrace.WithPosition(&target.Base().Position))
	}
	return is, nil
}

func (r *RAML) unwrapObjShape(base *BaseShape, objShape *ObjectShape, history []Shape) error {
	if objShape.Properties != nil {
		for pair := objShape.Properties.Oldest(); pair != nil; pair = pair.Next() {
			prop := pair.Value
			us, err := r.UnwrapShape(prop.Shape, history)
			if err != nil {
				return StacktraceNewWrapped("object property unwrap", err, base.Location,
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
				return StacktraceNewWrapped("object pattern property unwrap", err, base.Location,
					stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
			}
			*prop.Shape = us
		}
	}

	return nil
}

func (r *RAML) unwrapArrayShape(base *BaseShape, trg *ArrayShape, history []Shape) error {
	if trg.Items != nil {
		us, err := r.UnwrapShape(trg.Items, history)
		if err != nil {
			return StacktraceNewWrapped("array item unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*trg.Items = us
	}
	return nil
}

func (r *RAML) unwrapUnionShape(base *BaseShape, unionShape *UnionShape, history []Shape) error {
	for _, item := range unionShape.AnyOf {
		us, err := r.UnwrapShape(item, history)
		if err != nil {
			return StacktraceNewWrapped("union unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*item = us
	}
	return nil
}

func (r *RAML) unwrapParents(base *BaseShape, history []Shape) (Shape, error) {
	var source Shape
	switch {
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
	return source, nil
}

func (r *RAML) unwrapTarget(target Shape, history []Shape) error {
	switch trg := target.(type) {
	case *ArrayShape:
		if err := r.unwrapArrayShape(target.Base(), trg, history); err != nil {
			return fmt.Errorf("unwrap array shape: %w", err)
		}
	case *ObjectShape:
		if err := r.unwrapObjShape(target.Base(), trg, history); err != nil {
			return fmt.Errorf("unwrap object shape: %w", err)
		}
	case *UnionShape:
		if err := r.unwrapUnionShape(target.Base(), trg, history); err != nil {
			return fmt.Errorf("unwrap union shape: %w", err)
		}
	}
	return nil
}

// UnwrapShape recursively copies and unwraps a shape.
// Note that this method removes information about links.
func (r *RAML) UnwrapShape(s *Shape, history []Shape) (Shape, error) {
	if s == nil {
		return nil, fmt.Errorf("shape is nil")
	}
	// Perform deep copy to avoid modifying the original shape
	target := (*s).Clone()

	base := target.Base()
	// TODO: A more efficient way may be used.
	// FIXME: Detection with base.shapeVisited is not reliable and probably is not reset in some cases.
	for _, item := range history {
		if item.Base().ID == base.ID {
			return target, nil
		}
	}
	history = append(history, target)

	// Skip already unwrapped shapes
	if base.IsUnwrapped() {
		return target, nil
	}

	// NOTE: Type aliasing is not inheritance and is not used as a source. It must be unwrapped and returned as is.
	if base.Alias != nil {
		us, err := r.UnwrapShape(base.Alias, history)
		if err != nil {
			return nil, StacktraceNewWrapped("alias unwrap", err, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		us.Base().Name = base.Name
		return us, nil
	}

	source, err := r.unwrapParents(base, history)
	if err != nil {
		return nil, StacktraceNewWrapped("unwrap parents", err, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
	}

	if errUnwrap := r.unwrapTarget(target, history); errUnwrap != nil {
		return nil, StacktraceNewWrapped("unwrap target", errUnwrap, base.Location,
			stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
	}

	for pair := base.CustomShapeFacetDefinitions.Oldest(); pair != nil; pair = pair.Next() {
		prop := pair.Value
		us, errUnwrap := r.UnwrapShape(prop.Shape, history)
		if errUnwrap != nil {
			return nil, StacktraceNewWrapped("custom shape facet definition unwrap", errUnwrap, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		*prop.Shape = us
	}

	if source != nil {
		is, errInherit := r.Inherit(source, target)
		if errInherit != nil {
			return nil, StacktraceNewWrapped("merge shapes", errInherit, base.Location,
				stacktrace.WithPosition(&base.Position), stacktrace.WithType(stacktrace.TypeUnwrapping))
		}
		is.Base().unwrapped = true
		return is, nil
	}
	base.unwrapped = true
	return target, nil
}
